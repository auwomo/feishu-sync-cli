package cli

import (
	"context"
	"path/filepath"

	"github.com/your-org/feishu-sync/internal/auth"
	"github.com/your-org/feishu-sync/internal/config"
)

// resolveAccessToken returns either tenant_access_token or user_access_token based on config.
// It also handles loading/refreshing user token if needed.
func resolveAccessToken(ctx context.Context, wsPath string, cfg *config.Config, appSecret string) (token string, isUser bool, err error) {
	client := feishuNewClient()
	mode := authMode(cfg.Auth.Mode)
	if mode == "user" {
		store := auth.NewTokenStore(filepath.Join(wsPath, "token.json"))
		p := auth.UserTokenProvider{Client: client, Store: store, AppID: cfg.App.ID, AppSecret: appSecret}
		tok, err := p.AccessToken(ctx)
		return tok, true, err
	}
	// default: tenant
	tok, err := client.TenantAccessToken(ctx, cfg.App.ID, appSecret)
	return tok, false, err
}
