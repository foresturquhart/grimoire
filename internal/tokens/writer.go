package tokens

import (
	"bytes"
	"io"
)

// CaptureWriter is a writer that captures written content while also writing to the underlying writer.
// This allows both writing to the destination (file or stdout) while keeping a copy of the data
// for token counting.
type CaptureWriter struct {
	Writer     io.Writer     // The actual destination writer
	Buffer     *bytes.Buffer // Buffer that captures a copy of all written content
	TokenCount int           // Stores the counted tokens after processing
}

// NewCaptureWriter creates a new CaptureWriter wrapping the provided writer.
func NewCaptureWriter(w io.Writer) *CaptureWriter {
	return &CaptureWriter{
		Writer: w,
		Buffer: &bytes.Buffer{},
	}
}

// Write implements the io.Writer interface, writing data to both the underlying writer
// and the internal buffer.
func (cw *CaptureWriter) Write(p []byte) (n int, err error) {
	// Write to the buffer first (for capturing)
	n, err = cw.Buffer.Write(p)
	if err != nil {
		return n, err
	}

	// Then write to the actual destination
	return cw.Writer.Write(p)
}

// CountTokens counts the tokens in the captured content and stores the result
// in the TokenCount field.
func (cw *CaptureWriter) CountTokens() error {
	count, err := CountTokens(cw.Buffer.String())
	if err != nil {
		return err
	}

	cw.TokenCount = count
	return nil
}
