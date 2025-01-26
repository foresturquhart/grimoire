package config

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
	"os"
	"path/filepath"
	"regexp"
)

// Config holds all configuration data needed to run the application.
// This includes the target directory to walk, the path (if any) to write output,
// and various options for filtering or overriding existing files.
type Config struct {
	// TargetDir is the directory from which files will be walked.
	TargetDir string

	// OutputFile is the file where results may be written. If empty,
	// output is directed to stdout.
	OutputFile string

	// Force indicates whether existing output files should be overwritten.
	Force bool

	// AllowedFileExtensions is the list of file extensions that the walker should consider.
	AllowedFileExtensions []string

	// IgnoredPathRegexes is a set of compiled regex patterns for ignoring certain paths.
	IgnoredPathRegexes []*regexp.Regexp
}

// NewConfigFromCommand constructs a Config by extracting relevant values from
// the provided cli.Command.
func NewConfigFromCommand(cmd *cli.Command) *Config {
	var err error

	// Extract the target directory from the command arguments.
	targetDir := cmd.Args().First()
	if targetDir == "" {
		log.Fatal().Msg("You must specify a target directory")
	}

	// Convert target directory to an absolute path.
	targetDir, err = filepath.Abs(targetDir)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to resolve target directory %s", targetDir)
	}

	// Convert output file to an absolute path.
	outputFile := cmd.String("output")
	if outputFile != "" {
		outputFile, err = filepath.Abs(outputFile)
		if err != nil {
			log.Fatal().Err(err).Msgf("Failed to resolve output file %s", outputFile)
		}
	}

	// Fetch value of force command line flag.
	force := cmd.Bool("force")

	// If an output file is specified, and we are not forcing an overwrite,
	// check if the file already exists.
	if outputFile != "" && !force {
		if _, err := os.Stat(outputFile); err == nil {
			log.Fatal().Msgf("Output file %s already exists, use --force to overwrite", outputFile)
		}
	}

	// Set allowed file extensions and ignored path patterns.
	allowedFileExtensions := DefaultAllowedFileExtensions
	ignoredPathPatterns := DefaultIgnoredPathPatterns

	// Compile the ignored path regexes.
	ignoredPathRegexes, err := compileRegexes(ignoredPathPatterns)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to compile ignored path pattern regexes")
	}

	cfg := &Config{
		TargetDir:             targetDir,
		OutputFile:            outputFile,
		Force:                 force,
		AllowedFileExtensions: allowedFileExtensions,
		IgnoredPathRegexes:    ignoredPathRegexes,
	}

	return cfg
}

// ShouldWriteFile returns true if the configuration is set to write output
// to a file (i.e., if OutputFile is non-empty).
func (cfg *Config) ShouldWriteFile() bool {
	return cfg.OutputFile != ""
}

// compileRegexes takes a slice of regex pattern strings and compiles them into
// a slice of *regexp.Regexp. If any pattern is invalid, an error is returned.
func compileRegexes(regexes []string) ([]*regexp.Regexp, error) {
	var compiled []*regexp.Regexp
	for _, pattern := range regexes {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern %q: %w", pattern, err)
		}
		compiled = append(compiled, re)
	}
	return compiled, nil
}
