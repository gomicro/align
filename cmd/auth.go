package cmd

import (
	"context"
	"crypto/tls"
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
	authCmd.Flags().BoolVarP(&reapprove, "force", "f", false, "force align to reauth")
}

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "auth with github",
	Long:  `authorize align against github`,
	RunE:  authFunc,
}

const (
	state = "9292a768-34bf-4002-8a69-8ace5302709d"
)

func authFunc(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if clientID == "" || clientSecret == "" {
		cmd.SilenceUsage = true
		return fmt.Errorf("client id and secret must be baked into the binary, and are not present")
	}

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

	go startServer(ctx, listener, conf, token)

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

func startServer(ctx context.Context, listener net.Listener, conf *oauth2.Config, token chan string) {
	http.HandleFunc("/auth", authHandler(ctx, conf, token))

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

func authHandler(ctx context.Context, conf *oauth2.Config, token chan string) func(w http.ResponseWriter, req *http.Request) {
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

		body := `<html>
	<body>
		<h1>Config file updated, you can close this window.</h1>
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
