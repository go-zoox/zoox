package main

import (
	"github.com/go-zoox/cli"
	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/cmd/zoox/commands"
)

func main() {
	app := cli.NewMultipleProgram(&cli.MultipleProgramConfig{
		Name:    "zoox",
		Usage:   "zoox devtools",
		Version: zoox.Version,
	})

	commands.Dev(app)
	commands.Build(app)

	app.Run()
}
