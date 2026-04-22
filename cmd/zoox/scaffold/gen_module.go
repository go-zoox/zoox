package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DefaultAPIVersion is the API segment used in paths (api/<version>/..., services/<version>/..., models/<version>/...).
const DefaultAPIVersion = "v1"

// GenModule adds api/, services/, and models/ packages for name, and patches router/rest.go and models/register.go.
func GenModule(projectRoot, name string) error {
	if err := ValidateModuleSegment(name); err != nil {
		return err
	}

	modPath, err := ModulePath(projectRoot)
	if err != nil {
		return err
	}

	apiVer := DefaultAPIVersion
	vars := ModuleVars{
		Module:       modPath,
		APIVersion:   apiVer,
		Name:         name,
		Exported:     Exported(name),
		ResourcePath: ResourcePath(name),
	}

	paths := []struct {
		tmpl string
		out  string
	}{
		{"templates/module/api.go.tmpl", filepath.Join("api", apiVer, name, "api.go")},
		{"templates/module/router.go.tmpl", filepath.Join("api", apiVer, name, "router.go")},
		{"templates/module/api_impl.go.tmpl", filepath.Join("api", apiVer, name, "impl.go")},
		{"templates/module/service.go.tmpl", filepath.Join("services", apiVer, name, "service.go")},
		{"templates/module/service_impl.go.tmpl", filepath.Join("services", apiVer, name, "impl.go")},
		{"templates/module/model.go.tmpl", filepath.Join("models", apiVer, name, "model.go")},
		{"templates/module/dto.go.tmpl", filepath.Join("models", apiVer, name, "dto.go")},
		{"templates/module/model_impl.go.tmpl", filepath.Join("models", apiVer, name, "impl.go")},
	}

	for _, p := range paths {
		outPath := filepath.Join(projectRoot, filepath.FromSlash(p.out))
		if _, err := os.Stat(outPath); err == nil {
			return fmt.Errorf("refusing to overwrite existing file %s", outPath)
		} else if !os.IsNotExist(err) {
			return err
		}
	}

	for _, p := range paths {
		data, err := renderTemplate(p.tmpl, vars)
		if err != nil {
			return err
		}
		outPath := filepath.Join(projectRoot, filepath.FromSlash(p.out))
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(outPath, data, 0o644); err != nil {
			return err
		}
	}

	routerFile := filepath.Join(projectRoot, "router", "rest.go")
	if err := patchRouter(routerFile, modPath, apiVer, name); err != nil {
		return err
	}

	modelRegFile := filepath.Join(projectRoot, "models", "register.go")
	if err := patchModelRegister(modelRegFile, modPath, apiVer, name); err != nil {
		return err
	}

	return nil
}

func patchRouter(routerFile, module, apiVer, name string) error {
	b, err := os.ReadFile(routerFile)
	if err != nil {
		return fmt.Errorf("read %s: %w (run zoox new first)", routerFile, err)
	}
	content := string(b)
	if routerAlreadyRegisters(content, apiVer, name) {
		return fmt.Errorf("router already registers module %q", name)
	}
	if !strings.Contains(content, apiImportsMarker) {
		return fmt.Errorf("%s: missing %q", routerFile, strings.TrimSpace(apiImportsMarker))
	}
	if !strings.Contains(content, apiRoutesMarker) {
		return fmt.Errorf("%s: missing %q", routerFile, strings.TrimSpace(apiRoutesMarker))
	}

	imp := APIImportHookLine(module, apiVer, name) + "\n" + apiImportsMarker
	content = strings.Replace(content, apiImportsMarker, imp, 1)

	route := APIRouteHookLine(apiVer, name) + "\n" + apiRoutesMarker
	content = strings.Replace(content, apiRoutesMarker, route, 1)

	return os.WriteFile(routerFile, []byte(content), 0o644)
}

func patchModelRegister(modelRegFile, module, apiVer, name string) error {
	b, err := os.ReadFile(modelRegFile)
	if err != nil {
		return fmt.Errorf("read %s: %w (run zoox new first)", modelRegFile, err)
	}
	content := string(b)
	if modelImportAlreadyRegisters(content, apiVer, name) {
		return fmt.Errorf("models/register.go already imports %s/models/%s/%s", module, apiVer, name)
	}
	if !strings.Contains(content, modelImportsMarker) {
		return fmt.Errorf("%s: missing %q", modelRegFile, strings.TrimSpace(modelImportsMarker))
	}

	line := ModelImportHookLine(module, apiVer, name) + "\n" + modelImportsMarker
	content = strings.Replace(content, modelImportsMarker, line, 1)
	return os.WriteFile(modelRegFile, []byte(content), 0o644)
}
