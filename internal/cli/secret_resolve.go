package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/your-org/feishu-sync/internal/config"
	"github.com/your-org/feishu-sync/internal/workspace"
)

func resolveAppSecret(ws *workspace.Workspace, cfg *config.Config) (string, error) {
	if cfg.App.SecretEnv != "" {
		if v := os.Getenv(cfg.App.SecretEnv); v != "" {
			return v, nil
		}
	}

	path, err := resolveSecretFilePath(ws, cfg)
	if err != nil {
		return "", err
	}

	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("missing app secret: set env %s or run `feishu-sync secret set` to create %s", cfg.App.SecretEnv, path)
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	if !s.Scan() {
		if err := s.Err(); err != nil {
			return "", err
		}
		return "", errors.New("secret file is empty")
	}
	v := strings.TrimSpace(s.Text())
	if v == "" {
		return "", errors.New("secret file is empty")
	}
	return v, nil
}
