package main

import (
	"context"
	"log"
	"os"

	"github.com/foresturquhart/grimoire/internal/grimoire"
	"github.com/urfave/cli/v3"
)

// Version is injected at build time
var version = "dev"

func main() {
	cmd := &cli.Command{
		Name:      "grimoire",
		Usage:     "convert a directory to content suitable for LLM interpretation.",
		Version:   version,
		ArgsUsage: "[target directory]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output file path.",
			},
			&cli.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Overwrite existing file without prompt.",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			targetDir := cmd.Args().First()
			if targetDir == "" {
				return cli.Exit("Error: You must specify a target directory.\n", 1)
			}

			return grimoire.Run(&grimoire.Config{
				TargetDir:  cmd.Args().First(),
				OutputPath: cmd.String("output"),
				Force:      cmd.Bool("force"),
			})
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
