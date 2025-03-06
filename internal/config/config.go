package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
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

	// ShowTree indicates whether to display a directory tree at the beginning of output.
	ShowTree bool

	// DisableSort indicates whether to skip sorting files by Git commit frequency.
	DisableSort bool

	// Format specifies the output format (e.g., "md" or "xml")
	Format string

	// AllowedFileExtensions is the list of file extensions that the walker should consider.
	AllowedFileExtensions map[string]bool

	// IgnoredPathRegexes is a set of compiled regex patterns for ignoring certain paths.
	IgnoredPathRegexes []*regexp.Regexp

	// IgnoreSecrets indicates whether to proceed with output generation even if secrets are detected.
	IgnoreSecrets bool

	// RedactSecrets indicates whether to redact detected secrets in the output.
	RedactSecrets bool

	// LargeFileSizeThreshold defines the size in bytes above which a file is considered "large"
	// and a warning will be logged. Default is 1MB.
	LargeFileSizeThreshold int64

	// SkipTokenCount indicates whether to skip counting output tokens.
	SkipTokenCount bool

	// TokenCountMode specifies how tokens should be counted: "exact" or "fast".
	TokenCountMode string
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

	// Check if tree display is disabled (default is to show tree)
	showTree := !cmd.Bool("no-tree")

	// Check if sorting is disabled
	disableSort := cmd.Bool("no-sort")

	// Check if we should ignore detected secrets
	ignoreSecrets := cmd.Bool("ignore-secrets")

	// Check if we should redact detected secrets
	redactSecrets := cmd.Bool("redact-secrets")

	// Check if we should skip counting tokens
	skipTokenCount := cmd.Bool("skip-token-count")

	// Get token counting mode
	tokenCountMode := cmd.String("token-count-mode")

	// Get output format
	format := cmd.String("format")
	// Validate and normalize format
	format = strings.ToLower(format)
	switch format {
	case "md", "markdown":
		format = "md"
	case "xml":
		format = "xml"
	case "txt", "text", "plain", "plaintext":
		format = "txt"
	default:
		if format != "" {
			log.Fatal().Msgf("Unsupported format: %s", format)
		}
		// Default to markdown if no format specified
		format = "md"
	}

	// If an output file is specified, and we are not forcing an overwrite,
	// check if the file already exists.
	if outputFile != "" && !force {
		_, err := os.Stat(outputFile)
		if err == nil {
			log.Fatal().Msgf("Output file %s already exists, use --force to overwrite", outputFile)
		} else if !os.IsNotExist(err) {
			log.Fatal().Err(err).Msgf("Error checking output file %s", outputFile)
		}
	}

	// Set allowed file extensions and ignored path patterns.
	allowedFileExtensions := DefaultAllowedFileExtensions
	ignoredPathPatterns := DefaultIgnoredPathPatterns

	allowedFileExtensionsMap := make(map[string]bool)
	for _, ext := range allowedFileExtensions {
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		allowedFileExtensionsMap[ext] = true
	}

	// Compile the ignored path regexes.
	ignoredPathRegexes, err := compileRegexes(ignoredPathPatterns)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to compile ignored path pattern regexes")
	}

	// Use the default large file size threshold
	largeFileSizeThreshold := DefaultLargeFileSizeThreshold

	cfg := &Config{
		TargetDir:              targetDir,
		OutputFile:             outputFile,
		Force:                  force,
		ShowTree:               showTree,
		DisableSort:            disableSort,
		Format:                 format,
		AllowedFileExtensions:  allowedFileExtensionsMap,
		IgnoredPathRegexes:     ignoredPathRegexes,
		IgnoreSecrets:          ignoreSecrets,
		RedactSecrets:          redactSecrets,
		LargeFileSizeThreshold: largeFileSizeThreshold,
		SkipTokenCount:         skipTokenCount,
		TokenCountMode:         tokenCountMode,
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
