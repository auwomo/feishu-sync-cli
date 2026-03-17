package discovery

import (
  "context"
  "path"
  "sort"
  "strings"

  "github.com/your-org/feishu-sync/internal/feishu"
  "github.com/your-org/feishu-sync/internal/manifest"
)

func DiscoverWikiSpaces(ctx context.Context, client wikiClient, token string) ([]feishu.WikiSpace, error) {
  var out []feishu.WikiSpace
  pageToken := ""
  for {
    resp, err := client.WikiSpaces(ctx, token, pageToken)
    if err != nil {
      return nil, err
    }
    out = append(out, resp.Data.Items...)
    if !resp.Data.HasMore || resp.Data.PageToken == "" {
      break
    }
    pageToken = resp.Data.PageToken
  }
  return out, nil
}

func DiscoverWikiTree(ctx context.Context, client wikiClient, token, spaceID string) ([]manifest.WikiItem, []manifest.DiscoveryError) {
  items := []manifest.WikiItem{}
  errs := []manifest.DiscoveryError{}

  // Build node map + parent->children via BFS pagination per parent.
  type node = feishu.WikiNode
  nodeByTok := map[string]node{}
  children := map[string][]string{} // parent -> []childTok

  queue := []string{""} // root parent token
  seenParent := map[string]bool{}

  for len(queue) > 0 {
    parent := queue[0]
    queue = queue[1:]
    if seenParent[parent] {
      continue
    }
    seenParent[parent] = true

    pageToken := ""
    for {
      resp, err := client.WikiSpaceNodes(ctx, token, spaceID, parent, pageToken)
      if err != nil {
        errs = append(errs, manifest.DiscoveryError{Scope: "wiki", Token: spaceID + ":" + parent, Message: "failed to list nodes: " + err.Error()})
        break
      }
      for _, n := range resp.Data.Items {
        nodeByTok[n.NodeToken] = n
        children[parent] = append(children[parent], n.NodeToken)
        if n.HasChild {
          queue = append(queue, n.NodeToken)
        }
      }
      if !resp.Data.HasMore || resp.Data.PageToken == "" {
        break
      }
      pageToken = resp.Data.PageToken
    }
  }

  // Stable ordering.
  for k := range children {
    sort.Strings(children[k])
  }

  // Path recursion.
  var build func(parent string, prefix string)
  build = func(parent string, prefix string) {
    for _, tok := range children[parent] {
      n := nodeByTok[tok]
      title := strings.TrimSpace(n.Title)
      if title == "" {
        title = "untitled"
      }
      curDir := prefix
      if prefix == "" {
        curDir = ""
      }
      // folders implied by titles along the way
      itemPath := path.Join(curDir, title)
      dir := path.Dir(itemPath)
      if dir == "." {
        dir = ""
      }
      items = append(items, manifest.WikiItem{SpaceID: spaceID, NodeToken: tok, Title: title, Path: dir, ObjType: n.ObjType, ObjToken: n.ObjToken})
      build(tok, path.Join(curDir, title))
    }
  }
  build("", "")

  return items, errs
}
