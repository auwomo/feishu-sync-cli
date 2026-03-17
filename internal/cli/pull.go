package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/your-org/feishu-sync/internal/config"
	"github.com/your-org/feishu-sync/internal/discovery"
	"github.com/your-org/feishu-sync/internal/feishu"
	"github.com/your-org/feishu-sync/internal/manifest"
	"github.com/your-org/feishu-sync/internal/workspace"
)

func runPull(chdir, configPath string, dryRun bool) error {
	ws, cfg, err := loadWorkspaceAndConfig(chdir, configPath)
	if err != nil {
		return err
	}

	outAbs := filepath.Join(ws.Root, cfg.Output.Dir)

	if !dryRun {
		fmt.Println("Plan:")
		fmt.Println("  workspace:", ws.Root)
		fmt.Println("  mode:     ", cfg.Scope.Mode)
		fmt.Println("  output:   ", outAbs)
		fmt.Println("  drive folders:", len(cfg.Scope.DriveFolderTokens))
		fmt.Println("  wiki spaces:  ", len(cfg.Scope.WikiSpaceIDs))
		fmt.Println("  (no network calls yet)")
		return nil
	}

	secret := os.Getenv(cfg.App.SecretEnv)
	client := feishuNewClient()
	ctx := context.Background()
	tenantToken, err := client.TenantAccessToken(ctx, cfg.App.ID, secret)
	if err != nil {
		return err
	}

	m := manifest.PullManifest{
		WorkspaceRoot: ws.Root,
		OutputDir:     outAbs,
		Mode:          cfg.Scope.Mode,
	}
	m.Drive.Folders = map[string][]manifest.DriveItem{}

	for _, folderToken := range cfg.Scope.DriveFolderTokens {
		items, errs := discovery.DiscoverDriveFolder(ctx, client, tenantToken, folderToken)
		m.Drive.Folders[folderToken] = items
		m.Drive.Errors = append(m.Drive.Errors, errs...)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(m)
}

func loadWorkspaceAndConfig(chdir, configPath string) (*workspace.Workspace, *config.Config, error) {
	start := ""
	if chdir != "" {
		start = chdir
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, nil, err
		}
		start = cwd
	}

	ws, err := workspace.Find(start)
	if err != nil {
		return nil, nil, err
	}

	cfgFile := configPath
	if cfgFile == "" {
		cfgFile = ws.ConfigPath()
	}

	cfg, err := config.Load(cfgFile)
	if err != nil {
		return nil, nil, err
	}
	if err := cfg.ValidateRelativeOutputDir(); err != nil {
		return nil, nil, err
	}

	return ws, cfg, nil
}
