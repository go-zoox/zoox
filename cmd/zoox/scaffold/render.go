package scaffold

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
)

// NewProjectVars is passed to new-project templates.
type NewProjectVars struct {
	Module    string
	GoVersion string
}

// ModuleVars is passed to module scaffolding templates.
type ModuleVars struct {
	Module       string
	APIVersion   string
	Name         string
	Exported     string
	ResourcePath string
}

func renderTemplate(relPath string, data any) ([]byte, error) {
	b, err := files.ReadFile(filepath.ToSlash(relPath))
	if err != nil {
		return nil, fmt.Errorf("read template %s: %w", relPath, err)
	}
	tmpl, err := template.New(filepath.Base(relPath)).Parse(string(b))
	if err != nil {
		return nil, fmt.Errorf("parse template %s: %w", relPath, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute template %s: %w", relPath, err)
	}
	return buf.Bytes(), nil
}

const (
	apiImportsMarker   = "\t// zoox:register-api-imports"
	apiRoutesMarker    = "\t// zoox:register-routes"
	modelImportsMarker = "\t// zoox:register-model-imports"
)

// APIImportHookLine is inserted into router/rest.go import block.
func APIImportHookLine(module, apiVersion, name string) string {
	alias := ImportAlias(apiVersion, name)
	return fmt.Sprintf("\t%s \"%s/api/%s/%s\"", alias, module, apiVersion, name)
}

// APIRouteHookLine is inserted into router/rest.go Register body.
func APIRouteHookLine(apiVersion, name string) string {
	alias := ImportAlias(apiVersion, name)
	rp := ResourcePath(name)
	return fmt.Sprintf("\tr.Group(\"/api/%s/%s\", %s.Router())", apiVersion, rp, alias)
}

// ModelImportHookLine is inserted into models/register.go import block.
func ModelImportHookLine(module, apiVersion, name string) string {
	return fmt.Sprintf("\t_ \"%s/models/%s/%s\"", module, apiVersion, name)
}

func routerAlreadyRegisters(content, apiVersion, name string) bool {
	alias := ImportAlias(apiVersion, name)
	return strings.Contains(content, alias+".Router()")
}

func modelImportAlreadyRegisters(content, apiVersion, name string) bool {
	needle := fmt.Sprintf("/models/%s/%s\"", apiVersion, name)
	return strings.Contains(content, needle)
}
