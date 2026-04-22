package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/go-zoox/cli"
	"github.com/go-zoox/zoox/cmd/zoox/scaffold"
)

func main() {
	app := cli.NewMultipleProgram(&cli.MultipleProgramConfig{
		Name:        "zoox",
		Usage:       "Zoox scaffolding CLI for projects and domain modules.",
		Version:     Version,
		Description: "Create an api + services + models + migrate + config + middlewares + utils layout; generate domain modules under api/<ver>/name.",
	})

	if err := app.Register("new", newCommand()); err != nil {
		log.Fatal(err)
	}
	if err := app.Register("gen", genCommand()); err != nil {
		log.Fatal(err)
	}

	app.Run()
}

func attachScaffoldLogger() {
	scaffold.SetLogger(func(format string, args ...any) {
		fmt.Printf("[zoox] "+format+"\n", args...)
	})
}

func newCommand() *cli.Command {
	return &cli.Command{
		Name:      "new",
		Usage:     "Create a new project (cmd/server, api, services, models, router, migrate, config, middlewares, utils).",
		ArgsUsage: "<dir>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "module",
				Aliases: []string{"m"},
				Usage:   "Go module path (e.g. github.com/acme/api). If omitted, defaults to the base name of <dir> when it is not \".\"",
			},
			&cli.StringFlag{
				Name:  "go",
				Usage: "Go version written to go.mod (e.g. 1.22)",
				Value: "1.22",
			},
		},
		Action: func(ctx *cli.Context) error {
			if ctx.NArg() < 1 {
				return fmt.Errorf("usage: zoox new [--module path] [--go version] <dir>")
			}
			dir := ctx.Args().First()
			mod := ctx.String("module")
			absTarget, err := filepath.Abs(dir)
			if err != nil {
				return err
			}
			if mod == "" {
				base := filepath.Base(filepath.Clean(dir))
				if base == "." || base == string(filepath.Separator) || base == "" {
					return fmt.Errorf("set --module when <dir> does not yield a simple directory name")
				}
				mod = base
			}

			treeRoot := filepath.Base(absTarget)
			if treeRoot == "." || treeRoot == string(filepath.Separator) {
				treeRoot = absTarget
			}

			fmt.Println()
			scaffold.PrintPlannedLayout(os.Stdout, treeRoot)
			fmt.Println("Operations:")
			attachScaffoldLogger()
			defer scaffold.SetLogger(nil)

			if err := scaffold.NewProject(dir, mod, ctx.String("go")); err != nil {
				return err
			}
			scaffold.PrintNewProjectNextSteps(os.Stdout, absTarget)
			return nil
		},
	}
}

func genCommand() *cli.Command {
	return &cli.Command{
		Name:  "gen",
		Usage: "Generate code from embedded templates.",
		Subcommands: []*cli.Command{
			{
				Name:      "module",
				Usage:     "Add api/v1/<name>, services/v1/<name>, models/v1/<name> and patch router + models/register.",
				ArgsUsage: "<name>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "dir",
						Aliases: []string{"C"},
						Usage:   "Working directory to search upward for go.mod (default: current directory)",
					},
				},
				Action: func(ctx *cli.Context) error {
					if ctx.NArg() < 1 {
						return fmt.Errorf("usage: zoox gen module [--dir path] <name>")
					}
					name := ctx.Args().First()
					root := ctx.String("dir")
					if root == "" {
						wd, err := os.Getwd()
						if err != nil {
							return err
						}
						root = wd
					}
					proj, err := scaffold.FindGoModDir(root)
					if err != nil {
						return err
					}

					fmt.Println()
					fmt.Println("Operations:")
					attachScaffoldLogger()
					defer scaffold.SetLogger(nil)

					if err := scaffold.GenModule(proj, name); err != nil {
						return err
					}
					scaffold.PrintGenModuleNextSteps(os.Stdout, proj, name)
					return nil
				},
			},
		},
	}
}
