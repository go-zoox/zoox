package scaffold

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// ErrNotZooxProject means the directory is not a Zoox project (missing or invalid .zoox/config.yaml).
var ErrNotZooxProject = errors.New("not a zoox project")

// RequireZooxProjectRoot returns nil if abs path contains a valid .zoox/config.yaml (schema version >= 1).
// Non–zoox-new layouts or plain Go modules without this file cannot use `zoox install`, `dev`, `build`, or `database` commands.
func RequireZooxProjectRoot(dir string) error {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	p := filepath.Join(abs, ".zoox", "config.yaml")
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w: %s is missing; create the project with `zoox new` or add .zoox/config.yaml from a generated project", ErrNotZooxProject, p)
		}
		return fmt.Errorf("read %s: %w", p, err)
	}
	var head struct {
		Version int `yaml:"version"`
	}
	if err := yaml.Unmarshal(b, &head); err != nil {
		return fmt.Errorf("invalid .zoox/config.yaml: %w", err)
	}
	if head.Version < 1 {
		return fmt.Errorf("unsupported .zoox/config.yaml version: %d (expected >= 1)", head.Version)
	}
	return nil
}

// zooxConfigRoot is the on-disk file written by `zoox new` and read by other zoox commands.
// Module path stays in go.mod.
type zooxConfigRoot struct {
	Version     int    `yaml:"version"`
	Name        string `yaml:"name"`
	Author      string `yaml:"author"`
	Description string `yaml:"description"`
	CreatedAt   string `yaml:"created_at"`
	CLI         string `yaml:"cli"` // zoox CLI version used to create/update the file
	App         zooxConfigApp   `yaml:"app"`
	Build       zooxConfigBuild `yaml:"build"`
	Dev         zooxConfigDev   `yaml:"dev"`
}

type zooxConfigApp struct {
	Entry string `yaml:"entry"`
}

type zooxConfigBuild struct {
	Output string `yaml:"output"`
}

type zooxConfigDev struct {
	Ignores []string `yaml:"ignores"`
}

// WriteNewProjectZooxConfig writes .zoox/config.yaml for a new project.
func WriteNewProjectZooxConfig(projectRoot string, projectName, author, description, zooxVersion string) error {
	if err := os.MkdirAll(filepath.Join(projectRoot, ".zoox"), 0o755); err != nil {
		return err
	}
	if zooxVersion == "" {
		zooxVersion = "0.0.0"
	}
	cfg := zooxConfigRoot{
		Version:     1,
		Name:        projectName,
		Author:      author,
		Description: description,
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
		CLI:         zooxVersion,
		App: zooxConfigApp{
			Entry: DefaultAppPackage,
		},
		Build: zooxConfigBuild{
			Output: DefaultOutputBinary,
		},
		Dev: zooxConfigDev{
			Ignores: nil,
		},
	}
	b, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("marshal .zoox/config.yaml: %w", err)
	}
	out := filepath.Join(projectRoot, ".zoox", "config.yaml")
	// leading newline is conventional for many YAML files; keep compact for tooling
	if err := os.WriteFile(out, b, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", out, err)
	}
	return nil
}
