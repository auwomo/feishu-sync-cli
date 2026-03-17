package auth

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

// UserToken is a persisted user-level credential set.
// See Feishu user_access_token OAuth flow.
// Times are stored as RFC3339.
type UserToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
	Scope        string    `json:"scope,omitempty"`
}

func (t UserToken) Valid(now time.Time) bool {
	if t.AccessToken == "" {
		return false
	}
	if t.ExpiresAt.IsZero() {
		return true
	}
	// keep a small buffer
	return now.Before(t.ExpiresAt.Add(-30 * time.Second))
}

type TokenStore struct {
	Path string
}

func NewTokenStore(path string) *TokenStore {
	return &TokenStore{Path: path}
}

func (s *TokenStore) Load() (*UserToken, error) {
	b, err := os.ReadFile(s.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, os.ErrNotExist
		}
		return nil, err
	}
	var t UserToken
	if err := json.Unmarshal(b, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *TokenStore) Save(t *UserToken) error {
	if t == nil {
		return errors.New("token is nil")
	}
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	// 0600 because it contains access/refresh tokens
	return os.WriteFile(s.Path, b, 0o600)
}
