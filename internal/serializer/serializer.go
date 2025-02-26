package serializer

import (
	"fmt"
	"github.com/foresturquhart/grimoire/internal/secrets"
	"io"
	"path/filepath"
	"strings"
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

	// For each finding, replace the secret with a redaction notice
	redactedContent := content
	for _, finding := range findings {
		redactionNotice := "[REDACTED SECRET: " + finding.Description + "]"
		redactedContent = strings.Replace(redactedContent, finding.Secret, redactionNotice, -1)
	}

	return redactedContent
}
