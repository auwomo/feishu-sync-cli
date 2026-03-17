package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/your-org/feishu-sync/internal/workspace"
)

const configTemplate = `app:
  id: cli_xxx
  secret_env: FEISHU_APP_SECRET

scope:
  mode: all               # all | drive | wiki
  drive_folder_tokens: []
  wiki_space_ids: []

output:
  dir: %s                 # must be relative to workspace root

runtime:
  concurrency: 6
  rate_limit_qps: 8
  incremental: true
`

func runInit(force bool, out string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	ws := workspace.Workspace{Root: cwd}
	return ws.Init(force, fmt.Sprintf(configTemplate, out))
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
