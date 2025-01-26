package serializer

import "io"

// Serializer defines an interface for serializing multiple files into a desired format.
// Implementations should handle the specifics of formatting and output.
type Serializer interface {
	// Serialize writes the contents of the specified files, located relative to baseDir,
	// into the provided writer in a serialized format.
	// It returns an error if the serialization process fails.
	Serialize(writer io.Writer, baseDir string, filePaths []string) error
}
