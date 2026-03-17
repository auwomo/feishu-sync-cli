package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/your-org/feishu-sync/internal/config"
	"github.com/your-org/feishu-sync/internal/workspace"
)

func runPull(chdir, configPath string) error {
	ws, cfg, err := loadWorkspaceAndConfig(chdir, configPath)
	if err != nil {
		return err
	}

	outAbs := filepath.Join(ws.Root, cfg.Output.Dir)

	fmt.Println("Plan:")
	fmt.Println("  workspace:", ws.Root)
	fmt.Println("  mode:     ", cfg.Scope.Mode)
	fmt.Println("  output:   ", outAbs)
	fmt.Println("  drive folders:", len(cfg.Scope.DriveFolderTokens))
	fmt.Println("  wiki spaces:  ", len(cfg.Scope.WikiSpaceIDs))
	fmt.Println("  (no network calls yet)")

	return nil
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
