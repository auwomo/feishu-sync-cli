package feishu

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	BaseURL string
	http    *http.Client
}

func NewClient(httpClient *http.Client) *Client {
	c := httpClient
	if c == nil {
		c = &http.Client{Timeout: 30 * time.Second}
	}
	return &Client{BaseURL: "https://open.feishu.cn", http: c}
}

type TokenResponse struct {
	TenantAccessToken string `json:"tenant_access_token"`
	Expire            int    `json:"expire"`
	Code              int    `json:"code"`
	Msg               string `json:"msg"`
}

func (c *Client) TenantAccessToken(ctx context.Context, appID, appSecret string) (string, error) {
	if appID == "" {
		return "", errors.New("app id is required")
	}
	if appSecret == "" {
		return "", errors.New("app secret is required")
	}

	body, _ := json.Marshal(map[string]string{
		"app_id":     appID,
		"app_secret": appSecret,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/open-apis/auth/v3/tenant_access_token/internal", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("token request failed: http %d: %s", resp.StatusCode, string(b))
	}

	var tr TokenResponse
	if err := json.Unmarshal(b, &tr); err != nil {
		return "", err
	}
	if tr.Code != 0 {
		return "", fmt.Errorf("token request failed: code=%d msg=%s", tr.Code, tr.Msg)
	}
	if tr.TenantAccessToken == "" {
		return "", errors.New("token response missing tenant_access_token")
	}
	return tr.TenantAccessToken, nil
}
