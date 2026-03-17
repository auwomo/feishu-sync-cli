package cli

import (
	"fmt"
	"os"

	"github.com/your-org/feishu-sync/internal/config"
	"github.com/your-org/feishu-sync/internal/workspace"
)

func runValidate(chdir, configPath string) error {
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

	ws, err := workspace.Find(start)
	if err != nil {
		return err
	}

	cfgFile := configPath
	if cfgFile == "" {
		cfgFile = ws.ConfigPath()
	}

	cfg, err := config.Load(cfgFile)
	if err != nil {
		return err
	}

	var fails []string
	if err := cfg.ValidateRelativeOutputDir(); err != nil {
		fails = append(fails, err.Error())
	}
	if cfg.App.SecretEnv == "" && cfg.App.SecretFile == "" {
		fails = append(fails, "either app.secret_env or app.secret_file is required")
	} else {
		if _, err := resolveAppSecret(ws, cfg); err != nil {
			fails = append(fails, err.Error())
		}
	}
	if cfg.App.ID == "" {
		fails = append(fails, "app.id is required")
	}
	if cfg.Scope.Mode == "" {
		fails = append(fails, "scope.mode is required")
	}

	if len(fails) > 0 {
		for _, f := range fails {
			fmt.Fprintln(os.Stderr, "-", f)
		}
		return fmt.Errorf("validation failed")
	}
	return nil
}
