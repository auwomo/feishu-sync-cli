package workspace

import (
	"errors"
	"os"
	"path/filepath"
)

const DirName = ".feishu-sync"

type Workspace struct {
	Root string
}

func Find(start string) (*Workspace, error) {
	startAbs, err := filepath.Abs(start)
	if err != nil {
		return nil, err
	}
	cur := startAbs
	for {
		if exists(filepath.Join(cur, DirName)) {
			return &Workspace{Root: cur}, nil
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			break
		}
		cur = parent
	}
	return nil, errors.New("workspace not found (no .feishu-sync in parents)")
}

func (w Workspace) Path() string { return filepath.Join(w.Root, DirName) }

func (w Workspace) ConfigPath() string { return filepath.Join(w.Path(), "config.yaml") }

func (w Workspace) Init(force bool, configContent string) error {
	wsDir := w.Path()
	if exists(wsDir) {
		if !force {
			return errors.New(".feishu-sync already exists (use --force to overwrite)")
		}
		if err := os.RemoveAll(wsDir); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(filepath.Join(wsDir, "logs"), 0o755); err != nil {
		return err
	}
	return os.WriteFile(w.ConfigPath(), []byte(configContent), 0o600)
}

func exists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st != nil
}
