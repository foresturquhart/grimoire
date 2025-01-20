package grimoire

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// GenerateMarkdown generates a Markdown document from the given files and writes it to the provided writer.
// Each file's content is wrapped in triple-backtick code fences, and the file path is included as a heading.
func GenerateMarkdown(writer io.Writer, baseDir string, files []string) error {
	for i, relPath := range files {
		// Construct the absolute path
		fullPath := filepath.Join(baseDir, relPath)

		// Attempt to read the file content
		contentBytes, err := os.ReadFile(fullPath)
		if err != nil {
			slog.Warn("Failed to read file, skipping", "path", relPath, "error", err)
			continue // Skip files that cannot be read, but log the issue
		}

		// Normalise file whitespace
		normalisedContent := normaliseFileWhitespace(string(contentBytes))

		// Write the file's heading to the Markdown
		heading := fmt.Sprintf("## %s\n", relPath)
		if _, err := writer.Write([]byte(heading)); err != nil {
			return fmt.Errorf("failed to write heading for %s: %w", relPath, err)
		}

		// Wrap the file content in triple-backtick code fences
		if _, err := writer.Write([]byte("```\n")); err != nil {
			return fmt.Errorf("failed to write opening backticks for %s: %w", relPath, err)
		}

		if _, err := writer.Write([]byte(normalisedContent)); err != nil {
			return fmt.Errorf("failed to write content for %s: %w", relPath, err)
		}

		if _, err := writer.Write([]byte("\n```")); err != nil {
			return fmt.Errorf("failed to write closing backticks for %s: %w", relPath, err)
		}

		// Add a separating newline after each file (except the last one)
		if i < len(files)-1 {
			if _, err := writer.Write([]byte("\n\n")); err != nil {
				return fmt.Errorf("failed to write separating newlines: %w", err)
			}
		}
	}

	return nil
}

// normaliseFileWhitespace normalizes file whitespace for Markdown output.
// It trims leading/trailing whitespace and removes trailing spaces from each line.
func normaliseFileWhitespace(content string) string {
	content = strings.TrimSpace(content)
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t") // Remove trailing spaces/tabs
	}
	return strings.Join(lines, "\n")
}
