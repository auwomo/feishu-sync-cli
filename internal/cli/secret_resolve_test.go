package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/your-org/feishu-sync/internal/config"
	"github.com/your-org/feishu-sync/internal/workspace"
)

func TestResolveAppSecret_PrefersEnvOverFile(t *testing.T) {
	wsDir := t.TempDir()
	ws := &workspace.Workspace{Root: wsDir}

	cfg := &config.Config{}
	cfg.App.ID = "app"
	cfg.App.SecretEnv = "FEISHU_APP_SECRET_TEST"
	cfg.App.SecretFile = ".feishu-sync/secret"

	if err := os.MkdirAll(filepath.Join(wsDir, ".feishu-sync"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wsDir, ".feishu-sync", "secret"), []byte("file-secret\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	os.Setenv(cfg.App.SecretEnv, "env-secret")
	defer os.Unsetenv(cfg.App.SecretEnv)

	sec, err := resolveAppSecret(ws, cfg)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if sec != "env-secret" {
		t.Fatalf("expected env-secret, got %q", sec)
	}
}

func TestResolveAppSecret_FallsBackToFile(t *testing.T) {
	wsDir := t.TempDir()
	ws := &workspace.Workspace{Root: wsDir}

	cfg := &config.Config{}
	cfg.App.ID = "app"
	cfg.App.SecretEnv = "FEISHU_APP_SECRET_TEST"
	cfg.App.SecretFile = ".feishu-sync/secret"
	defer os.Unsetenv(cfg.App.SecretEnv)

	if err := os.MkdirAll(filepath.Join(wsDir, ".feishu-sync"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wsDir, ".feishu-sync", "secret"), []byte("file-secret\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	sec, err := resolveAppSecret(ws, cfg)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if sec != "file-secret" {
		t.Fatalf("expected file-secret, got %q", sec)
	}
}
