package cli

import (
	"context"
	"io"
)

func runDriveRoots(ctx context.Context, chdir, configPath string, out io.Writer) error {
	ws, cfg, err := loadWorkspaceAndConfig(chdir, configPath)
	if err != nil {
		return err
	}
	_ = ws
	if authMode(cfg.Auth.Mode) == "user" {
		return runDriveRootsUser(ctx, chdir, configPath, out)
	}
	return runDriveRootsTenant(out)
}
