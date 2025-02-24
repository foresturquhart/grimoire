package serializer

import (
	"fmt"
	"io"
)

// Serializer defines an interface for serializing multiple files into a desired format.
// Implementations should handle the specifics of formatting and output.
type Serializer interface {
	// Serialize writes the contents of the specified files, located relative to baseDir,
	// into the provided writer in a serialized format.
	// If showTree is true, it includes a directory tree visualization.
	// It returns an error if the serialization process fails.
	Serialize(writer io.Writer, baseDir string, filePaths []string, showTree bool) error
}

// SerializerFactory creates serializers based on the specified format string
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
