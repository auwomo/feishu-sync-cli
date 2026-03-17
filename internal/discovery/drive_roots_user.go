package discovery

import (
	"context"

	"github.com/your-org/feishu-sync/internal/feishu"
)

// DiscoverUserDriveRoots attempts to find the current user's accessible root folder tokens.
// For now, we treat "root" as the user's personal space root.
func DiscoverUserDriveRoots(ctx context.Context, client *feishu.Client, userAccessToken string) ([]string, error) {
	root, err := client.DriveUserRootFolderToken(ctx, userAccessToken)
	if err != nil {
		return nil, err
	}
	if root == "" {
		return []string{}, nil
	}
	return []string{root}, nil
}
