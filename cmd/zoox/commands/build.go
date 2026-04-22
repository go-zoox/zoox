package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-zoox/chalk"
	"github.com/go-zoox/cli"
	"github.com/go-zoox/fs"
	"github.com/go-zoox/logger"
	"github.com/go-zoox/zoox/cmd/zoox/scaffold"
)

// Build is the build command
func Build(app *cli.MultipleProgram) {
	app.Register("build", &cli.Command{
		Name:  "build",
		Usage: "Run go mod tidy, then go build in a Zoox project (default main package: " + scaffold.DefaultAppPackage + ").",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "entry",
				Usage:   "Main package to build (Zoox `new` layout: " + scaffold.DefaultAppPackage + ")",
				Aliases: []string{"e"},
				EnvVars: []string{"ZOOX_ENTRY"},
				Value:   scaffold.DefaultAppPackage,
			},
			&cli.StringFlag{
				Name:    "output",
				Usage:   "Output binary path",
				Aliases: []string{"o"},
				EnvVars: []string{"ZOOX_OUTPUT"},
				Value:   scaffold.DefaultOutputBinary,
			},
			&cli.StringFlag{
				Name:    "context",
				Aliases: []string{"C"},
				Usage:   "Zoox project root (go.mod + .zoox/config.yaml); default: current directory",
				Value:   fs.CurrentDir(),
			},
		},
		Action: func(ctx *cli.Context) error {
			dir, err := filepath.Abs(ctx.String("context"))
			if err != nil {
				return err
			}
			if err := install(ctx.String("context")); err != nil {
				return err
			}

			args := []string{"build"}
			if o := ctx.String("output"); o != "" {
				var out string
				if filepath.IsAbs(o) {
					out = filepath.Clean(o)
				} else {
					out = filepath.Clean(filepath.Join(dir, o))
				}
				args = append(args, "-o", out)
			}
			if e := ctx.String("entry"); e != "" {
				args = append(args, e)
			}

			logger.Debugf("go %v (dir=%s)", args, dir)
			logger.Infof("go build ...")
			cmd := exec.Command("go", args...)
			cmd.Dir = dir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("go build: %w", err)
			}
			out := ctx.String("output")
			if out == "" {
				out = "current directory"
			}
			logger.Infof("built %s", chalk.Green(out))
			return nil
		},
	})
}
