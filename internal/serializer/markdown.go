package serializer

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// MarkdownSerializer provides methods to write multiple files' contents into a
// Markdown-formatted document. Each file is written under an H2 heading,
// and its content is placed inside a fenced code block.
type MarkdownSerializer struct{}

// NewMarkdownSerializer returns a new instance of MarkdownSerializer.
func NewMarkdownSerializer() *MarkdownSerializer {
	return &MarkdownSerializer{}
}

// Serialize takes a list of file paths relative to baseDir, reads and normalizes
// each file’s content, and writes them to writer in Markdown format. If reading
// any file fails, it logs a warning and skips that file.
func (s *MarkdownSerializer) Serialize(writer io.Writer, baseDir string, filePaths []string) error {
	for i, relPath := range filePaths {
		// Write the heading (e.g. ## path/to/file.ext)
		heading := fmt.Sprintf("## %s\n\n", relPath)
		if _, err := writer.Write([]byte(heading)); err != nil {
			return fmt.Errorf("failed to write heading for %s: %w", relPath, err)
		}

		// Read and normalize file content
		content, err := s.readAndNormalizeContent(baseDir, relPath)
		if err != nil {
			log.Warn().Err(err).Msgf("Skipping file %s due to read error", relPath)
			continue
		}

		// Wrap content in fenced code block
		formattedContent := fmt.Sprintf("```\n%s\n```", content)
		// Add an extra blank line between files, except for the last one
		if i < len(filePaths)-1 {
			formattedContent += "\n\n"
		}

		if _, err := writer.Write([]byte(formattedContent)); err != nil {
			return fmt.Errorf("failed to write content for %s: %w", relPath, err)
		}
	}

	return nil
}

// readAndNormalizeContent reads a file from baseDir/relPath and normalizes its
// content by trimming surrounding whitespace and trailing spaces on each line.
func (s *MarkdownSerializer) readAndNormalizeContent(baseDir, relPath string) (string, error) {
	fullPath := filepath.Join(baseDir, relPath)
	contentBytes, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", fullPath, err)
	}

	// Convert bytes to string and normalize
	content := string(contentBytes)
	return s.normalizeContent(content), nil
}

// normalizeContent trims surrounding whitespace and trailing spaces from each line
// of the input text, then returns the transformed string.
func (s *MarkdownSerializer) normalizeContent(content string) string {
	content = strings.TrimSpace(content)
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}

	return strings.Join(lines, "\n")
}
