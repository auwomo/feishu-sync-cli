package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunInit_WithChdir(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "proj")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := runInit(target, false, "backup"); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	cfgPath := filepath.Join(target, ".feishu-sync", "config.yaml")
	if _, err := os.Stat(cfgPath); err != nil {
		t.Fatalf("expected config at %s: %v", cfgPath, err)
	}
}
