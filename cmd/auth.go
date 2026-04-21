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
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Align — Authorized</title>
  <style>
    :root {
      color-scheme: light dark;
    }
    body {
      margin: 0;
      padding: 0;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
      background-color: light-dark(#f6f8fa, #0d1117);
      min-height: 100vh;
      display: flex;
      flex-direction: column;
    }
    .header {
      background-color: #161b22;
      padding: 20px 24px 80px;
      display: flex;
      flex-direction: column;
      align-items: flex-start;
      gap: 12px;
    }
    .logo {
      font-family: ui-monospace, SFMono-Regular, "SF Mono", Menlo, monospace;
      font-size: 22px;
      font-weight: 700;
      color: #e6edf3;
      letter-spacing: -0.5px;
    }
    .title {
      color: #e6edf3;
      margin: 0 0 -30px;
      width: 100%;
      font-weight: 700;
      font-size: 24px;
      line-height: 30px;
      text-align: center;
      letter-spacing: -0.48px;
    }
    .main {
      flex: 1;
      display: flex;
      align-items: flex-start;
      justify-content: center;
      padding: 0 20px 60px;
      margin-top: -30px;
    }
    .container {
      background-color: light-dark(#ffffff, #161b22);
      border-radius: 16px;
      padding: 20px;
      max-width: 600px;
      width: 100%;
      border: 1px solid light-dark(#d0d7de, #30363d);
    }
    .success-text,
    .error-text {
      color: light-dark(#1f2328, #c9d1d9);
      font-size: 14px;
      line-height: 20px;
    }
    .success-content {
      background-color: light-dark(#dafbe1, #0d1f17);
      border: 2px solid light-dark(#1a7f37, #3fb950);
      border-radius: 12px;
      padding: 14px 16px;
      display: flex;
      align-items: flex-start;
      gap: 8px;
    }
    .error-content {
      background-color: light-dark(#fff5f5, #1f0000);
      border: 2px solid light-dark(#cf222e, #f85149);
      border-radius: 12px;
      padding: 14px 16px;
      display: flex;
      align-items: flex-start;
      gap: 8px;
    }
    .success-content svg,
    .error-content svg {
      flex-shrink: 0;
    }
    .hidden {
      display: none;
    }
  </style>
</head>
<body>
  <header class="header">
    <div class="logo">align</div>
    <h1 class="title">GitHub Authorization</h1>
  </header>
  <main class="main">
    <div class="container">
      <div id="success-message" class="success-content">
        <svg width="16" height="16" viewBox="0 0 16 16" xmlns="http://www.w3.org/2000/svg" focusable="false" aria-hidden="true">
          <circle cx="8" cy="8" r="7" fill="none" stroke="light-dark(#1a7f37, #3fb950)" stroke-width="2"></circle>
          <path d="M4.5 7.5 7 10l4-5" fill="none" stroke-linejoin="round" stroke="light-dark(#1a7f37, #3fb950)" stroke-width="2"></path>
        </svg>
        <div class="success-text">
          Authorization complete. Align is connected to GitHub. You can close this tab and return to your terminal.
        </div>
      </div>
      <div id="error-message" class="error-content hidden">
        <svg width="16" height="16" xmlns="http://www.w3.org/2000/svg" focusable="false" aria-hidden="true">
          <circle cx="8" cy="8" r="7" fill="none" stroke="light-dark(#cf222e, #f85149)" stroke-width="2"></circle>
          <path d="m5.5 5.5 5 5M10.5 5.5l-5 5" stroke="light-dark(#cf222e, #f85149)" stroke-width="2"></path>
        </svg>
        <div class="error-text">Authorization failed: <span id="error-text"></span></div>
      </div>
    </div>
  </main>
  <script>
    window.onload = () => {
      const params = new URLSearchParams(window.location.search);
      const error = params.get('error');
      if (error) {
        document.getElementById('success-message').classList.add('hidden');
        document.getElementById('error-message').classList.remove('hidden');
        document.getElementById('error-text').innerText = error;
      }
    };
  </script>
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
