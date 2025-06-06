package serializer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/foresturquhart/grimoire/internal/tokens"
	"github.com/rs/zerolog/log"
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
// each file's content, and writes them to writer in Markdown format. If reading
// any file fails, it logs a warning and skips that file.
// If showTree is true, it prepends a directory tree visualization.
// If redactionInfo is not nil, it redacts secrets from the output.
// largeFileSizeThreshold defines the size in bytes above which a file is considered "large"
// and a warning will be logged.
// highTokenThreshold defines the token count above which a file is considered
// to have a high token count and a warning will be logged.
// skipTokenCount indicates whether to skip token counting entirely for warnings.
func (s *MarkdownSerializer) Serialize(writer io.Writer, baseDir string, filePaths []string, showTree bool, redactionInfo *RedactionInfo, largeFileSizeThreshold int64, highTokenThreshold int, skipTokenCount bool) error {
	// Write the header with timestamp
	timestamp := time.Now().UTC().Format(time.RFC3339Nano)
	header := fmt.Sprintf("This document contains a structured representation of the entire codebase, merging all files into a single Markdown file.\n\nGenerated by Grimoire on: %s\n\n", timestamp)

	if _, err := writer.Write([]byte(header)); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write the summary section
	summary := "## Summary\n\n"
	summary += "This file contains a packed representation of the entire codebase's contents. "
	summary += "It is designed to be easily consumable by AI systems for analysis, code review, or other automated processes.\n\n"
	summary += "- This file should be treated as read-only. Any changes should be made to the original codebase files.\n"
	summary += "- When processing this file, use the file path headings to distinguish between different files.\n"
	summary += "- This file may contain sensitive information and should be handled with appropriate care.\n"

	if redactionInfo != nil && redactionInfo.Enabled {
		summary += "- Detected secrets have been redacted with the format [REDACTED SECRET: description].\n"
	}

	summary += "- Some files may have been excluded based on .gitignore rules and Grimoire's configuration.\n"

	if showTree {
		summary += "- The file begins with this summary, followed by the directory structure, and then includes all codebase files.\n\n"
	} else {
		summary += "- The file begins with this summary, followed by all codebase files.\n\n"
	}

	if _, err := writer.Write([]byte(summary)); err != nil {
		return fmt.Errorf("failed to write summary: %w", err)
	}

	// Add directory tree if requested
	if showTree && len(filePaths) > 0 {
		treeGen := NewDefaultTreeGenerator()
		rootNode := treeGen.GenerateTree(filePaths)

		treeContent := "## Directory Structure\n\n"
		treeContent += s.renderTreeAsMarkdownList(rootNode, 0)
		treeContent += "\n"

		if _, err := writer.Write([]byte(treeContent)); err != nil {
			return fmt.Errorf("failed to write directory tree: %w", err)
		}
	}

	// Write files heading
	filesHeading := "## Files\n\n"
	if _, err := writer.Write([]byte(filesHeading)); err != nil {
		return fmt.Errorf("failed to write files heading: %w", err)
	}

	// Process each file
	for i, relPath := range filePaths {
		// Write the heading (e.g. ## path/to/file.ext)
		heading := fmt.Sprintf("### File: %s\n\n", relPath)
		if _, err := writer.Write([]byte(heading)); err != nil {
			return fmt.Errorf("failed to write heading for %s: %w", relPath, err)
		}

		// Read and normalize file content
		content, isLargeFile, err := s.readAndNormalizeContent(baseDir, relPath, redactionInfo, largeFileSizeThreshold, highTokenThreshold, skipTokenCount)
		if err != nil {
			log.Warn().Err(err).Msgf("Skipping file %s due to read error", relPath)
			continue
		}

		if isLargeFile {
			log.Warn().Msgf("File %s exceeds the large file threshold (%d bytes). Including in output but this may impact performance.", relPath, largeFileSizeThreshold)
		}

		// Check if the file is minified (only applicable file type)
		if IsMinifiedFile(content, relPath, DefaultMinifiedFileThresholds) {
			log.Warn().Msgf("File %s appears to be minified. Consider excluding it to reduce token counts.", relPath)
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

// renderTreeAsMarkdownList recursively builds a nested Markdown list representation of the tree.
// This is specific to the Markdown serializer's formatting needs.
func (s *MarkdownSerializer) renderTreeAsMarkdownList(node *TreeNode, depth int) string {
	if node == nil {
		return ""
	}

	var builder strings.Builder

	// Skip the root node in output
	if node.Name != "" {
		// Create indentation based on depth
		indent := strings.Repeat("  ", depth)

		// Add list item marker and name
		builder.WriteString(indent)
		builder.WriteString("- ")
		builder.WriteString(node.Name)

		// Add a directory indicator for directories
		if node.IsDir {
			builder.WriteString("/")
		}

		builder.WriteString("\n")
	}

	// Process children with incremented depth
	for _, child := range node.Children {
		childDepth := depth
		if node.Name != "" {
			// Only increment depth for non-root nodes
			childDepth++
		}
		builder.WriteString(s.renderTreeAsMarkdownList(child, childDepth))
	}

	return builder.String()
}

// readAndNormalizeContent reads a file from baseDir/relPath and normalizes its
// content by trimming surrounding whitespace and trailing spaces on each line.
// If redactionInfo is not nil, it redacts secrets from the output.
// It also checks if the file exceeds the large file size threshold and returns a flag if it does.
func (s *MarkdownSerializer) readAndNormalizeContent(baseDir, relPath string, redactionInfo *RedactionInfo, largeFileSizeThreshold int64, highTokenThreshold int, skipTokenCount bool) (string, bool, error) {
	fullPath := filepath.Join(baseDir, relPath)

	// Check file size before reading
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		return "", false, fmt.Errorf("failed to stat file %s: %w", fullPath, err)
	}

	// Check if file exceeds large file threshold
	isLargeFile := fileInfo.Size() > largeFileSizeThreshold

	contentBytes, err := os.ReadFile(fullPath)
	if err != nil {
		return "", false, fmt.Errorf("failed to read file %s: %w", fullPath, err)
	}

	// Convert bytes to string and normalize
	content := string(contentBytes)
	normalizedContent := s.normalizeContent(content)

	// If redaction is enabled, redact any secrets
	if redactionInfo != nil && redactionInfo.Enabled {
		fileFindings := GetFindingsForFile(redactionInfo, relPath, baseDir)
		if len(fileFindings) > 0 {
			normalizedContent = RedactSecrets(normalizedContent, fileFindings)
		}
	}

	// Count tokens for this file and warn if it exceeds the threshold
	if !skipTokenCount && highTokenThreshold > 0 {
		tokenCount, err := tokens.CountFileTokens(relPath, normalizedContent)
		if err != nil {
			log.Warn().Err(err).Msgf("Failed to count tokens for file %s", relPath)
		} else if tokenCount > highTokenThreshold {
			log.Warn().Msgf("File %s has a high token count (%d tokens, threshold: %d). This will consume significant LLM context.", relPath, tokenCount, highTokenThreshold)
		}
	}

	return normalizedContent, isLargeFile, nil
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
