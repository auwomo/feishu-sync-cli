package cli

import (
	"path/filepath"

	"github.com/your-org/feishu-sync/internal/config"
	"github.com/your-org/feishu-sync/internal/workspace"
)

// resolveSecretFilePath resolves the secret file path from config.
// If cfg.App.SecretFile is empty, it defaults to `.feishu-sync/secret` under the workspace root.
// If cfg.App.SecretFile is relative, it is treated as relative to the workspace root.
func resolveSecretFilePath(ws *workspace.Workspace, cfg *config.Config) (string, error) {
	p := ""
	if cfg != nil {
		p = cfg.App.SecretFile
	}
	if p == "" {
		p = filepath.Join(".feishu-sync", "secret")
	}
	if filepath.IsAbs(p) {
		return p, nil
	}
	return filepath.Join(ws.Root, p), nil
}
