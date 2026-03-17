package feishu

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type DocxBlock struct {
	BlockID   string   `json:"block_id"`
	ParentID  string   `json:"parent_id"`
	Children  []string `json:"children"`
	BlockType int      `json:"block_type"`

	Page *struct {
		Elements []DocxElement `json:"elements"`
	} `json:"page,omitempty"`
	Text *struct {
		Elements []DocxElement `json:"elements"`
		Style    any           `json:"style"`
	} `json:"text,omitempty"`
	Heading1 *struct {
		Elements []DocxElement `json:"elements"`
		Style    any           `json:"style"`
	} `json:"heading1,omitempty"`
	Heading2 *struct {
		Elements []DocxElement `json:"elements"`
		Style    any           `json:"style"`
	} `json:"heading2,omitempty"`
	Heading3 *struct {
		Elements []DocxElement `json:"elements"`
		Style    any           `json:"style"`
	} `json:"heading3,omitempty"`
	Heading4 *struct {
		Elements []DocxElement `json:"elements"`
		Style    any           `json:"style"`
	} `json:"heading4,omitempty"`
	Heading5 *struct {
		Elements []DocxElement `json:"elements"`
		Style    any           `json:"style"`
	} `json:"heading5,omitempty"`
	Heading6 *struct {
		Elements []DocxElement `json:"elements"`
		Style    any           `json:"style"`
	} `json:"heading6,omitempty"`

	Bullet *struct {
		Elements []DocxElement `json:"elements"`
	} `json:"bullet,omitempty"`
	Ordered *struct {
		Elements []DocxElement `json:"elements"`
	} `json:"ordered,omitempty"`
	Code *struct {
		Elements []DocxElement `json:"elements"`
		Style    struct {
			Language int `json:"language"`
		} `json:"style"`
	} `json:"code,omitempty"`

	Quote *struct{} `json:"quote,omitempty"`

	Image *struct {
		Token  string `json:"token"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	} `json:"image,omitempty"`
	File *struct {
		Name  string `json:"name"`
		Token string `json:"token"`
	} `json:"file,omitempty"`
}

type DocxElement struct {
	TextRun *struct {
		Content string `json:"content"`
		Style   struct {
			Bold          bool `json:"bold"`
			Italic        bool `json:"italic"`
			Underline     bool `json:"underline"`
			Strikethrough bool `json:"strikethrough"`
			InlineCode    bool `json:"inline_code"`
			Link          *struct {
				URL string `json:"url"`
			} `json:"link,omitempty"`
		} `json:"text_element_style"`
	} `json:"text_run,omitempty"`
	Equation *struct {
		Content string `json:"content"`
	} `json:"equation,omitempty"`
	MentionDoc *struct {
		Title string `json:"title"`
		URL   string `json:"url"`
	} `json:"mention_doc,omitempty"`
}

type docxBlocksResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		HasMore   bool        `json:"has_more"`
		PageToken string      `json:"page_token"`
		Items     []DocxBlock `json:"items"`
	} `json:"data"`
}

func (c *Client) DocxListBlocks(ctx context.Context, accessToken, documentID, pageToken string) (*docxBlocksResp, error) {
	q := url.Values{}
	q.Set("page_size", "200")
	if pageToken != "" {
		q.Set("page_token", pageToken)
	}
	u := c.BaseURL + "/open-apis/docx/v1/documents/" + documentID + "/blocks?" + q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

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
		return nil, fmt.Errorf("docx list blocks failed: http %d: %s", resp.StatusCode, string(b))
	}
	var out docxBlocksResp
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	if out.Code != 0 {
		return nil, fmt.Errorf("docx list blocks failed: code=%d msg=%s", out.Code, out.Msg)
	}
	return &out, nil
}

func (c *Client) DocxAllBlocks(ctx context.Context, accessToken, documentID string) ([]DocxBlock, error) {
	var all []DocxBlock
	pt := ""
	for {
		r, err := c.DocxListBlocks(ctx, accessToken, documentID, pt)
		if err != nil {
			return nil, err
		}
		all = append(all, r.Data.Items...)
		if !r.Data.HasMore || r.Data.PageToken == "" {
			break
		}
		pt = r.Data.PageToken
	}
	return all, nil
}

func DecodeDocxBlockRaw(b []byte) (DocxBlock, error) {
	var blk DocxBlock
	err := json.Unmarshal(b, &blk)
	return blk, err
}
