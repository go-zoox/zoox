package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestValidateModuleSegment(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{"ok", "user", false},
		{"ok_digit", "user2", false},
		{"empty", "", true},
		{"upper", "User", true},
		{"hyphen", "user-x", true},
		{"slash", "user/x", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateModuleSegment(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateModuleSegment(%q) err=%v wantErr=%v", tt.in, err, tt.wantErr)
			}
		})
	}
}

func TestExported_ResourcePath_ImportAlias(t *testing.T) {
	if g, w := Exported("user"), "User"; g != w {
		t.Fatalf("Exported(user)=%q want %q", g, w)
	}
	if g, w := ResourcePath("user"), "users"; g != w {
		t.Fatalf("ResourcePath(user)=%q want %q", g, w)
	}
	if g, w := ResourcePath("status"), "status"; g != w {
		t.Fatalf("ResourcePath(status)=%q want %q", g, w)
	}
	if g, w := ImportAlias("v1", "user"), "v1User"; g != w {
		t.Fatalf("ImportAlias=%q want %q", g, w)
	}
}

func TestAPIImportHookLine_APIRouteHookLine(t *testing.T) {
	imp := APIImportHookLine("example.com/app", "v1", "order")
	if want := "\tv1Order \"example.com/app/api/v1/order\""; imp != want {
		t.Fatalf("APIImportHookLine:\n got  %q\n want %q", imp, want)
	}
	route := APIRouteHookLine("v1", "order")
	if want := "\tr.Group(\"/api/v1/orders\", v1Order.Router())"; route != want {
		t.Fatalf("APIRouteHookLine:\n got  %q\n want %q", route, want)
	}
}

func TestModulePath_FindGoModDir(t *testing.T) {
	root := t.TempDir()
	mod := "example.com/scaffoldtest"
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module "+mod+"\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(root, "deep", "nested")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := FindGoModDir(sub)
	if err != nil {
		t.Fatal(err)
	}
	if got != root {
		t.Fatalf("FindGoModDir = %q want %q", got, root)
	}

	mp, err := ModulePath(root)
	if err != nil {
		t.Fatal(err)
	}
	if mp != mod {
		t.Fatalf("ModulePath = %q want %q", mp, mod)
	}
}

func TestNewProject_ErrorsWhenNotEmpty(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "x"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := NewProject(dir, NewProjectOptions{Module: "m", GoVersion: "1.22"}); err == nil {
		t.Fatal("NewProject: want error for non-empty dir")
	}
}

