package discovery

import (
	"context"
	"testing"

	"github.com/your-org/feishu-sync/internal/feishu"
	"github.com/your-org/feishu-sync/internal/manifest"
)

type fakeWiki struct {
	spaces []*feishu.WikiSpacesResp
	nodes  map[string][]*feishu.WikiNodesResp // key: parent token ("" for root)
}

func (f *fakeWiki) WikiSpaces(ctx context.Context, accessToken, pageToken string) (*feishu.WikiSpacesResp, error) {
	if len(f.spaces) == 0 {
		return &feishu.WikiSpacesResp{}, nil
	}
	resp := f.spaces[0]
	f.spaces = f.spaces[1:]
	return resp, nil
}

func (f *fakeWiki) WikiSpaceNodes(ctx context.Context, accessToken, spaceID, parentNodeToken, pageToken string) (*feishu.WikiNodesResp, error) {
	pages := f.nodes[parentNodeToken]
	if len(pages) == 0 {
		return &feishu.WikiNodesResp{}, nil
	}
	resp := pages[0]
	f.nodes[parentNodeToken] = pages[1:]
	return resp, nil
}

func TestDiscoverWikiSpaces_Pagination(t *testing.T) {
	f := &fakeWiki{spaces: []*feishu.WikiSpacesResp{
		{Data: struct {
			HasMore   bool        `json:"has_more"`
			PageToken string      `json:"page_token"`
			Items     []feishu.WikiSpace `json:"items"`
		}{HasMore: true, PageToken: "p2", Items: []feishu.WikiSpace{{SpaceID: "s1", Name: "A"}}}},
		{Data: struct {
			HasMore   bool        `json:"has_more"`
			PageToken string      `json:"page_token"`
			Items     []feishu.WikiSpace `json:"items"`
		}{HasMore: false, PageToken: "", Items: []feishu.WikiSpace{{SpaceID: "s2", Name: "B"}}}},
	}}

	spaces, err := DiscoverWikiSpaces(context.Background(), f, "tok")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(spaces) != 2 {
		t.Fatalf("want 2, got %d", len(spaces))
	}
}

func TestDiscoverWikiTree_PathMapping(t *testing.T) {
	f := &fakeWiki{nodes: map[string][]*feishu.WikiNodesResp{
		"": {
			{Data: struct {
				HasMore   bool       `json:"has_more"`
				PageToken string     `json:"page_token"`
				Items     []feishu.WikiNode `json:"items"`
			}{HasMore: false, Items: []feishu.WikiNode{
				{SpaceID: "s", NodeToken: "n1", Title: "Folder", ObjType: "docx", ObjToken: "d1", HasChild: true},
			}}},
		},
		"n1": {
			{Data: struct {
				HasMore   bool       `json:"has_more"`
				PageToken string     `json:"page_token"`
				Items     []feishu.WikiNode `json:"items"`
			}{HasMore: false, Items: []feishu.WikiNode{
				{SpaceID: "s", NodeToken: "n2", Title: "Child", ObjType: "docx", ObjToken: "d2", HasChild: false},
			}}},
		},
	}}

	items, errs := DiscoverWikiTree(context.Background(), f, "tok", "s")
	if len(errs) != 0 {
		t.Fatalf("errs: %+v", errs)
	}
	if len(items) != 2 {
		t.Fatalf("want 2, got %d", len(items))
	}

	var folder, child manifest.WikiItem
	for _, it := range items {
		if it.NodeToken == "n1" {
			folder = it
		}
		if it.NodeToken == "n2" {
			child = it
		}
	}
	if folder.Path != "" {
		t.Fatalf("folder path want empty, got %q", folder.Path)
	}
	if child.Path != "Folder" {
		t.Fatalf("child path want Folder, got %q", child.Path)
	}
}
