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

// Database registers database subcommands (e.g. migrate).
func Database(app *cli.MultipleProgram) {
	app.Register("database", &cli.Command{
		Name:        "database",
		Usage:       "Database tasks for a Zoox project (run app migrate package without starting HTTP).",
		Subcommands: []*cli.Command{migrateSubcommand()},
	})
}

func migrateSubcommand() *cli.Command {
	return &cli.Command{
		Name:  "migrate",
		Usage: "Run go mod tidy, then go run ./cmd/migrate (calls migrate.Run() in the project).",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "context",
				Aliases: []string{"C"},
				Usage:   "project root (containing go.mod); default: current directory",
				Value:   fs.CurrentDir(),
			},
		},
		Action: func(ctx *cli.Context) error {
			dir, err := filepath.Abs(ctx.String("context"))
			if err != nil {
				return err
			}
			mainGo := filepath.Join(dir, "cmd", "migrate", "main.go")
			if _, err := os.Stat(mainGo); err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("expected %s (re-run `zoox new` on a recent template, or add cmd/migrate/main.go; see `zoox database migrate --help`)", mainGo)
				}
				return err
			}
			if err := install(ctx.String("context")); err != nil {
				return err
			}
			logger.Infof("go run ./cmd/migrate (dir=%s)", dir)
			cmd := exec.Command("go", "run", "./cmd/migrate")
			cmd.Dir = dir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("go run ./cmd/migrate: %w", err)
			}
			return nil
		},
	}
}
