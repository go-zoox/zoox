package commands

import (
	"github.com/go-zoox/cli"
	"github.com/go-zoox/core-utils/fmt"
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
			context := ctx.String("context")
			tmpBin := fs.TmpFilePath()
			cmdText := fmt.Sprintf("go build -o %s %s && %s", tmpBin, ctx.String("entry"), tmpBin)

			logger.Debugf("Running command: %s", cmdText)

			if err := install(context); err != nil {
				return err
			}

			watcher := watch.New(&watch.Config{
				Context:  context,
				Commands: []string{cmdText},
				Ignores:  ctx.StringSlice("ignore"),
			})

			return watcher.Watch()
		},
	})
}
