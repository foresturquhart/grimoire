package grimoire

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/boyter/gocodewalker"
)

// GetFiles retrieves all files in the given directory and its subdirectories,
// filtering by allowed extensions and ignoring patterns defined in IgnoredPatterns.
func GetFiles(dir string) ([]string, error) {
	fileListQueue := make(chan *gocodewalker.File, 100)
	walker := gocodewalker.NewFileWalker(dir, fileListQueue)

	// Set allowed file extensions
	walker.AllowListExtensions = append(walker.AllowListExtensions, AllowedExtensions...)

	// Handle errors during file walking
	walker.SetErrorHandler(func(err error) bool {
		slog.Warn("Error encountered during file walk", "error", err)
		return true // Continue walking despite the error
	})

	// Start the file walker in a goroutine
	go walker.Start()

	// Compile ignored patterns into regexes
	ignoredRegexes, err := compileIgnoredPatterns()
	if err != nil {
		return nil, fmt.Errorf("failed to compile ignored patterns: %w", err)
	}

	// Collect files from the walker
	var files []string
	for fileEntry := range fileListQueue {
		relPath := strings.TrimPrefix(fileEntry.Location, dir)
		relPath = strings.TrimPrefix(relPath, string(filepath.Separator))

		if shouldInclude(relPath, ignoredRegexes) {
			files = append(files, relPath)
		}
	}

	return files, nil
}

// compileIgnoredPatterns compiles the IgnoredPatterns into regex objects.
func compileIgnoredPatterns() ([]*regexp.Regexp, error) {
	var ignoredRegexes []*regexp.Regexp

	for _, pattern := range IgnoredPatterns {
		reg, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid ignore pattern %q: %w", pattern, err)
		}
		ignoredRegexes = append(ignoredRegexes, reg)
	}

	return ignoredRegexes, nil
}

// shouldInclude determines whether a file path should be included based on the ignored patterns.
func shouldInclude(path string, ignoreRegexes []*regexp.Regexp) bool {
	for _, re := range ignoreRegexes {
		if re.MatchString(path) {
			return false
		}
	}
	return true
}

// CreateOutputWriter initializes a writer for the given output path.
// If the path is empty, the function writes to stdout.
func CreateOutputWriter(outputPath string) (*os.File, func() error, error) {
	if outputPath == "" {
		// Use stdout as the writer, with a no-op closer
		return os.Stdout, func() error { return nil }, nil
	}

	// Attempt to create the file for writing
	file, err := os.Create(outputPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create output file: %w", err)
	}

	// Provide a closer that ensures the file is properly closed
	return file, func() error { return file.Close() }, nil
}
