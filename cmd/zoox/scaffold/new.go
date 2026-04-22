package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
)

// NewProject creates a Zoox app skeleton at projectDir with the given Go module path.
func NewProject(projectDir, module, goVersion string) error {
	if module == "" {
		return fmt.Errorf("module path is required (use --module)")
	}
	abs, err := filepath.Abs(projectDir)
	if err != nil {
		return err
	}

	if _, statErr := os.Stat(abs); statErr == nil {
		entries, readErr := os.ReadDir(abs)
		if readErr != nil {
			return readErr
		}
		if len(entries) > 0 {
			return fmt.Errorf("directory %s is not empty", abs)
		}
	} else if !os.IsNotExist(statErr) {
		return statErr
	}

	if err := os.MkdirAll(abs, 0o755); err != nil {
		return err
	}

	vars := NewProjectVars{
		Module:    module,
		GoVersion: goVersion,
	}

	files := []struct {
		tmpl string
		out  string
	}{
		{"templates/new/go.mod.tmpl", "go.mod"},
		{"templates/new/cmd/server/main.go.tmpl", "cmd/server/main.go"},
		{"templates/new/config/config.go.tmpl", "config/config.go"},
		{"templates/new/config/load.go.tmpl", "config/load.go"},
		{"templates/new/migrate/migrate.go.tmpl", "migrate/migrate.go"},
		{"templates/new/models/register.go.tmpl", "models/register.go"},
		{"templates/new/router/rest.go.tmpl", "router/rest.go"},
		{"templates/new/middlewares/middlewares.go.tmpl", "middlewares/middlewares.go"},
		{"templates/new/utils/utils.go.tmpl", "utils/utils.go"},
	}

	for _, f := range files {
		data, err := renderTemplate(f.tmpl, vars)
		if err != nil {
			return err
		}
		outPath := filepath.Join(abs, filepath.FromSlash(f.out))
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(outPath, data, 0o644); err != nil {
			return err
		}
	}

	return nil
}
