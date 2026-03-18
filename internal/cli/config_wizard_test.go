package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/your-org/feishu-sync/internal/workspace"
)

func TestConfigWizard_WritesAppIDAndSecret(t *testing.T) {
	root := t.TempDir()
	ws := workspace.Workspace{Root: root}
	if err := ws.Init(false, configTemplate); err != nil {
		t.Fatal(err)
	}

	in := bytes.NewBufferString("cli_test\nmy-secret\n")
	var out, errOut bytes.Buffer
	if err := runConfigWizard(root, "", in, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "OK") {
		t.Fatalf("expected OK, got %q", out.String())
	}

	cfgBytes, err := os.ReadFile(filepath.Join(root, ".feishu-sync", "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(cfgBytes), "id: cli_test") {
		t.Fatalf("expected app id updated, got:\n%s", string(cfgBytes))
	}

	secBytes, err := os.ReadFile(filepath.Join(root, ".feishu-sync", "secret"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(secBytes)) != "my-secret" {
		t.Fatalf("expected secret written, got %q", string(secBytes))
	}
}
