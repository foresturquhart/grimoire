package core

import (
	"fmt"
	"github.com/foresturquhart/grimoire/internal/secrets"
	"os"
	"path/filepath"

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
		return fmt.Errorf("error walking target directory: %w", err)
	}

	log.Info().Msgf("Found %d files in %s", len(files), cfg.TargetDir)

	// Initialize variables for secret findings
	var findings []secrets.Finding
	secretsFound := false

	// Get absolute file paths for secret detection
	var absoluteFilePaths []string
	for _, file := range files {
		absPath := filepath.Join(cfg.TargetDir, file)
		absoluteFilePaths = append(absoluteFilePaths, absPath)
	}

	log.Info().Msg("Checking for secrets in files...")

	// Create a secrets detector
	detector, err := secrets.NewDetector()
	if err != nil {
		return fmt.Errorf("failed to create secrets detector: %w", err)
	}

	// Detect secrets in the files
	findings, secretsFound, err = detector.DetectSecretsInFiles(absoluteFilePaths)
	if err != nil {
		return fmt.Errorf("failed to check for secrets: %w", err)
	}

	if secretsFound {
		// Choose logging level based on how we're handling the secrets
		logFn := log.Error
		if cfg.IgnoreSecrets || cfg.RedactSecrets {
			logFn = log.Warn
		}

		for _, finding := range findings {
			logFn().
				Str("type", finding.Description).
				Str("secret", finding.Secret).
				Str("file", finding.File).
				Int("line", finding.Line).
				Msg("Detected possible secret")
		}

		if !cfg.IgnoreSecrets && !cfg.RedactSecrets {
			log.Fatal().Msg("Potential secrets detected in codebase. please review findings and remove sensitive data. use --ignore-secrets to bypass or --redact-secrets to redact (recommended)")
		}

		if cfg.IgnoreSecrets {
			log.Warn().Msg("Continuing despite detected secrets due to --ignore-secrets flag")
		}

		if cfg.RedactSecrets {
			log.Warn().Msg("Secrets will be redacted in the output due to --redact-secrets flag")
		}
	} else {
		log.Info().Msg("No secrets detected in files")
	}

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
			log.Warn().Msg("Skipped sorting files by commit frequency: git executable not found")
		}
	} else {
		log.Info().Msg("Skipped sorting files by commit frequency: sorting disabled by flag")
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

	// Prepare redaction info if needed
	var redactionInfo *serializer.RedactionInfo
	if cfg.RedactSecrets && secretsFound {
		redactionInfo = &serializer.RedactionInfo{
			Enabled:  true,
			Findings: findings,
			BaseDir:  cfg.TargetDir,
		}
	}

	// Serialize files to the configured format
	if err := formatSerializer.Serialize(writer, cfg.TargetDir, files, cfg.ShowTree, redactionInfo, cfg.LargeFileSizeThreshold); err != nil {
		return fmt.Errorf("failed to serialize content: %w", err)
	}

	// Log where we wrote results, if we wrote to a file.
	if cfg.ShouldWriteFile() {
		log.Info().Msgf("File written to %s", cfg.OutputFile)
	}

	return nil
}
