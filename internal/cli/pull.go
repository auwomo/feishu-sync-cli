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
	"github.com/your-org/feishu-sync/internal/feishu"
	"github.com/your-org/feishu-sync/internal/manifest"
	"github.com/your-org/feishu-sync/internal/meta"
	"github.com/your-org/feishu-sync/internal/workspace"
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

		progress := newPullProgress(os.Stderr, 5*time.Second)
		defer progress.Close()
		progress.SetStage("discover")

		runID := meta.NewRunID()
		run := meta.NewRun(runID, cfg)

		fmt.Fprintf(os.Stderr, "%s mode=%s scope=%s run_id=%s out=%s\n", newTermStyle(os.Stderr).heading("[pull]"), newTermStyle(os.Stderr).bold("export"), newTermStyle(os.Stderr).bold(cfg.Scope.Mode), newTermStyle(os.Stderr).bold(runID), outAbs)

		// Build manifest (discovery) then export.
		m, err := buildPullManifest(ctx, ws, cfg, client, token)
		if err != nil {
			return err
		}
		progress.AddDriveDiscovered(m.Drive.Summary.ItemCount)
		progress.AddWikiDiscovered(m.Wiki.Summary.ItemCount)

		metaDir := filepath.Join(outAbs, "_meta")
		if err := os.MkdirAll(metaDir, 0o755); err != nil {
			return err
		}

		ledgerPath := filepath.Join(metaDir, "ledger.jsonl")
		led, err := meta.OpenLedger(ledgerPath, runID)
		if err != nil {
			return err
		}
		defer led.Close()

		// Keep old errors-*.jsonl for compatibility.
		ts := time.Now().Format("20060102")
		errorsPath := filepath.Join(metaDir, "errors-"+ts+".jsonl")
		p, err := export.NewPuller(client, token, cfg, outAbs, errorsPath, led, runID)
		if err != nil {
			return err
		}
		defer p.Close()

		progress.SetStage("export")
		// Export all discovered items.
		if cfg.Scope.Mode == "all" || cfg.Scope.Mode == "drive" {
			for _, folderTok := range m.Drive.Roots {
				items := m.Drive.Folders[folderTok]
				before := p.DriveExportedCount()
				p.ExportDriveItems(ctx, items)
				progress.AddDriveExported(p.DriveExportedCount() - before)
				progress.AddErrors(p.ErrorCount())
			}
		}

		if cfg.Scope.Mode == "all" || cfg.Scope.Mode == "wiki" {
			for _, sp := range m.Wiki.Spaces {
				items := m.Wiki.Items[sp.SpaceID]
				before := p.WikiExportedCount()
				p.ExportWikiItems(ctx, items)
				progress.AddWikiExported(p.WikiExportedCount() - before)
				progress.AddErrors(p.ErrorCount())
			}
		}

		run.End()
		mm := run.BuildManifest()
		mm.Counts.Drive.RootsDiscovered = len(m.Drive.Roots)
		mm.Counts.Drive.ItemsDiscovered = m.Drive.Summary.ItemCount
		mm.Counts.Drive.Exported = p.DriveExportedCount()
		mm.Counts.Drive.Unsupported = p.UnsupportedCount()
		mm.Counts.Drive.Errors = p.ErrorCount() + len(m.Drive.Errors)
		mm.Counts.Wiki.SpacesDiscovered = len(m.Wiki.Spaces)
		mm.Counts.Wiki.ItemsDiscovered = m.Wiki.Summary.ItemCount
		mm.Counts.Wiki.Exported = p.WikiExportedCount()
		mm.Counts.Wiki.Errors = p.ErrorCount() + len(m.Wiki.Errors)

		manifestPath := filepath.Join(metaDir, "manifest.json")
		enc, _ := json.MarshalIndent(mm, "", "  ")
		if err := os.WriteFile(manifestPath, append(enc, '\n'), 0o644); err != nil {
			return err
		}
		fmt.Fprintln(os.Stdout, "OK")
		fmt.Fprintln(os.Stdout, "manifest:", manifestPath)
		fmt.Fprintln(os.Stdout, "ledger:", ledgerPath)
		fmt.Fprintln(os.Stdout, "errors:", errorsPath)

		// Human-friendly summary (stderr), keep stdout machine-readable.
		fmt.Fprintf(os.Stderr, "Summary: drive exported=%d, wiki exported=%d, unsupported=%d, errors=%d\n", p.DriveExportedCount(), p.WikiExportedCount(), p.UnsupportedCount(), p.ErrorCount())
		fmt.Fprintln(os.Stderr, "manifest:", manifestPath)
		fmt.Fprintln(os.Stderr, "ledger:", ledgerPath)
		fmt.Fprintln(os.Stderr, "errors:", errorsPath)
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
	m.Wiki.Items = map[string][]manifest.WikiItem{}

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

	if cfg.Scope.Mode == "all" || cfg.Scope.Mode == "wiki" {
		spaces, err := discovery.DiscoverWikiSpaces(ctx, client, token)
		if err != nil {
			m.Wiki.Errors = append(m.Wiki.Errors, manifest.DiscoveryError{Scope: "wiki", Token: "", Message: "failed to list spaces: " + err.Error()})
			spaces = []feishu.WikiSpace{}
		}
		allowed := map[string]bool{}
		if len(cfg.Scope.WikiSpaceIDs) > 0 {
			for _, id := range cfg.Scope.WikiSpaceIDs {
				allowed[id] = true
			}
		}
		for _, sp := range spaces {
			if len(allowed) > 0 && !allowed[sp.SpaceID] {
				continue
			}
			m.Wiki.Spaces = append(m.Wiki.Spaces, struct {
				SpaceID string `json:"space_id"`
				Name    string `json:"name"`
			}{SpaceID: sp.SpaceID, Name: sp.Name})
			wsItems, wsErrs := discovery.DiscoverWikiTree(ctx, client, token, sp.SpaceID)
			for i := range wsItems {
				wsItems[i].SpaceName = sp.Name
			}
			m.Wiki.Items[sp.SpaceID] = wsItems
			m.Wiki.Errors = append(m.Wiki.Errors, wsErrs...)
			m.Wiki.Summary.SpaceCount++
			m.Wiki.Summary.ItemCount += len(wsItems)
			m.Wiki.Summary.NodeCount += len(wsItems)
		}
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
