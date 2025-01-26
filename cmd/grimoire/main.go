package main

import (
	"context"
	"github.com/foresturquhart/grimoire/internal/config"
	"github.com/foresturquhart/grimoire/internal/core"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
	"os"
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
