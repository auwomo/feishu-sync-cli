package manifest

type WikiItem struct {
  SpaceID   string `json:"space_id"`
  SpaceName string `json:"space_name"`
  NodeToken string `json:"node_token"`
  Title     string `json:"title"`
  // Path is the directory path under the wiki space (excluding Title)
  Path string `json:"path"`

  ObjType  string `json:"obj_type"`
  ObjToken string `json:"obj_token"`
}
