package cli

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/your-org/feishu-sync/internal/auth"
	"github.com/your-org/feishu-sync/internal/feishu"
)

type authLoginOptions struct {
	Timeout   time.Duration
	NoBrowser bool

	// For manual flow only: override redirect_uri (must be whitelisted in Feishu app).
	RedirectURI string
	Verbose     bool
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

	// Fixed local callback for the built-in local flow.
	localRedirectURI := authLoginRedirectURI("127.0.0.1", 18900, "/callback")

	effectiveRedirectURI := localRedirectURI
	if opts.RedirectURI != "" {
		effectiveRedirectURI = opts.RedirectURI
	}

	state := feishu.RandomState()
	authURL, err := feishu.OAuthAuthorizeURL(cfg.App.ID, effectiveRedirectURI, state)
	if err != nil {
		return err
	}

	printAuthLoginOptions(out, opts, authURL, effectiveRedirectURI, localRedirectURI)

	// Offer manual flow first (user can paste immediately). Empty input skips.
	code, err := tryManualAuth(ctx, out, state)
	if err != nil || code == "" {
		// Fall back to local callback flow.
		code, err = tryLocalAuth(ctx, out, authURL, state, opts.NoBrowser, opts.Timeout)
		if err != nil {
			return err
		}
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
