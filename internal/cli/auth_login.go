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
	"strings"

	"github.com/your-org/feishu-sync/internal/auth"
	"github.com/your-org/feishu-sync/internal/feishu"
)

type authLoginOptions struct {
	ListenHost   string
	Port         int
	CallbackPath string
	Timeout      time.Duration
	NoBrowser    bool

	Remote      bool
	RedirectURI string
}

func runAuthLogin(ctx context.Context, chdir, configPath string, opts authLoginOptions, out io.Writer, errOut io.Writer) error {
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

	if listenHost == "" {
		listenHost = "127.0.0.1"
	}
	if port == 0 {
		port = 18900
	}
	if callbackPath == "" {
		callbackPath = "/callback"
	}
	if callbackPath[0] != '/' {
		callbackPath = "/" + callbackPath
	}

	redirectURI := opts.RedirectURI
	if redirectURI == "" {
		redirectURI = authLoginRedirectURI(listenHost, port, callbackPath)
		if opts.Remote {
			fmt.Fprintln(out, "WARNING: --remote used without --redirect-uri; using local callback redirect (likely not whitelisted):", redirectURI)
		}
	}

	state := feishu.RandomState()
	authURL, err := feishu.OAuthAuthorizeURL(cfg.App.ID, redirectURI, state)
	if err != nil {
		return err
	}

	fmt.Fprintln(out, "Auth login:")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Local callback (recommended on your laptop):")
	fmt.Fprintf(out, "  feishu-sync auth login --host %s --port %d --callback-path %s\n", listenHost, port, callbackPath)
	fmt.Fprintf(out, "  redirect_uri: %s\n", authLoginRedirectURI(listenHost, port, callbackPath))
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Remote/manual (recommended on a server):")
	fmt.Fprintln(out, "  feishu-sync auth login --remote --redirect-uri <WHITELISTED_REDIRECT_URI>")
	fmt.Fprintf(out, "  redirect_uri currently used: %s\n", redirectURI)
	fmt.Fprintln(out, "  NOTE: after you authorize, the browser may show 404/blank — this is normal.")
	fmt.Fprintln(out, "  Copy the FULL URL from the address bar (must include code= and state=),")
	fmt.Fprintln(out, "  then paste it back into this terminal.")
	fmt.Fprintln(out)

	var code string
	if opts.Remote {
		fmt.Fprintln(out, "Open this URL to authorize:")
		fmt.Fprintln(out, authURL)
		if !opts.NoBrowser {
			_ = openBrowser(authURL)
		}
		fmt.Fprintln(out)
		fmt.Fprintln(out, "NOTE: after you authorize, the browser may show 404/blank — this is normal.")
		fmt.Fprintln(out, "Copy the FULL URL from the address bar (must include code= and state=),")
		fmt.Fprintln(out, "then paste it back into this terminal.")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Paste the full callback URL (containing ?code=...&state=...) OR paste just the code:")
		input, err := readLine(ctx, "")
		if err != nil {
			return err
		}
		c, st, err := parseOAuthPastedInput(input)
		if err != nil {
			return err
		}
		if st != "" {
			if st != state {
				return errors.New("invalid oauth state")
			}
		} else if strings.Contains(strings.TrimSpace(input), "://") {
			return errors.New("callback url missing state")
		} else {
			fmt.Fprintln(out, "WARNING: no state provided (raw code pasted); cannot validate state")
		}
		code = c
	} else {
		addr := fmt.Sprintf("%s:%d", listenHost, port)
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("cannot listen on %s: %w (port may be in use; try --port %d)", addr, err, port+1)
		}
		defer ln.Close()

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
	}

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
	fmt.Fprintln(out, "token:", store.Path)
	if errOut != nil {
		fmt.Fprintln(errOut, "Token saved:", store.Path)
		fmt.Fprintln(errOut, "Next:")
		fmt.Fprintln(errOut, "  feishu-sync pull --dry-run   # preview")
		fmt.Fprintln(errOut, "  feishu-sync pull            # export")
	}
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
