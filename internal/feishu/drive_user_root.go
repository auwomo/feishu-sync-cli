package feishu

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type driveRootResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Token string `json:"token"`
	} `json:"data"`
}

// DriveUserRootFolderToken returns the current user's personal root folder token.
// API: GET /open-apis/drive/explorer/v2/root_folder/meta
func (c *Client) DriveUserRootFolderToken(ctx context.Context, userAccessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/open-apis/drive/explorer/v2/root_folder/meta", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+userAccessToken)

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
		return "", fmt.Errorf("drive root meta failed: http %d: %s", resp.StatusCode, string(b))
	}
	var rr driveRootResp
	if err := json.Unmarshal(b, &rr); err != nil {
		return "", err
	}
	if rr.Code != 0 {
		return "", fmt.Errorf("drive root meta failed: code=%d msg=%s", rr.Code, rr.Msg)
	}
	return rr.Data.Token, nil
}
