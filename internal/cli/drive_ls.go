package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/your-org/feishu-sync/internal/discovery"
	"github.com/your-org/feishu-sync/internal/feishu"
	"github.com/your-org/feishu-sync/internal/manifest"
)

type driveLsOptions struct {
	FolderToken string
	Depth       int
}

func runDriveLs(ctx context.Context, chdir, configPath string, opts driveLsOptions, out io.Writer) error {
	if opts.Depth < 0 {
		return errors.New("--depth must be >= 0")
	}

	ws, cfg, err := loadWorkspaceAndConfig(chdir, configPath)
	if err != nil {
		return err
	}

	secret, err := resolveAppSecret(ws, cfg)
	if err != nil {
		return err
	}
	client := feishuNewClient()
	tenantToken, err := client.TenantAccessToken(ctx, cfg.App.ID, secret)
	if err != nil {
		return err
	}

	folder := opts.FolderToken
	if folder == "" {
		if len(cfg.Scope.DriveFolderTokens) == 0 {
			return errors.New("no folder specified and scope.drive_folder_tokens is empty; run `feishu-sync drive roots` for help")
		}
		folder = cfg.Scope.DriveFolderTokens[0]
	}

	fmt.Fprintf(out, "drive ls folder=%s depth=%d\n", folder, opts.Depth)

	var errs []manifest.DiscoveryError
	driveLsWalk(ctx, client, tenantToken, folder, opts.Depth, "", out, &errs)
	if len(errs) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "errors:")
		for _, e := range errs {
			fmt.Fprintf(out, "  - %s (%s): %s\n", e.Scope, e.Token, e.Message)
		}
	}
	return nil
}

func driveLsFormat(items []manifest.DriveItem, indent string, out io.Writer) {
	for _, it := range items {
		marker := "-"
		if strings.EqualFold(it.Type, "folder") {
			marker = "+"
		}
		fmt.Fprintf(out, "%s%s %s  (%s)\n", indent, marker, it.Name, it.Token)
	}
}

func driveLsWalk(ctx context.Context, client *feishu.Client, tenantToken, folderToken string, depth int, indent string, out io.Writer, errs *[]manifest.DiscoveryError) {
	items, derrs := discovery.DiscoverDriveFolder(ctx, client, tenantToken, folderToken)
	*errs = append(*errs, derrs...)
	driveLsFormat(items, indent, out)
	if depth <= 0 {
		return
	}
	childIndent := indent + "  "
	for _, it := range items {
		if strings.EqualFold(it.Type, "folder") {
			driveLsWalk(ctx, client, tenantToken, it.Token, depth-1, childIndent, out, errs)
		}
	}
}
