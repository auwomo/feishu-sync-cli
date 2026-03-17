package discovery

import (
	"context"
	"path/filepath"

	"github.com/your-org/feishu-sync/internal/feishu"
	"github.com/your-org/feishu-sync/internal/manifest"
)

// DiscoverDriveTree recursively walks a folder token up to any depth, emitting items with Path.
func DiscoverDriveTree(ctx context.Context, client *feishu.Client, token, folderToken string) ([]manifest.DriveItem, []manifest.DiscoveryError) {
	var out []manifest.DriveItem
	var errs []manifest.DiscoveryError

	var walk func(folderTok string, prefix string)
	walk = func(folderTok string, prefix string) {
		page := ""
		for {
			resp, err := client.DriveFolderChildren(ctx, token, folderTok, page)
			if err != nil {
				errs = append(errs, manifest.DiscoveryError{Scope: "drive", Token: folderTok, Message: err.Error()})
				return
			}
			for _, f := range resp.Data.Files {
				p := filepath.ToSlash(filepath.Join(prefix, f.Name))
				out = append(out, manifest.DriveItem{Token: f.Token, Name: f.Name, Type: f.Type, Path: p})
				if f.Type == "folder" {
					walk(f.Token, p)
				}
			}
			if !resp.Data.HasMore || resp.Data.PageToken == "" {
				break
			}
			page = resp.Data.PageToken
		}
	}

	walk(folderToken, "")
	return out, errs
}
