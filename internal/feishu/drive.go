package feishu

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type DriveFolderChildrenItem struct {
	Token string `json:"token"`
	Name  string `json:"name"`
	Type  string `json:"type"`
}

type DriveFolderChildrenResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		HasMore   bool                      `json:"has_more"`
		PageToken string                    `json:"page_token"`
		Files     []DriveFolderChildrenItem `json:"files"`
	} `json:"data"`
}

func (c *Client) DriveFolderChildren(ctx context.Context, tenantToken, folderToken, pageToken string) (*DriveFolderChildrenResponse, error) {
	q := url.Values{}
	q.Set("folder_token", folderToken)
	q.Set("page_size", "200")
	if pageToken != "" {
		q.Set("page_token", pageToken)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/open-apis/drive/v1/files?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tenantToken)

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
		return nil, fmt.Errorf("drive list failed: http %d: %s", resp.StatusCode, string(b))
	}

	var out DriveFolderChildrenResponse
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	if out.Code != 0 {
		return nil, fmt.Errorf("drive list failed: code=%d msg=%s", out.Code, out.Msg)
	}
	return &out, nil
}
