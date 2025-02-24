package core

import (
	"fmt"
	"os"

	"github.com/foresturquhart/grimoire/internal/config"
	"github.com/foresturquhart/grimoire/internal/serializer"
	"github.com/rs/zerolog/log"
)

// Run is the main entry point for processing files. It uses a Walker to retrieve
// files from cfg.TargetDir, optionally sorts them by Git commit frequency,
// and serializes them (e.g., to Markdown) via the specified Serializer.
//
// The function returns an error if any critical step (such as starting the walker
// or creating the output file) fails.
func Run(cfg *config.Config) error {
	// Create a new walker to recursively find and filter files in TargetDir.
	walker := NewDefaultWalker(cfg.TargetDir, cfg.AllowedFileExtensions, cfg.IgnoredPathRegexes, cfg.OutputFile)

	// Recursively find and filter files in TargetDir, returning a slice of string paths.
	files, err := walker.Walk()
	if err != nil {
		return fmt.Errorf("ferror walking target directory: %w", err)
	}

	log.Info().Msgf("Found %d files in %s", len(files), cfg.TargetDir)

	if !cfg.DisableSort {
		// If Git is available, attempt to sort files by commit frequency.
		gitExecutor := NewDefaultGitExecutor()
		git := NewGit(gitExecutor)

		if git.IsAvailable() {
			// If directory is within a Git repository, find the repository root
			repoDir, err := git.FindRepositoryRoot(cfg.TargetDir)
			if err != nil {
				log.Warn().Err(err).Msg("Git repository not found, skipping commit frequency file sorting")
			} else {
				log.Info().Msgf("Found Git repository at %s, sorting files by commit frequency", repoDir)

				files, err = git.SortFilesByCommitCounts(repoDir, files, git.GetCommitCounts)
				if err != nil {
					return fmt.Errorf("failed to sort files by commit frequency: %w", err)
				}
			}
		} else {
			log.Warn().Msg("skipped sorting files by commit frequency: git executable not found")
		}
	} else {
		log.Info().Msg("skipped sorting files by commit frequency: sorting disabled by flag")
	}

	// Determine where to write output. If cfg.ShouldWriteFile(), create the file, otherwise use stdout.
	var writer *os.File
	if cfg.ShouldWriteFile() {
		writer, err = os.Create(cfg.OutputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer func() {
			if cerr := writer.Close(); cerr != nil {
				log.Warn().Err(cerr).Msg("Failed to close output file")
			}
		}()
	} else {
		writer = os.Stdout
	}

	// Create a serializer based on the configured format
	formatSerializer, err := serializer.NewSerializer(cfg.Format)
	if err != nil {
		return fmt.Errorf("failed to create serializer: %w", err)
	}

	// Serialize files to the configured format
	if err := formatSerializer.Serialize(writer, cfg.TargetDir, files, cfg.ShowTree); err != nil {
		return fmt.Errorf("failed to serialize content: %w", err)
	}

	// Log where we wrote results, if we wrote to a file.
	if cfg.ShouldWriteFile() {
		log.Info().Msgf("File written to %s", cfg.OutputFile)
	}

	return nil
}
