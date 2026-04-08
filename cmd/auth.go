package cmd

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"github.com/gomicro/align/config"

	"github.com/gomicro/trust"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var (
	reapprove    bool
	clientID     string
	clientSecret string
)

func init() {
	RootCmd.AddCommand(authCmd)

	authCmd.Flags().BoolVarP(&reapprove, "force", "f", false, "force align to reauth")
}

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with GitHub",
	Long:  `Authorize align to access GitHub on your behalf via OAuth.`,
	RunE:  authFunc,
}

func authFunc(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if clientID == "" || clientSecret == "" {
		cmd.SilenceUsage = true
		return fmt.Errorf("client id and secret must be baked into the binary, and are not present")
	}

	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("generating state nonce: %w", err)
	}
	state := hex.EncodeToString(stateBytes)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		cmd.SilenceUsage = true
		return err
	}

	port := listener.Addr().(*net.TCPAddr).Port

	pool := trust.New()

	certs, err := pool.CACerts()
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("failed to create cert pool: %w", err)
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:    certs,
				MinVersion: tls.VersionTLS12,
			},
		},
	}

	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)

	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"repo"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://github.com/login/oauth/authorize",
			TokenURL: "https://github.com/login/oauth/access_token",
		},
		RedirectURL: fmt.Sprintf("http://localhost:%v/auth", port),
	}

	token := make(chan string)

	go startServer(ctx, listener, conf, state, token)

	var opts []oauth2.AuthCodeOption
	if reapprove {
		opts = []oauth2.AuthCodeOption{oauth2.AccessTypeOffline, oauth2.ApprovalForce}
	} else {
		opts = []oauth2.AuthCodeOption{oauth2.AccessTypeOffline}
	}

	url := conf.AuthCodeURL(state, opts...)

	err = openBrowser(url)
	if err != nil {
		cmd.SilenceUsage = true
		return err
	}

	tkn := <-token
	close(token)

	c, err := config.ParseFromFile()
	if err != nil {
		cmd.SilenceUsage = true
		return err
	}

	c.Github.Token = tkn

	err = c.WriteFile()
	if err != nil {
		cmd.SilenceUsage = true
		return err
	}

	return nil
}

func startServer(ctx context.Context, listener net.Listener, conf *oauth2.Config, state string, token chan string) {
	http.HandleFunc("/auth", authHandler(ctx, conf, state, token))

	srv := &http.Server{}

	go func() {
		<-ctx.Done()
		err := srv.Shutdown(ctx)
		if err != nil {
			fmt.Printf("Error shutting down server: %v", err.Error())
			os.Exit(1)
		}
	}()

	err := srv.Serve(listener)
	if err != nil {
		fmt.Printf("Error: %v", err.Error())
		os.Exit(1)
	}
}

func authHandler(ctx context.Context, conf *oauth2.Config, state string, token chan string) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		code := req.URL.Query().Get("code")
		rstate := req.URL.Query().Get("state")

		if rstate != state {
			fmt.Println("bad response from oauth server")
			os.Exit(1)
		}

		tok, err := conf.Exchange(ctx, code)
		if err != nil {
			fmt.Printf("errored exchanging token: %v", err.Error())
			os.Exit(1)
		}

		body := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Align — Authorized</title>
  <style>
    @media (prefers-color-scheme: dark) {
      :root {
        --bg:      #0d1117;
        --card-bg: #161b22;
        --border:  #30363d;
        --text:    #e6edf3;
        --muted:   #8b949e;
        --accent:  #3fb950;
      }
    }
    @media (prefers-color-scheme: light) {
      :root {
        --bg:      #f6f8fa;
        --card-bg: #ffffff;
        --border:  #d0d7de;
        --text:    #1f2328;
        --muted:   #636c76;
        --accent:  #1a7f37;
      }
    }
    *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      background: var(--bg);
      color: var(--text);
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
      min-height: 100vh;
      display: flex;
      align-items: center;
      justify-content: center;
    }
    .card {
      background: var(--card-bg);
      border: 1px solid var(--border);
      border-radius: 12px;
      padding: 48px 56px;
      max-width: 420px;
      width: 100%;
      text-align: center;
      box-shadow: 0 4px 24px rgba(0,0,0,0.12);
    }
    .icon {
      font-size: 48px;
      line-height: 1;
      margin-bottom: 20px;
    }
    h1 {
      font-size: 22px;
      font-weight: 600;
      margin-bottom: 12px;
      color: var(--accent);
    }
    p {
      font-size: 14px;
      color: var(--muted);
      line-height: 1.6;
    }
  </style>
</head>
<body>
  <div class="card">
    <div class="icon">✅</div>
    <h1>Authorization complete</h1>
    <p>Align is connected to GitHub. You can close this tab and return to your terminal.</p>
  </div>
</body>
</html>`

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body)) //nolint
		token <- tok.AccessToken
	}
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return fmt.Errorf("unsupported platform")
	}
}
