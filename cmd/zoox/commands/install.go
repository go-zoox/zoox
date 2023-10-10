package commands

import (
	"fmt"
	"os/exec"

	"github.com/go-zoox/cli"
	"github.com/go-zoox/fs"
	"github.com/go-zoox/logger"
)

// Install is the install command
func Install(app *cli.MultipleProgram) {
	app.Register("install", &cli.Command{
		Name:  "install",
		Usage: "Install zoox application dependencies",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "context",
				Usage: "the command context",
				Value: fs.CurrentDir(),
			},
		},
		Action: func(ctx *cli.Context) error {
			return install(ctx.String("context"))
		},
	})
}

func install(context string) error {
	logger.Infof("start to install dependencies ...")

	cmd := exec.Command("go", "mod", "tidy")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install dependencies: %s", err.Error())
	}

	logger.Infof("succeed to install dependencies")

	return nil
}
