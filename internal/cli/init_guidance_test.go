package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitTemplate_DefaultsToUserMode(t *testing.T) {
	// quick smoke test: template itself should default to user mode
	if !strings.Contains(configTemplate, "auth:\n  mode: user") {
		t.Fatalf("expected init config template to default to user mode")
	}
}

func TestInit_PrintsNextStepsToStderr(t *testing.T) {
	root := t.TempDir()
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	oldOut, oldErr := os.Stdout, os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout, os.Stderr = wOut, wErr
	defer func() {
		os.Stdout, os.Stderr = oldOut, oldErr
	}()

	exit := Run([]string{"init"})
	_ = wOut.Close()
	_ = wErr.Close()
	_, _ = stdout.ReadFrom(rOut)
	_, _ = stderr.ReadFrom(rErr)
	if exit != 0 {
		t.Fatalf("expected exit 0, got %d", exit)
	}
	if !strings.Contains(stdout.String(), "OK") {
		t.Fatalf("expected OK on stdout, got: %s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "Next steps") {
		t.Fatalf("expected next steps on stderr, got: %s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "feishu-sync config") {
		t.Fatalf("expected config guidance on stderr, got: %s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "feishu-sync login") {
		t.Fatalf("expected login guidance on stderr, got: %s", stderr.String())
	}

	// sanity: workspace created
	if _, err := os.Stat(filepath.Join(root, ".feishu-sync", "config.yaml")); err != nil {
		t.Fatalf("expected config created: %v", err)
	}
}
