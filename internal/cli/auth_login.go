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
	ListenHost string
	Port       int // 0 means random
	Timeout    time.Duration
	NoBrowser  bool
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
	if listenHost == "" {
		listenHost = "127.0.0.1"
	}
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", listenHost, opts.Port))
	if err != nil {
		return err
	}
	defer ln.Close()
	addr := ln.Addr().String()
	redirectURI := fmt.Sprintf("http://%s/callback", addr)

	state := feishu.RandomState()
	authURL, err := feishu.OAuthAuthorizeURL(cfg.App.ID, redirectURI, state)
	if err != nil {
		return err
	}

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
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
	r, err := client.ExchangeUserCode(ctx, cfg.App.ID, secret, code, redirectURI)
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
