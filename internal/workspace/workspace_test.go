package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindWorkspaceWalkUp(t *testing.T) {
	root := t.TempDir()
	ws := filepath.Join(root, DirName)
	if err := os.MkdirAll(ws, 0o755); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := Find(nested)
	if err != nil {
		t.Fatal(err)
	}
	if got.Root != root {
		t.Fatalf("expected root %s, got %s", root, got.Root)
	}
}

func TestFindWorkspaceNotFound(t *testing.T) {
	root := t.TempDir()
	_, err := Find(root)
	if err == nil {
		t.Fatal("expected error")
	}
}
