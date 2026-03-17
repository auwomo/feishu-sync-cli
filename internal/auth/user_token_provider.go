package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/your-org/feishu-sync/internal/feishu"
)

type UserTokenProvider struct {
	Client    *feishu.Client
	Store     *TokenStore
	AppID     string
	AppSecret string
	Now       func() time.Time
}

func (p *UserTokenProvider) AccessToken(ctx context.Context) (string, error) {
	if p.Client == nil || p.Store == nil {
		return "", errors.New("token provider missing client/store")
	}
	now := time.Now
	if p.Now != nil {
		now = p.Now
	}
	t, err := p.Store.Load()
	if err != nil {
		return "", fmt.Errorf("load user token: %w (run `feishu-sync auth login`) ", err)
	}
	if t.Valid(now()) {
		return t.AccessToken, nil
	}
	if t.RefreshToken == "" {
		return "", errors.New("user token expired and refresh_token missing (run `feishu-sync auth login`) ")
	}
	r, err := p.Client.RefreshUserToken(ctx, p.AppID, p.AppSecret, t.RefreshToken)
	if err != nil {
		return "", err
	}
	nt := &UserToken{
		AccessToken:  r.Data.AccessToken,
		RefreshToken: r.Data.RefreshToken,
		ExpiresAt:    feishu.ExpiresAt(now(), r.Data.ExpiresIn),
		TokenType:    r.Data.TokenType,
		Scope:        r.Data.Scope,
	}
	if err := p.Store.Save(nt); err != nil {
		return "", err
	}
	return nt.AccessToken, nil
}
