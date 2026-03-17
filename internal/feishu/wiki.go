package feishu

import (
  "context"
  "encoding/json"
  "fmt"
  "io"
  "net/http"
  "net/url"
)

type WikiSpace struct {
  SpaceID string `json:"space_id"`
  Name    string `json:"name"`
}

type wikiSpacesResp struct {
  Code int    `json:"code"`
  Msg  string `json:"msg"`
  Data struct {
    HasMore   bool        `json:"has_more"`
    PageToken string      `json:"page_token"`
    Items     []WikiSpace `json:"items"`
  } `json:"data"`
}

func (c *Client) WikiSpaces(ctx context.Context, accessToken, pageToken string) (*wikiSpacesResp, error) {
  q := url.Values{}
  q.Set("page_size", "50")
  if pageToken != "" {
    q.Set("page_token", pageToken)
  }
  req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/open-apis/wiki/v2/spaces?"+q.Encode(), nil)
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
    return nil, fmt.Errorf("wiki spaces failed: http %d: %s", resp.StatusCode, string(b))
  }
  var out wikiSpacesResp
  if err := json.Unmarshal(b, &out); err != nil {
    return nil, err
  }
  if out.Code != 0 {
    return nil, fmt.Errorf("wiki spaces failed: code=%d msg=%s", out.Code, out.Msg)
  }
  return &out, nil
}

type WikiNode struct {
  SpaceID   string `json:"space_id"`
  NodeToken string `json:"node_token"`
  Parent    string `json:"parent_node_token"`
  Title     string `json:"title"`
  ObjType   string `json:"obj_type"`
  ObjToken  string `json:"obj_token"`
  HasChild  bool   `json:"has_child"`
}

type wikiNodesResp struct {
  Code int    `json:"code"`
  Msg  string `json:"msg"`
  Data struct {
    HasMore   bool       `json:"has_more"`
    PageToken string     `json:"page_token"`
    Items     []WikiNode `json:"items"`
  } `json:"data"`
}

func (c *Client) WikiSpaceNodes(ctx context.Context, accessToken, spaceID, parentNodeToken, pageToken string) (*wikiNodesResp, error) {
  q := url.Values{}
  q.Set("space_id", spaceID)
  q.Set("page_size", "50")
  if parentNodeToken != "" {
    q.Set("parent_node_token", parentNodeToken)
  }
  if pageToken != "" {
    q.Set("page_token", pageToken)
  }
  req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/open-apis/wiki/v2/space/get_node_list?"+q.Encode(), nil)
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
    return nil, fmt.Errorf("wiki nodes failed: http %d: %s", resp.StatusCode, string(b))
  }
  var out wikiNodesResp
  if err := json.Unmarshal(b, &out); err != nil {
    return nil, err
  }
  if out.Code != 0 {
    return nil, fmt.Errorf("wiki nodes failed: code=%d msg=%s", out.Code, out.Msg)
  }
  return &out, nil
}
