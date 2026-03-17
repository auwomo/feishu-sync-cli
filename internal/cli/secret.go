package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/your-org/feishu-sync/internal/config"
	"github.com/your-org/feishu-sync/internal/workspace"
)

func runSecretSet(chdir string, in io.Reader) error {
	ws, cfg, err := loadWorkspaceAndConfig(chdir, "")
	if err != nil {
		return err
	}

	secretPath, err := resolveSecretFilePath(ws, cfg)
	if err != nil {
		return err
	}

	b, err := io.ReadAll(in)
	if err != nil {
		return err
	}
	secret := strings.TrimSpace(string(b))
	if secret == "" {
		return errors.New("secret is empty; pipe/paste secret into stdin")
	}

	if err := os.MkdirAll(filepath.Dir(secretPath), 0o755); err != nil {
		return err
	}
	// 0600, overwrite
	if err := os.WriteFile(secretPath, []byte(secret+"\n"), 0o600); err != nil {
		return err
	}
	return nil
}

func runSecretShow(chdir string, reveal bool, out io.Writer) error {
	ws, cfg, err := loadWorkspaceAndConfig(chdir, "")
	if err != nil {
		return err
	}

	// env
	if cfg.App.SecretEnv != "" {
		if v := os.Getenv(cfg.App.SecretEnv); v != "" {
			fmt.Fprintf(out, "source: env %s\n", cfg.App.SecretEnv)
			if reveal {
				fmt.Fprintln(out, v)
			} else {
				fmt.Fprintln(out, "(hidden; use --reveal to print)")
			}
			return nil
		}
	}

	secretPath, err := resolveSecretFilePath(ws, cfg)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "source: file %s\n", secretPath)
	if !reveal {
		fmt.Fprintln(out, "(hidden; use --reveal to print)")
		return nil
	}

	f, err := os.Open(secretPath)
	if err != nil {
		return err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	if s.Scan() {
		fmt.Fprintln(out, strings.TrimSpace(s.Text()))
	}
	return s.Err()
}

func resolveSecretFilePath(ws *workspace.Workspace, cfg *config.Config) (string, error) {
	p := cfg.App.SecretFile
	if p == "" {
		p = filepath.Join(workspace.DirName, "secret")
	}
	if filepath.IsAbs(p) {
		return "", fmt.Errorf("app.secret_file must be relative to workspace root, got absolute: %s", p)
	}
	return filepath.Join(ws.Root, p), nil
}
