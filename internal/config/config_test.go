package config

import "testing"

func TestValidateRelativeOutputDir(t *testing.T) {
	var c Config
	c.Output.Dir = "backup"
	if err := c.ValidateRelativeOutputDir(); err != nil {
		t.Fatalf("expected ok, got %v", err)
	}
	c.Output.Dir = "/abs/path"
	if err := c.ValidateRelativeOutputDir(); err == nil {
		t.Fatalf("expected error for abs path")
	}
	c.Output.Dir = ""
	if err := c.ValidateRelativeOutputDir(); err == nil {
		t.Fatalf("expected error for empty")
	}
}
