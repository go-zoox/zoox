package scaffold

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestRequireZooxProjectRoot_MissingFile(t *testing.T) {
	dir := t.TempDir()
	err := RequireZooxProjectRoot(dir)
	if err == nil {
		t.Fatal("expected error when .zoox/config.yaml is missing")
	}
	if !errors.Is(err, ErrNotZooxProject) {
		t.Fatalf("want ErrNotZooxProject: %v", err)
	}
}

func TestRequireZooxProjectRoot_OK(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".zoox"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := map[string]any{"version": 1, "name": "x"}
	b, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".zoox", "config.yaml"), b, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := RequireZooxProjectRoot(dir); err != nil {
		t.Fatal(err)
	}
}

func TestRequireZooxProjectRoot_InvalidVersion(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".zoox"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".zoox", "config.yaml"), []byte("version: 0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := RequireZooxProjectRoot(dir); err == nil {
		t.Fatal("want error for version < 1")
	}
}