func TestNewProject_CreatesExpectedFiles(t *testing.T) {
	dir := t.TempDir()
	mod := "example.com/zooxscaffold"
	if err := NewProject(dir, NewProjectOptions{Module: mod, GoVersion: "1.22"}); err != nil {
		t.Fatal(err)
	}

	paths := []string{
		".zoox/config.yaml",
		"go.mod",
		"cmd/server/main.go",
		"config/config.go",
		"config/load.go",
		"migrate/migrate.go",
		"models/register.go",
		"router/rest.go",
		"middlewares/middlewares.go",
		"utils/utils.go",
	}
	for _, p := range paths {
		full := filepath.Join(dir, filepath.FromSlash(p))
		if _, err := os.Stat(full); err != nil {
			t.Errorf("missing %s: %v", p, err)
		}
	}

	goMod, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(goMod), "module "+mod) {
		t.Fatalf("go.mod missing module line")
	}
	if !strings.Contains(string(goMod), "github.com/go-zoox/zoox") {
		t.Fatalf("go.mod missing zoox require")
	}

	cfgData, err := os.ReadFile(filepath.Join(dir, ".zoox", "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	if err := yaml.Unmarshal(cfgData, &root); err != nil {
		t.Fatalf("config yaml: %v", err)
	}
	if _, ok := root["name"]; !ok {
		t.Fatal("expected name in .zoox/config.yaml")
	}
	if _, ok := root["cli"]; !ok {
		t.Fatal("expected cli in .zoox/config.yaml")
	}
	if v, _ := root["app"].(map[string]any); v == nil || v["entry"] != DefaultAppPackage {
		t.Fatalf("app.entry: %v", v)
	}

	rest, err := os.ReadFile(filepath.Join(dir, "router", "rest.go"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(rest)
	if !strings.Contains(s, apiImportsMarker) {
		t.Fatal("router/rest.go missing api imports marker")
	}
	if !strings.Contains(s, apiRoutesMarker) {
		t.Fatal("router/rest.go missing routes marker")
	}
}

func TestGenModule_CreatesPackagesAndPatchesRouter(t *testing.T) {
	dir := t.TempDir()
	mod := "example.com/gentest"
	if err := NewProject(dir, NewProjectOptions{Module: mod, GoVersion: "1.22"}); err != nil {
		t.Fatal(err)
	}

	if err := GenModule(dir, "item"); err != nil {
		t.Fatal(err)
	}

	for _, p := range []string{
		"api/v1/item/api.go",
		"api/v1/item/router.go",
		"api/v1/item/impl.go",
		"services/v1/item/service.go",
		"services/v1/item/impl.go",
		"models/v1/item/model.go",
		"models/v1/item/dto.go",
		"models/v1/item/impl.go",
	} {
		full := filepath.Join(dir, filepath.FromSlash(p))
		if _, err := os.Stat(full); err != nil {
			t.Errorf("missing %s: %v", p, err)
		}
	}

	rest, err := os.ReadFile(filepath.Join(dir, "router", "rest.go"))
	if err != nil {
		t.Fatal(err)
	}
	rs := string(rest)
	if !strings.Contains(rs, "v1Item.Router()") {
		t.Fatalf("router should register v1Item.Router(); got:\n%s", rs)
	}
	if !strings.Contains(rs, mod+"/api/v1/item") {
		t.Fatalf("router should import api package; got:\n%s", rs)
	}
	// Patch prepends an import line and keeps the tabbed marker for the next `zoox gen module`.
	if !strings.Contains(rs, apiImportsMarker) {
		t.Fatal("expected // zoox:register-api-imports to remain for next gen")
	}

	reg, err := os.ReadFile(filepath.Join(dir, "models", "register.go"))
	if err != nil {
		t.Fatal(err)
	}
	rg := string(reg)
	if !strings.Contains(rg, mod+"/models/v1/item") {
		t.Fatalf("models/register.go should blank-import model package; got:\n%s", rg)
	}
	if !strings.Contains(rg, "// zoox:register-model-imports") {
		t.Fatal("model imports marker should remain")
	}
}

func TestGenModule_RejectDuplicate(t *testing.T) {
	dir := t.TempDir()
	if err := NewProject(dir, NewProjectOptions{Module: "example.com/dup", GoVersion: "1.22"}); err != nil {
		t.Fatal(err)
	}
	if err := GenModule(dir, "x"); err != nil {
		t.Fatal(err)
	}
	if err := GenModule(dir, "x"); err == nil {
		t.Fatal("second GenModule should fail")
	}
}

func TestGenModule_PatchRouterMissingMarker(t *testing.T) {
	dir := t.TempDir()
	if err := NewProject(dir, NewProjectOptions{Module: "example.com/bad", GoVersion: "1.22"}); err != nil {
		t.Fatal(err)
	}
	routerFile := filepath.Join(dir, "router", "rest.go")
	if err := os.WriteFile(routerFile, []byte("package router\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := GenModule(dir, "y"); err == nil {
		t.Fatal("GenModule should fail when router markers missing")
	}
}

func TestRouterAlreadyRegisters(t *testing.T) {
	content := "\tr.Group(\"/api/v1/users\", v1User.Router())\n"
	if !routerAlreadyRegisters(content, "v1", "user") {
		t.Fatal("routerAlreadyRegisters should be true")
	}
	if routerAlreadyRegisters("// no routes", "v1", "user") {
		t.Fatal("routerAlreadyRegisters should be false")
	}
}

func TestModelImportAlreadyRegisters(t *testing.T) {
	content := "\t_ \"example.com/m/models/v1/user\"\n"
	if !modelImportAlreadyRegisters(content, "v1", "user") {
		t.Fatal("modelImportAlreadyRegisters should be true")
	}
}
