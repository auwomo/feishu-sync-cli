package discovery

import (
	"context"

	"github.com/your-org/feishu-sync/internal/feishu"
	"github.com/your-org/feishu-sync/internal/manifest"
)

func DiscoverDriveFolder(ctx context.Context, client *feishu.Client, tenantToken, folderToken string) ([]manifest.DriveItem, []manifest.DiscoveryError) {
	var items []manifest.DriveItem
	var errs []manifest.DiscoveryError

	page := ""
	for {
		resp, err := client.DriveFolderChildren(ctx, tenantToken, folderToken, page)
		if err != nil {
			errs = append(errs, manifest.DiscoveryError{Scope: "drive", Token: folderToken, Message: err.Error()})
			return items, errs
		}
		for _, f := range resp.Data.Files {
			items = append(items, manifest.DriveItem{Token: f.Token, Name: f.Name, Type: f.Type})
		}
		if !resp.Data.HasMore {
			break
		}
		page = resp.Data.PageToken
		if page == "" {
			break
		}
	}

	return items, errs
}
