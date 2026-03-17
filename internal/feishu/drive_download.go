package feishu

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// DriveDownload downloads a Drive media/file by token. Caller must close body.
func (c *Client) DriveDownload(ctx context.Context, accessToken, fileToken string) (filename string, contentType string, body io.ReadCloser, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/open-apis/drive/v1/medias/"+fileToken+"/download", nil)
	if err != nil {
		return "", "", nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", "", nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return "", "", nil, fmt.Errorf("drive download failed: http %d: %s", resp.StatusCode, string(b))
	}

	// Note: Feishu doesn't always include a filename.
	return resp.Header.Get("Content-Disposition"), resp.Header.Get("Content-Type"), resp.Body, nil
}
