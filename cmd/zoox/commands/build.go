package commands

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/go-zoox/chalk"
	"github.com/go-zoox/cli"
	"github.com/go-zoox/logger"
)

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
		},
		Action: func(ctx *cli.Context) error {
			logger.Infof("start to build ...")

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

			cmd := exec.Command("sh", "-c", cmdText)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to build: %s", err.Error())
			}

			logger.Infof("build successfully, output: %s", chalk.Green(ctx.String("output")))
			return nil
		},
	})
}
