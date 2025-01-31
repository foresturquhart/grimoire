package core

import (
	"fmt"
	"github.com/rs/zerolog/log"
	gitignore "github.com/sabhiram/go-gitignore"
	"os"
	"path/filepath"
	"regexp"
)

// Walker defines an interface for traversing directories and returning a list of file paths.
type Walker interface {
	Walk() ([]string, error)
}

// DefaultWalker is a concrete implementation of Walker that traverses a directory tree
// starting at targetDir, while filtering files based on allowed extensions, ignored path patterns,
// and ignore files (.gitignore and .grimoireignore). It also excludes a specific output file.
type DefaultWalker struct {
	// targetDir is the base directory from which we begin walking.
	targetDir string

	// outputFile is the absolute path of the file to exclude from the results.
	outputFile string

	// allowedFileExtensions is a set of allowed file extensions.
	allowedFileExtensions map[string]bool

	// ignoredPathRegexes is a slice of compiled regular expressions for paths that should be ignored.
	ignoredPathRegexes []*regexp.Regexp
}

// NewDefaultWalker constructs and returns a new DefaultWalker configured with the given parameters.
func NewDefaultWalker(targetDir string, allowedFileExtensions map[string]bool, ignoredPathRegexes []*regexp.Regexp, outputFile string) *DefaultWalker {
	return &DefaultWalker{
		targetDir:             targetDir,
		allowedFileExtensions: allowedFileExtensions,
		ignoredPathRegexes:    ignoredPathRegexes,
		outputFile:            outputFile,
	}
}

// Walk initiates a recursive traversal starting at targetDir.
// It returns a slice of file paths (relative to targetDir) that meet the specified filtering criteria.
func (dw *DefaultWalker) Walk() ([]string, error) {
	var files []string
	// Start traversal with no inherited ignore rules.
	if err := dw.traverse(dw.targetDir, nil, &files); err != nil {
		return nil, fmt.Errorf("directory traversal failed: %w", err)
	}
	return files, nil
}

// traverse walks the directory tree starting at the given directory.
// It accumulates ignore rules from any local .gitignore and .grimoireignore files,
// applies the allowed extension and ignored path regex filters,
// and appends any qualifying file paths (relative to targetDir) to the files slice.
func (dw *DefaultWalker) traverse(dir string, inheritedIgnores []*gitignore.GitIgnore, files *[]string) error {
	// Start with the ignore rules inherited from parent directories.
	currentIgnores := append([]*gitignore.GitIgnore{}, inheritedIgnores...)

	// Define the names of local ignore files to check.
	ignoreFilenames := []string{".gitignore", ".grimoireignore"}

	// For each ignore file, if it exists in the current directory, compile and add its rules.
	for _, ignoreFilename := range ignoreFilenames {
		ignorePath := filepath.Join(dir, ignoreFilename)
		if info, err := os.Stat(ignorePath); err == nil && !info.IsDir() {
			if gi, err := gitignore.CompileIgnoreFile(ignorePath); err == nil {
				currentIgnores = append(currentIgnores, gi)
			} else {
				log.Warn().Err(err).Msgf("Error parsing ignore file at %s", ignorePath)
			}
		}
	}

	// Read all entries (files and directories) in the current directory.
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	// Process each entry in the directory.
	for _, entry := range entries {
		// Compute the full path of the entry.
		fullPath := filepath.Join(dir, entry.Name())

		// Calculate the relative path from targetDir. This will be used for both filtering and output.
		relPath, err := filepath.Rel(dw.targetDir, fullPath)
		if err != nil {
			// If there is an error, fall back to using the full path.
			relPath = fullPath
		}
		// Normalize the relative path to use forward slashes.
		relPath = filepath.ToSlash(relPath)

		// Exclude the specific output file from being processed.
		if fullPath == dw.outputFile {
			continue
		}

		// Check if the relative path matches any of the default ignored regex patterns.
		skipByPattern := false
		for _, r := range dw.ignoredPathRegexes {
			if r.MatchString(relPath) {
				skipByPattern = true
				break
			}
		}
		if skipByPattern {
			continue
		}

		// Check the cumulative ignore rules (from .gitignore and .grimoireignore files).
		skipByGitignore := false
		for _, gi := range currentIgnores {
			if gi.MatchesPath(relPath) {
				skipByGitignore = true
				break
			}
		}
		if skipByGitignore {
			continue
		}

		// If the entry is a directory, recursively traverse it.
		if entry.IsDir() {
			if err := dw.traverse(fullPath, currentIgnores, files); err != nil {
				return err
			}
		} else {
			// For files, check whether their extension is allowed.
			ext := filepath.Ext(entry.Name())
			if dw.allowedFileExtensions[ext] {
				// Append the file's relative path to the list.
				*files = append(*files, relPath)
			}
		}
	}

	return nil
}
