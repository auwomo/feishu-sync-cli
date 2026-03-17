package discovery

import (
	"context"

	"github.com/your-org/feishu-sync/internal/feishu"
)

type wikiClient interface {
	WikiSpaces(ctx context.Context, accessToken, pageToken string) (*feishu.wikiSpacesResp, error)
	WikiSpaceNodes(ctx context.Context, accessToken, spaceID, parentNodeToken, pageToken string) (*feishu.wikiNodesResp, error)
}
