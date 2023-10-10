package commands

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/go-zoox/chalk"
	"github.com/go-zoox/cli"
	"github.com/go-zoox/fs"
	"github.com/go-zoox/logger"
)

// Build is the build command
func Build(app *cli.MultipleProgram) {
	app.Register("build", &cli.Command{
		Name:  "build",
		Usage: "Build zoox application",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "entry",
				Usage:   "The entry file of the application",
				Aliases: []string{"e"},
				EnvVars: []string{"ZOOX_ENTRY"},
				Value:   "main.go",
			},
			&cli.StringFlag{
				Name:    "output",
				Usage:   "The output file of the application",
				Aliases: []string{"o"},
				EnvVars: []string{"ZOOX_OUTPUT"},
				Value:   "./bin/app",
			},
			&cli.StringFlag{
				Name:  "context",
				Usage: "the command context",
				Value: fs.CurrentDir(),
			},
		},
		Action: func(ctx *cli.Context) error {
			context := ctx.String("context")
			command := []string{
				"go build",
			}

			if ctx.String("output") != "" {
				command = append(command, "-o", ctx.String("output"))
			}

			if ctx.String("entry") != "" {
				command = append(command, ctx.String("entry"))
			}

			cmdText := strings.Join(command, " ")
			logger.Debugf("Running command: %s", cmdText)

			if err := install(context); err != nil {
				return err
			}

			logger.Infof("start to build ...")
			cmd := exec.Command("sh", "-c", cmdText)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to build: %s", err.Error())
			}

			logger.Infof("succeed to build, output: %s", chalk.Green(ctx.String("output")))
			return nil
		},
	})
}
