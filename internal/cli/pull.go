package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/your-org/feishu-sync/internal/config"
	"github.com/your-org/feishu-sync/internal/discovery"
	"github.com/your-org/feishu-sync/internal/export"
	"github.com/your-org/feishu-sync/internal/manifest"
	"github.com/your-org/feishu-sync/internal/workspace"

	"github.com/your-org/feishu-sync/internal/feishu"
)

func runPull(chdir, configPath string, dryRun bool) error {
	ws, cfg, err := loadWorkspaceAndConfig(chdir, configPath)
	if err != nil {
		return err
	}

	outAbs := filepath.Join(ws.Root, cfg.Output.Dir)

	if !dryRun {
		secret, err := resolveAppSecret(ws, cfg)
		if err != nil {
			return err
		}
		ctx := context.Background()
		token, _, err := resolveAccessToken(ctx, ws.Path(), cfg, secret)
		if err != nil {
			return err
		}
		client := feishuNewClient()

		// Build manifest (discovery) then export.
		m, err := buildPullManifest(ctx, ws, cfg, client, token)
		if err != nil {
			return err
		}

		metaDir := filepath.Join(outAbs, "_meta")
		if err := os.MkdirAll(metaDir, 0o755); err != nil {
			return err
		}
		ts := time.Now().Format("20060102")
		errorsPath := filepath.Join(metaDir, "errors-"+ts+".jsonl")
		p, err := export.NewPuller(client, token, cfg, outAbs, errorsPath)
		if err != nil {
			return err
		}
		defer p.Close()

		// Export all discovered items.
		for _, folderTok := range m.Drive.Roots {
			items := m.Drive.Folders[folderTok]
			p.ExportDriveItems(ctx, items)
		}

		manifestPath := filepath.Join(metaDir, "manifest.json")
		enc, _ := json.MarshalIndent(m, "", "  ")
		if err := os.WriteFile(manifestPath, append(enc, '\n'), 0o644); err != nil {
			return err
		}
		fmt.Fprintln(os.Stdout, "OK")
		fmt.Fprintln(os.Stdout, "manifest:", manifestPath)
		fmt.Fprintln(os.Stdout, "errors:", errorsPath)
		return nil
	}

	secret, err := resolveAppSecret(ws, cfg)
	if err != nil {
		return err
	}
	ctx := context.Background()
	token, _, err := resolveAccessToken(ctx, ws.Path(), cfg, secret)
	if err != nil {
		return err
	}
	client := feishuNewClient()

	m, err := buildPullManifest(ctx, ws, cfg, client, token)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(m)
}

func buildPullManifest(ctx context.Context, ws *workspace.Workspace, cfg *config.Config, client *feishu.Client, token string) (manifest.PullManifest, error) {
	outAbs := filepath.Join(ws.Root, cfg.Output.Dir)
	m := manifest.PullManifest{WorkspaceRoot: ws.Root, OutputDir: outAbs, Mode: cfg.Scope.Mode}
	m.Drive.Folders = map[string][]manifest.DriveItem{}

	roots := cfg.Scope.DriveFolderTokens
	if len(roots) == 0 {
		if authMode(cfg.Auth.Mode) == "user" {
			userRoots, err := discovery.DiscoverUserDriveRoots(ctx, client, token)
			if err != nil {
				m.Drive.Errors = append(m.Drive.Errors, manifest.DiscoveryError{Scope: "drive", Token: "", Message: "failed to discover user drive roots: " + err.Error()})
				roots = []string{}
			} else {
				roots = userRoots
			}
		} else {
			roots = []string{"tenant"}
			m.Drive.Errors = append(m.Drive.Errors, manifest.DiscoveryError{Scope: "drive", Token: "", Message: "no drive_folder_tokens configured; tenant-only mode cannot enumerate Drive roots. Run `feishu-sync drive roots` for help, or set scope.drive_folder_tokens."})
		}
	}
	m.Drive.Roots = roots

	for _, folderToken := range roots {
		if folderToken == "tenant" {
			continue
		}
		items, errs := discovery.DiscoverDriveTree(ctx, client, token, folderToken)
		m.Drive.Folders[folderToken] = items
		m.Drive.Errors = append(m.Drive.Errors, errs...)
		m.Drive.Summary.FolderCount++
		m.Drive.Summary.ItemCount += len(items)
	}
	return m, nil
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
