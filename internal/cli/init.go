package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/your-org/feishu-sync/internal/workspace"
)

type initOptions struct {
	AppID string
}

const configTemplate = `app:
  id: cli_xxx
  secret_env: FEISHU_APP_SECRET
  secret_file: .feishu-sync/secret

auth:
  mode: user              # tenant | user

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

func runInit(chdir string, force bool, opt initOptions) error {
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

	tpl := configTemplate
	if opt.AppID != "" {
		// Replace the placeholder "cli_xxx" in the template.
		// Keep it simple and explicit: users may still edit config.yaml manually later.
		tpl = strings.Replace(tpl, "id: cli_xxx", fmt.Sprintf("id: %s", opt.AppID), 1)
	}

	ws := workspace.Workspace{Root: abs}
	return ws.Init(force, tpl)
}
