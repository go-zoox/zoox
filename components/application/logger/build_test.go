package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-zoox/zoox/config"
)

func TestBuild_fileLevels(t *testing.T) {
	dir := t.TempDir()
	accessPath := filepath.Join(dir, "access.log")
	errPath := filepath.Join(dir, "error.log")

	l, err := Build(&config.Logger{
		Level: "DEBUG",
		Transports: []config.Transport{{
			Type: "file",
			Path: accessPath,
			Levels: map[string]string{
				"error": errPath,
			},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}

	l.Info("info-msg")
	l.Error("err-msg")

	ab, _ := os.ReadFile(accessPath)
	if !strings.Contains(string(ab), "info-msg") {
		t.Fatalf("access/default: %q", ab)
	}
	eb, _ := os.ReadFile(errPath)
	if !strings.Contains(string(eb), "err-msg") {
		t.Fatalf("error: %q", eb)
	}
}

func TestBuild_consoleOnly(t *testing.T) {
	l, err := Build(&config.Logger{Level: "INFO"})
	if err != nil {
		t.Fatal(err)
	}
	l.Info("console-only")
}

func TestBuild_emptyTransportsIsConsoleOnly(t *testing.T) {
	l, err := Build(&config.Logger{})
	if err != nil {
		t.Fatal(err)
	}
	if l == nil {
		t.Fatal("expected logger")
	}
}
