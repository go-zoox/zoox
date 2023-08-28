package commands

import (
	"strings"

	"github.com/go-zoox/cli"
	"github.com/go-zoox/fs"
	"github.com/go-zoox/logger"
	"github.com/go-zoox/watch"
)

func Dev(app *cli.MultipleProgram) {
	app.Register("dev", &cli.Command{
		Name:  "dev",
		Usage: "Develop zoox application",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "entry",
				Usage:   "The entry file of the application",
				Aliases: []string{"e"},
				EnvVars: []string{"ZOOX_ENTRY"},
				Value:   ".",
			},
			&cli.StringFlag{
				Name:  "context",
				Usage: "the command context",
				Value: fs.CurrentDir(),
			},
			&cli.StringSliceFlag{
				Name:  "ignore",
				Usage: "the ignored files",
			},
		},
		Action: func(ctx *cli.Context) error {
			command := []string{
				"go run",
			}

			if ctx.String("entry") != "" {
				command = append(command, ctx.String("entry"))
			}

			cmdText := strings.Join(command, " ")
			logger.Debugf("Running command: %s", cmdText)

			watcher := watch.New(&watch.Config{
				Context:  ctx.String("context"),
				Commands: []string{cmdText},
				Ignores:  ctx.StringSlice("ignore"),
			})

			return watcher.Watch()
		},
	})
}
