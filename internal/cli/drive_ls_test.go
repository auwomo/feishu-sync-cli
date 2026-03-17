package cli

import (
	"bytes"
	"testing"

	"github.com/your-org/feishu-sync/internal/manifest"
)

func TestDriveLsFormat(t *testing.T) {
	items := []manifest.DriveItem{
		{Token: "f1", Name: "Folder", Type: "folder"},
		{Token: "d1", Name: "Doc", Type: "docx"},
	}
	var buf bytes.Buffer
	driveLsFormat(items, "", &buf)
	got := buf.String()
	want := "+ Folder  (f1)\n- Doc  (d1)\n"
	if got != want {
		t.Fatalf("unexpected output:\n--- got ---\n%s--- want ---\n%s", got, want)
	}
}
