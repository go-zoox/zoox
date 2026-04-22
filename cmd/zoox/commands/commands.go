package commands

import "github.com/go-zoox/cli"

// RegisterDevTools registers install, dev, and build — same defaults as `zoox new` (cmd/server, bin/server).
func RegisterDevTools(app *cli.MultipleProgram) {
	Install(app)
	Dev(app)
	Build(app)
}
