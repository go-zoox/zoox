package scaffold

import (
	"fmt"
	"io"
	"path/filepath"
)

// logfn is optional; when nil, scaffold steps are silent.
var logfn func(format string, args ...any)

// SetLogger sets a printf-style logger for scaffold operations. Pass nil to silence output.
func SetLogger(fn func(format string, args ...any)) {
	logfn = fn
}

func stepf(format string, args ...any) {
	if logfn != nil {
		logfn(format, args...)
	}
}

// PlannedLayoutTree returns the ASCII tree shown before `zoox new` creates files.
func PlannedLayoutTree(rootLabel string) string {
	if rootLabel == "" {
		rootLabel = "."
	}
	return fmt.Sprintf(`%s/
├── go.mod
├── .zoox/
│   └── config.yaml
├── cmd/
│   └── server/
│       └── main.go
├── config/
│   ├── config.go
│   └── load.go
├── migrate/
│   └── migrate.go
├── models/
│   └── register.go
├── router/
│   └── rest.go
├── middlewares/
│   └── middlewares.go
└── utils/
    └── utils.go
`, rootLabel)
}

// PrintPlannedLayout writes the directory layout header and tree to w.
func PrintPlannedLayout(w io.Writer, rootLabel string) {
	fmt.Fprintf(w, "Planned directory layout:\n\n")
	fmt.Fprint(w, PlannedLayoutTree(rootLabel))
	fmt.Fprint(w, "\n")
}

// PrintNewProjectNextSteps writes development, build, and production-oriented commands after a successful new.
func PrintNewProjectNextSteps(w io.Writer, projectAbsDir string) {
	rel := filepath.Clean(projectAbsDir)
	fmt.Fprint(w, "\nDone. Next commands:\n\n")
	fmt.Fprintln(w, "Development")
	fmt.Fprintf(w, "  cd %q\n", rel)
	fmt.Fprintln(w, "  zoox install")
	fmt.Fprintln(w, "  zoox dev")
	fmt.Fprintln(w, "\nBuild (one-off binary)")
	fmt.Fprintf(w, "  cd %q\n", rel)
	fmt.Fprintln(w, "  zoox build")
	fmt.Fprintln(w, "\nProduction build (example, Linux amd64, stripped)")
	fmt.Fprintf(w, "  cd %q\n", rel)
	fmt.Fprintln(w, "  GOOS=linux GOARCH=amd64 go build -trimpath -ldflags=\"-s -w\" -o server ./cmd/server")
	fmt.Fprintln(w, "\nRun binary")
	fmt.Fprintln(w, "  PORT=8080 ./server")
	fmt.Fprintln(w, "\nAdd a domain module")
	fmt.Fprintln(w, "  cd <project> && zoox gen module user")
	fmt.Fprint(w, "\n")
}

// PrintGenModuleNextSteps writes hints after a successful gen module.
func PrintGenModuleNextSteps(w io.Writer, projectAbsDir, name string) {
	rp := ResourcePath(name)
	ver := DefaultAPIVersion
	fmt.Fprint(w, "\nDone. Next commands:\n\n")
	fmt.Fprintln(w, "Rebuild / run")
	fmt.Fprintf(w, "  cd %q\n", projectAbsDir)
	fmt.Fprintln(w, "  zoox dev")
	fmt.Fprintf(w, "  # or: zoox build && %s\n", DefaultOutputBinary)
	fmt.Fprintln(w, "\nTry HTTP (default :8080)")
	fmt.Fprintf(w, "  curl -sS \"http://127.0.0.1:8080/api/%s/%s\"\n", ver, rp)
	fmt.Fprintf(w, "  curl -sS \"http://127.0.0.1:8080/api/%s/%s/123\"\n", ver, rp)
	fmt.Fprintln(w, "\nAdd another module")
	fmt.Fprintln(w, "  zoox gen module <name>")
	fmt.Fprint(w, "\n")
}
