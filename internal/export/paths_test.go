package export

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/your-org/feishu-sync/internal/config"
	"github.com/your-org/feishu-sync/internal/manifest"
)

func TestDriveOutPath_NoBackupDuplication(t *testing.T) {
	p := &Puller{Cfg: &config.Config{}, OutDir: filepath.FromSlash("/ws")}
	it := manifest.DriveItem{Type: "docx", Token: "tok", Name: "Doc", Path: "a/b/c"}
	out, assets := p.driveOutPath(it)

	if strings.Contains(filepath.ToSlash(out), "/backup/") {
		t.Fatalf("unexpected backup segment in out path: %s", out)
	}
	if strings.Contains(filepath.ToSlash(assets), "/backup/") {
		t.Fatalf("unexpected backup segment in assets path: %s", assets)
	}
	if !strings.HasPrefix(filepath.ToSlash(out), "/ws/drive/") {
		t.Fatalf("expected out under /ws/drive, got: %s", out)
	}
	if !strings.HasPrefix(filepath.ToSlash(assets), "/ws/drive/") {
		t.Fatalf("expected assets under /ws/drive, got: %s", assets)
	}
}
