package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-zoox/cli"
	"github.com/go-zoox/fs"
	"github.com/go-zoox/logger"
)

// Install is the install command
func Install(app *cli.MultipleProgram) {
	app.Register("install", &cli.Command{
		Name:  "install",
		Usage: "Run go mod tidy in a Zoox project (directory containing go.mod).",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "context",
				Aliases: []string{"C"},
				Usage:   "project root (containing go.mod); default: current directory",
				Value:   fs.CurrentDir(),
			},
		},
		Action: func(ctx *cli.Context) error {
			return install(ctx.String("context"))
		},
	})
}

func install(context string) error {
	abs, err := filepath.Abs(context)
	if err != nil {
		return err
	}

	logger.Infof("go mod tidy in %s", abs)
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = abs
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go mod tidy: %w", err)
	}
	logger.Infof("done")
	return nil
}
