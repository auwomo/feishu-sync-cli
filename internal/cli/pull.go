package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/your-org/feishu-sync/internal/config"
	"github.com/your-org/feishu-sync/internal/discovery"
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

	secret, err := resolveAppSecret(ws, cfg)
	if err != nil {
		return err
	}
	ctx := context.Background()
	token, isUser, err := resolveAccessToken(ctx, ws.Path(), cfg, secret)
	if err != nil {
		return err
	}
	// isUser indicates whether token is user_access_token (vs tenant_access_token)
	_ = isUser
	client := feishuNewClient()

	m := manifest.PullManifest{
		WorkspaceRoot: ws.Root,
		OutputDir:     outAbs,
		Mode:          cfg.Scope.Mode,
	}
	m.Drive.Folders = map[string][]manifest.DriveItem{}

	// If no folder tokens are configured, attempt discovery from Drive roots.
	// Note: for tenant_access_token, Feishu Drive has no universal "root".
	// We keep tenant-only for now and guide the user to provide folder tokens.
	roots := cfg.Scope.DriveFolderTokens
	if len(roots) == 0 {
		if authMode(cfg.Auth.Mode) == "user" {
			// In user mode we can discover the current user's root folder.
			userRoots, err := discovery.DiscoverUserDriveRoots(ctx, client, token)
			if err != nil {
				m.Drive.Errors = append(m.Drive.Errors, manifest.DiscoveryError{
					Scope:   "drive",
					Token:   "",
					Message: "failed to discover user drive roots: " + err.Error(),
				})
				roots = []string{}
			} else {
				roots = userRoots
			}
		} else {
			roots = []string{"tenant"}
			m.Drive.Errors = append(m.Drive.Errors, manifest.DiscoveryError{
				Scope:   "drive",
				Token:   "",
				Message: "no drive_folder_tokens configured; tenant-only mode cannot enumerate Drive roots. Run `feishu-sync drive roots` for help, or set scope.drive_folder_tokens.",
			})
		}
	}
	m.Drive.Roots = roots

	for _, folderToken := range roots {
		// Skip placeholder roots.
		if folderToken == "tenant" {
			continue
		}
		items, errs := discovery.DiscoverDriveFolder(ctx, client, token, folderToken)
		m.Drive.Folders[folderToken] = items
		m.Drive.Errors = append(m.Drive.Errors, errs...)
		m.Drive.Summary.FolderCount++
		m.Drive.Summary.ItemCount += len(items)
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
