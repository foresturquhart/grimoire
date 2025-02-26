package main

import (
	"context"
	"os"

	"github.com/foresturquhart/grimoire/internal/config"
	"github.com/foresturquhart/grimoire/internal/core"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

// Version is injected at build time
var version = "dev"

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

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
			&cli.BoolFlag{
				Name:  "no-tree",
				Usage: "Disable directory tree display at the beginning of output.",
			},
			&cli.BoolFlag{
				Name:  "no-sort",
				Usage: "Disable sorting files by Git commit frequency.",
			},
			&cli.BoolFlag{
				Name:  "ignore-secrets",
				Usage: "Proceed with output generation even if secrets are detected.",
			},
			&cli.BoolFlag{
				Name:  "redact-secrets",
				Usage: "Redact detected secrets in output rather than failing.",
			},
			&cli.BoolFlag{
				Name:  "skip-token-count",
				Usage: "Skip counting output tokens.",
			},
			&cli.StringFlag{
				Name:  "format",
				Usage: "Output format (md, xml, or txt). Defaults to md.",
				Value: "md",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return core.Run(
				config.NewConfigFromCommand(cmd),
			)
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal().Msg(err.Error())
	}
}
