package commands

import (
	"path/filepath"

	"github.com/go-zoox/cli"
	"github.com/go-zoox/core-utils/fmt"
	"github.com/go-zoox/fs"
	"github.com/go-zoox/logger"
	"github.com/go-zoox/watch"
	"github.com/go-zoox/zoox/cmd/zoox/scaffold"
)

// Dev is the dev command
func Dev(app *cli.MultipleProgram) {
	app.Register("dev", &cli.Command{
		Name:  "dev",
		Usage: "Run go mod tidy, then watch and rebuild the main package on file changes (default: " + scaffold.DefaultAppPackage + ").",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "entry",
				Usage:   "Main package to build and run (Zoox `new` layout: " + scaffold.DefaultAppPackage + ")",
				Aliases: []string{"e"},
				EnvVars: []string{"ZOOX_ENTRY"},
				Value:   scaffold.DefaultAppPackage,
			},
			&cli.StringFlag{
				Name:    "context",
				Aliases: []string{"C"},
				Usage:   "project root (containing go.mod); default: current directory",
				Value:   fs.CurrentDir(),
			},
			&cli.StringSliceFlag{
				Name:  "ignore",
				Usage: "glob paths to ignore for reload (passed to the file watcher)",
			},
		},
		Action: func(ctx *cli.Context) error {
			root, err := filepath.Abs(ctx.String("context"))
			if err != nil {
				return err
			}
			tmpBin := fs.TmpFilePath()
			entry := ctx.String("entry")
			cmdText := fmt.Sprintf("go build -o %q %q && %q", tmpBin, entry, tmpBin)

			logger.Debugf("Running command: %s", cmdText)

			if err := install(ctx.String("context")); err != nil {
				return err
			}

			watcher := watch.New(&watch.Config{
				Context:  root,
				Commands: []string{cmdText},
				Ignores:  ctx.StringSlice("ignore"),
			})

			return watcher.Watch()
		},
	})
}
