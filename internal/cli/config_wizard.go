package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/your-org/feishu-sync/internal/config"
)

func runConfigWizard(chdir, configPath string, in io.Reader, out, errOut io.Writer) error {
	ws, cfg, err := loadWorkspaceAndConfig(chdir, configPath)
	if err != nil {
		return err
	}

	// app id
	fmt.Fprint(errOut, "App ID (e.g. cli_xxx): ")
	appID, err := readLine(in)
	if err != nil {
		return err
	}
	appID = strings.TrimSpace(appID)
	if appID == "" {
		return errors.New("app id is empty")
	}

	// app secret (no-echo when possible)
	var secret string
	if isTTYFile(os.Stdin) {
		fmt.Fprint(errOut, "App Secret: ")
		b, err := readPassword(os.Stdin)
		if err != nil {
			return err
		}
		fmt.Fprintln(errOut)
		secret = strings.TrimSpace(string(b))
	} else {
		fmt.Fprintln(errOut, "App Secret: (reading from stdin; input may be echoed)")
		v, err := readLine(in)
		if err != nil {
			return err
		}
		secret = strings.TrimSpace(v)
	}
	if secret == "" {
		return errors.New("app secret is empty")
	}

	// write config.yaml (update app.id only)
	cfg.App.ID = appID
	cfgFile := configPath
	if cfgFile == "" {
		cfgFile = ws.ConfigPath()
	}
	if err := writeConfigYAML(cfgFile, cfg); err != nil {
		return err
	}

	// write secret file (default .feishu-sync/secret unless overridden)
	secretPath, err := resolveSecretFilePath(ws, cfg)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(secretPath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(secretPath, []byte(secret+"\n"), 0o600); err != nil {
		return err
	}

	fmt.Fprintln(out, "OK")
	fmt.Fprintln(errOut, "Next steps:")
	fmt.Fprintln(errOut, "  1) feishu-sync login")
	fmt.Fprintln(errOut, "  2) feishu-sync pull --dry-run   # preview")
	fmt.Fprintln(errOut, "  3) feishu-sync pull            # export")
	return nil
}

func readLine(r io.Reader) (string, error) {
	br := bufio.NewReader(r)
	line, err := br.ReadString('\n')
	if err == io.EOF {
		return strings.TrimRight(line, "\r\n"), nil
	}
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func writeConfigYAML(path string, cfg *config.Config) error {
	b, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

// Wrapper so we can unit test without reaching into term directly.
func readPassword(f *os.File) ([]byte, error) {
	return readPasswordFromFD(int(f.Fd()))
}

// implemented in password.go
func readPasswordFromFD(fd int) ([]byte, error)

// ensure imported packages are used
