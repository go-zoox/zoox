package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
)

// NewProjectOptions configures [NewProject].
type NewProjectOptions struct {
	Module      string
	GoVersion   string
	ProjectName string // display name; if empty, derived from project directory base name
	Author      string
	Description string
	ZooxVersion string // zoox CLI version (e.g. main.Version)
}

// NewProject creates a Zoox app skeleton at projectDir with the given Go module path.
func NewProject(projectDir string, opt NewProjectOptions) error {
	module := opt.Module
	goVersion := opt.GoVersion
	if module == "" {
		return fmt.Errorf("module path is required (use --module)")
	}
	if goVersion == "" {
		return fmt.Errorf("go version is required")
	}
	abs, err := filepath.Abs(projectDir)
	if err != nil {
		return err
	}
	stepf("resolved target directory: %s", abs)

	if _, statErr := os.Stat(abs); statErr == nil {
		entries, readErr := os.ReadDir(abs)
		if readErr != nil {
			return readErr
		}
		if len(entries) > 0 {
			return fmt.Errorf("directory %s is not empty", abs)
		}
		stepf("directory exists and is empty — ok")
	} else if !os.IsNotExist(statErr) {
		return statErr
	} else {
		stepf("directory does not exist — will create")
	}

	if err := os.MkdirAll(abs, 0o755); err != nil {
		return err
	}
	stepf("ensure project root: %s", abs)

	projectName := opt.ProjectName
	if projectName == "" {
		projectName = filepath.Base(abs)
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
		{"templates/new/cmd/migrate/main.go.tmpl", "cmd/migrate/main.go"},
		{"templates/new/cmd/server/main.go.tmpl", "cmd/server/main.go"},
		{"templates/new/config/config.go.tmpl", "config/config.go"},
		{"templates/new/config/load.go.tmpl", "config/load.go"},
		{"templates/new/migrate/migrate.go.tmpl", "migrate/migrate.go"},
		{"templates/new/models/register.go.tmpl", "models/register.go"},
		{"templates/new/router/rest.go.tmpl", "router/rest.go"},
		{"templates/new/middlewares/middlewares.go.tmpl", "middlewares/middlewares.go"},
		{"templates/new/utils/utils.go.tmpl", "utils/utils.go"},
	}

	stepf("go module in go.mod: %s (go %s)", module, goVersion)
	for _, f := range files {
		data, err := renderTemplate(f.tmpl, vars)
		if err != nil {
			return fmt.Errorf("template %s: %w", f.tmpl, err)
		}
		outPath := filepath.Join(abs, filepath.FromSlash(f.out))
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(outPath, data, 0o644); err != nil {
			return err
		}
		stepf("write %s", f.out)
	}

	if err := WriteNewProjectZooxConfig(abs, projectName, opt.Author, opt.Description, opt.ZooxVersion); err != nil {
		return err
	}
	stepf("write .zoox/config.yaml")

	stepf("project scaffold finished (%d template files + .zoox/config.yaml)", len(files))
	return nil
}
