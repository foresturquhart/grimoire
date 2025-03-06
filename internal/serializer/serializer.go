package serializer

import (
	"bufio"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/foresturquhart/grimoire/internal/secrets"
)

// RedactionInfo contains information about secrets that need to be redacted.
type RedactionInfo struct {
	// Enabled indicates whether redaction is enabled.
	Enabled bool

	// Findings contains all the secrets that were detected.
	Findings []secrets.Finding

	// BaseDir is the base directory, used to normalize paths.
	BaseDir string
}

// Serializer defines an interface for serializing multiple files into a desired format.
// Implementations should handle the specifics of formatting and output.
type Serializer interface {
	// Serialize writes the contents of the specified files, located relative to baseDir,
	// into the provided writer in a serialized format.
	// If showTree is true, it includes a directory tree visualization.
	// If redactionInfo is not nil, secrets should be redacted from the output.
	// It returns an error if the serialization process fails.
	// largeFileSizeThreshold defines the size in bytes above which a file is considered "large"
	// and a warning will be logged.
	Serialize(writer io.Writer, baseDir string, filePaths []string, showTree bool, redactionInfo *RedactionInfo, largeFileSizeThreshold int64) error
}

// NewSerializer creates serializers based on the specified format string
func NewSerializer(format string) (Serializer, error) {
	switch format {
	case "md", "markdown":
		return NewMarkdownSerializer(), nil
	case "xml":
		return NewXMLSerializer(), nil
	case "txt", "text", "plain", "plaintext":
		return NewPlainTextSerializer(), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// GetFindingsForFile returns all findings for a specific file
func GetFindingsForFile(redactionInfo *RedactionInfo, filePath string, baseDir string) []secrets.Finding {
	if redactionInfo == nil || !redactionInfo.Enabled || len(redactionInfo.Findings) == 0 {
		return nil
	}

	absPath := filepath.Join(baseDir, filePath)
	var fileFindings []secrets.Finding

	for _, finding := range redactionInfo.Findings {
		if finding.File == absPath {
			fileFindings = append(fileFindings, finding)
		}
	}

	return fileFindings
}

// RedactSecrets takes original content and redacts all secrets specified in the findings
func RedactSecrets(content string, findings []secrets.Finding) string {
	if len(findings) == 0 {
		return content
	}

	// Line-based replacement for findings that include line numbers
	lineBasedFindings := make([]secrets.Finding, 0)
	generalFindings := make([]secrets.Finding, 0)

	// Separate findings into line-based and general
	for _, finding := range findings {
		if finding.Line > 0 {
			lineBasedFindings = append(lineBasedFindings, finding)
		} else {
			generalFindings = append(generalFindings, finding)
		}
	}

	// If we have line-based findings, use line-by-line replacement
	if len(lineBasedFindings) > 0 {
		return redactByLine(content, lineBasedFindings, generalFindings)
	}

	// Sort findings by secret length in descending order (longest first)
	// This helps avoid substring replacement issues
	sort.Slice(generalFindings, func(i, j int) bool {
		return len(generalFindings[i].Secret) > len(generalFindings[j].Secret)
	})

	// Handle general replacements
	redactedContent := content
	for _, finding := range generalFindings {
		if finding.Secret == "" {
			continue
		}
		redactionNotice := "[REDACTED SECRET: " + finding.Description + "]"
		redactedContent = strings.Replace(redactedContent, finding.Secret, redactionNotice, -1)
	}

	return redactedContent
}

// redactByLine performs redaction on a line-by-line basis, which is more accurate
// when line numbers are available in the findings
func redactByLine(content string, lineBasedFindings []secrets.Finding, generalFindings []secrets.Finding) string {
	// Group findings by line number for efficient lookup
	findingsByLine := make(map[int][]secrets.Finding)
	for _, finding := range lineBasedFindings {
		findingsByLine[finding.Line] = append(findingsByLine[finding.Line], finding)
	}

	// Sort general findings by length (longest first)
	sort.Slice(generalFindings, func(i, j int) bool {
		return len(generalFindings[i].Secret) > len(generalFindings[j].Secret)
	})

	var result strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 1

	for scanner.Scan() {
		line := scanner.Text()

		// First apply line-specific redactions
		if findings, ok := findingsByLine[lineNum]; ok {
			// For each line, sort by secret length (longest first)
			sort.Slice(findings, func(i, j int) bool {
				return len(findings[i].Secret) > len(findings[j].Secret)
			})

			// Apply redactions for this line
			for _, finding := range findings {
				if finding.Secret == "" {
					continue
				}
				redactionNotice := "[REDACTED SECRET: " + finding.Description + "]"
				line = strings.Replace(line, finding.Secret, redactionNotice, -1)
			}
		}

		// Then apply general redactions that might span multiple lines
		for _, finding := range generalFindings {
			if finding.Secret == "" {
				continue
			}
			redactionNotice := "[REDACTED SECRET: " + finding.Description + "]"
			line = strings.Replace(line, finding.Secret, redactionNotice, -1)
		}

		result.WriteString(line)
		result.WriteString("\n")
		lineNum++
	}

	// Remove the trailing newline if the original content didn't have one
	resultStr := result.String()
	if !strings.HasSuffix(content, "\n") && strings.HasSuffix(resultStr, "\n") {
		resultStr = resultStr[:len(resultStr)-1]
	}

	return resultStr
}
