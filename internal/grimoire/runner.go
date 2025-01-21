package grimoire

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

type Config struct {
	TargetDir  string // Directory to scan
	OutputPath string // Path to the output file
	Force      bool   // Whether to overwrite existing files without prompting
}

// Run executes the main logic for the grimoire tool based on the provided configuration.
func Run(cfg *Config) error {
	// Resolve the absolute path of the target directory
	absDir, err := filepath.Abs(cfg.TargetDir)
	if err != nil {
		return fmt.Errorf("failed to resolve target directory: %w", err)
	}

	// Ensure the output file does not already exist unless --force is set
	if cfg.OutputPath != "" && !cfg.Force {
		if _, err := os.Stat(cfg.OutputPath); err == nil {
			return fmt.Errorf("output file %s already exists. Use --force to overwrite", cfg.OutputPath)
		}
	}

	// Retrieve the list of files to process
	files, err := GetFiles(absDir)
	if err != nil {
		return fmt.Errorf("failed to retrieve files: %w", err)
	}

	// If git is available, we'll attempt to sort by commit frequency
	if IsGitAvailable() {
		// Attempt to find the Git repo for sorting files by importance
		gitRoot, err := FindGitRoot(absDir)
		if err != nil {
			// If we can't find a repo, we simply skip commit-based sorting
			slog.Warn("No repository found, skipping commit-based sorting", "error", err)
		} else {
			slog.Info("Found repository, sorting files by commit frequency", "path", gitRoot)

			// sort files by change frequency using Git commit data
			files, err = SortByChangeFrequency(gitRoot, files, DefaultChangeCounter)
			if err != nil {
				return fmt.Errorf("failed to sort files by commit frequency: %w", err)
			}
		}
	} else {
		slog.Warn("Git not found, skipping commit-based sorting")
	}

	// Create an output writer for the Markdown output
	writer, closer, err := CreateOutputWriter(cfg.OutputPath)
	if err != nil {
		return fmt.Errorf("failed to create output writer: %w", err)
	}
	defer closer()

	// Generate Markdown content
	if err := GenerateMarkdown(writer, absDir, files); err != nil {
		return fmt.Errorf("failed to generate Markdown: %w", err)
	}

	// Log a success message if writing to a file
	if cfg.OutputPath != "" {
		slog.Info("Markdown file successfully written", "path", cfg.OutputPath)
	}

	return nil
}
