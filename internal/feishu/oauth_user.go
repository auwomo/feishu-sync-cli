package feishu

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type UserAccessTokenResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		AccessToken           string `json:"access_token"`
		TokenType             string `json:"token_type"`
		ExpiresIn             int    `json:"expires_in"`
		RefreshToken          string `json:"refresh_token"`
		RefreshExpiresIn      int    `json:"refresh_expires_in"`
		Scope                 string `json:"scope"`
		TenantKey             string `json:"tenant_key"`
		OpenID                string `json:"open_id"`
		UnionID               string `json:"union_id"`
		UserID                string `json:"user_id"`
		Name                  string `json:"name"`
		AvatarURL             string `json:"avatar_url"`
		AvatarThumb           string `json:"avatar_thumb"`
		AvatarMiddle          string `json:"avatar_middle"`
		AvatarBig             string `json:"avatar_big"`
		EnName                string `json:"en_name"`
		Email                 string `json:"email"`
		Mobile                string `json:"mobile"`
		EnterpriseEmail       string `json:"enterprise_email"`
		TenantName            string `json:"tenant_name"`
		EmployeeNo            string `json:"employee_no"`
		UserTenantID          string `json:"user_tenant_id"`
		CurrentEnterpriseName string `json:"current_enterprise_name"`
	} `json:"data"`
}

type UserRefreshTokenResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		AccessToken      string `json:"access_token"`
		TokenType        string `json:"token_type"`
		ExpiresIn        int    `json:"expires_in"`
		RefreshToken     string `json:"refresh_token"`
		Scope            string `json:"scope"`
		TenantKey        string `json:"tenant_key"`
		OpenID           string `json:"open_id"`
		UnionID          string `json:"union_id"`
		UserID           string `json:"user_id"`
		Name             string `json:"name"`
		AvatarURL        string `json:"avatar_url"`
		AvatarThumb      string `json:"avatar_thumb"`
		AvatarMiddle     string `json:"avatar_middle"`
		AvatarBig        string `json:"avatar_big"`
		EnName           string `json:"en_name"`
		Email            string `json:"email"`
		Mobile           string `json:"mobile"`
		EnterpriseEmail  string `json:"enterprise_email"`
		RefreshExpiresIn int    `json:"refresh_expires_in"`
	} `json:"data"`
}

// OAuthAuthorizeURL builds the Feishu OAuth authorization URL.
func OAuthAuthorizeURL(appID, redirectURI, state string) (string, error) {
	u, err := url.Parse("https://open.feishu.cn/open-apis/authen/v1/index")
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("app_id", appID)
	q.Set("redirect_uri", redirectURI)
	if state != "" {
		q.Set("state", state)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (c *Client) ExchangeUserCode(ctx context.Context, appID, appSecret, code, redirectURI string) (*UserAccessTokenResponse, error) {
	body, _ := json.Marshal(map[string]string{
		"app_id":       appID,
		"app_secret":   appSecret,
		"grant_type":   "authorization_code",
		"code":         code,
		"redirect_uri": redirectURI,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/open-apis/authen/v2/oauth/token", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("oauth token exchange failed: http %d: %s", resp.StatusCode, string(b))
	}
	var tr UserAccessTokenResponse
	if err := json.Unmarshal(b, &tr); err != nil {
		return nil, err
	}
	if tr.Code != 0 {
		return nil, fmt.Errorf("oauth token exchange failed: code=%d msg=%s", tr.Code, tr.Msg)
	}
	if tr.Data.AccessToken == "" {
		return nil, fmt.Errorf("oauth token exchange missing access_token (code=%d msg=%s)", tr.Code, tr.Msg)
	}
	return &tr, nil
}

func (c *Client) RefreshUserToken(ctx context.Context, appID, appSecret, refreshToken string) (*UserRefreshTokenResponse, error) {
	body, _ := json.Marshal(map[string]string{
		"app_id":     appID,
		"app_secret": appSecret,
		"grant_type": "refresh_token",
		"refresh_token": refreshToken,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/open-apis/authen/v2/oauth/token", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("oauth refresh failed: http %d: %s", resp.StatusCode, string(b))
	}
	var tr UserRefreshTokenResponse
	if err := json.Unmarshal(b, &tr); err != nil {
		return nil, err
	}
	if tr.Code != 0 {
		return nil, fmt.Errorf("oauth refresh failed: code=%d msg=%s", tr.Code, tr.Msg)
	}
	return &tr, nil
}

func ExpiresAt(now time.Time, expiresIn int) time.Time {
	if expiresIn <= 0 {
		return time.Time{}
	}
	return now.Add(time.Duration(expiresIn) * time.Second)
}

func ParseOAuthCallback(r *http.Request) (code string, state string, err error) {
	q := r.URL.Query()
	code = q.Get("code")
	state = q.Get("state")	
	if code == "" {
		// sometimes error params are present
		errDesc := q.Get("error_description")
		errCode := q.Get("error")
		if errCode != "" || errDesc != "" {
			return "", state, fmt.Errorf("oauth error: %s %s", errCode, errDesc)
		}
		return "", state, fmt.Errorf("missing code in callback")
	}
	return code, state, nil
}

func RandomState() string {
	// cheap + deterministic-free enough for CLI flow; can be swapped for crypto/rand
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}
