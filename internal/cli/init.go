package cli

import (	"os"
	"path/filepath"

	"github.com/your-org/feishu-sync/internal/workspace"
)

const configTemplate = `app:
  id: cli_xxx
  secret_env: FEISHU_APP_SECRET
  secret_file: .feishu-sync/secret

auth:
  mode: tenant            # tenant | user

scope:
  mode: all               # all | drive | wiki
  drive_folder_tokens: []
  wiki_space_ids: []

output:
  dir: .                  # relative to workspace root (default: workspace root)

runtime:
  concurrency: 6
  rate_limit_qps: 8
  incremental: true
`

func runInit(chdir string, force bool, out string) error {
	start := ""
	if chdir != "" {
		start = chdir
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		start = cwd
	}

	abs, err := filepath.Abs(start)
	if err != nil {
		return err
	}
	ws := workspace.Workspace{Root: abs}
	return ws.Init(force, configTemplate)
}

func ensureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

func writeFile(path, content string, perm os.FileMode) error {
	if err := ensureDir(filepath.Dir(path)); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), perm)
}
