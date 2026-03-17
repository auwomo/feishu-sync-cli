package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/your-org/feishu-sync/internal/auth"
	"github.com/your-org/feishu-sync/internal/feishu"
)

type authLoginOptions struct {
	ListenHost    string
	Port          int
	CallbackPath  string
	Timeout       time.Duration
	NoBrowser     bool
}

func runAuthLogin(ctx context.Context, chdir, configPath string, opts authLoginOptions, out io.Writer) error {
	ws, cfg, err := loadWorkspaceAndConfig(chdir, configPath)
	if err != nil {
		return err
	}
	secret, err := resolveAppSecret(ws, cfg)
	if err != nil {
		return err
	}

	listenHost := opts.ListenHost
	port := opts.Port
	callbackPath := opts.CallbackPath

	listenHost = opts.ListenHost
	if listenHost == "" {
		listenHost = "127.0.0.1"
	}
	port = opts.Port
	if port == 0 {
		port = 18900
	}
	callbackPath = opts.CallbackPath
	if callbackPath == "" {
		callbackPath = "/callback"
	}
	if callbackPath[0] != '/' {
		callbackPath = "/" + callbackPath
	}

	addr := fmt.Sprintf("%s:%d", listenHost, port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("cannot listen on %s: %w (port may be in use; try --port %d)", addr, err, port+1)
	}
	defer ln.Close()
	redirectURI := authLoginRedirectURI(listenHost, port, callbackPath)

	state := feishu.RandomState()
	authURL, err := feishu.OAuthAuthorizeURL(cfg.App.ID, redirectURI, state)
	if err != nil {
		return err
	}

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc(callbackPath, func(w http.ResponseWriter, r *http.Request) {
		code, st, err := feishu.ParseOAuthCallback(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("Auth failed: " + err.Error()))
			errCh <- err
			return
		}
		if st != state {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("Auth failed: invalid state"))
			errCh <- errors.New("invalid oauth state")
			return
		}
		_, _ = w.Write([]byte("OK. You can close this tab and return to the terminal."))
		codeCh <- code
	})

	srv := &http.Server{Handler: mux}
	go func() {
		_ = srv.Serve(ln)
	}()

	fmt.Fprintln(out, "Open this URL to authorize:")
	fmt.Fprintln(out, authURL)

	if !opts.NoBrowser {
		_ = openBrowser(authURL)
	}

	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}
	ctx2, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var code string
	select {
	case code = <-codeCh:
	case err := <-errCh:
		_ = srv.Shutdown(context.Background())
		return err
	case <-ctx2.Done():
		_ = srv.Shutdown(context.Background())
		return fmt.Errorf("auth timeout after %s", timeout)
	}
	_ = srv.Shutdown(context.Background())

	client := feishuNewClient()
	tenantTok, err := client.TenantAccessToken(ctx, cfg.App.ID, secret)
	if err != nil {
		return err
	}
	r, err := client.ExchangeUserCode(ctx, tenantTok, code)
	if err != nil {
		return err
	}

	tok := &auth.UserToken{
		AccessToken:  r.Data.AccessToken,
		RefreshToken: r.Data.RefreshToken,
		ExpiresAt:    feishu.ExpiresAt(time.Now(), r.Data.ExpiresIn),
		TokenType:    r.Data.TokenType,
		Scope:        r.Data.Scope,
	}
	store := auth.NewTokenStore(filepath.Join(ws.Path(), "token.json"))
	if err := store.Save(tok); err != nil {
		return err
	}

	fmt.Fprintln(out, "OK: user token saved to", store.Path)
	return nil
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}
