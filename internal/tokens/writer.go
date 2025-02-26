package tokens

import (
	"bytes"
	"io"
	"sync"
)

// CaptureWriter is a writer that captures written content while also writing to the underlying writer.
// This allows both writing to the destination (file or stdout) while keeping a copy of the data
// for token counting.
type CaptureWriter struct {
	Writer     io.Writer              // The actual destination writer
	Buffer     *bytes.Buffer          // Buffer that captures a copy of all written content (used for legacy mode)
	Counter    *StreamingTokenCounter // Incremental token counter (used in streaming mode)
	TokenCount int                    // Stores the counted tokens after processing
	Streaming  bool                   // Whether to use streaming token counting
	mu         sync.Mutex             // Mutex to protect concurrent writes
	chunkSize  int                    // Size of chunks for streaming processing
}

// TokenCounterOptions configures the behavior of the CaptureWriter
type TokenCounterOptions struct {
	Streaming bool // Use streaming token counting
	ChunkSize int  // Size of chunks for streaming (default 4096)
}

// NewCaptureWriter creates a new CaptureWriter wrapping the provided writer.
func NewCaptureWriter(w io.Writer, opts *TokenCounterOptions) (*CaptureWriter, error) {
	cw := &CaptureWriter{
		Writer: w,
		Buffer: &bytes.Buffer{},
	}

	if opts != nil {
		cw.Streaming = opts.Streaming

		if opts.ChunkSize > 0 {
			cw.chunkSize = opts.ChunkSize
		} else {
			cw.chunkSize = 4096 // Default chunk size
		}

		if opts.Streaming {
			counter, err := NewStreamingCounter()
			if err != nil {
				return nil, err
			}
			cw.Counter = counter
		}
	}

	return cw, nil
}

// Write implements the io.Writer interface, writing data to both the underlying writer
// and handling token counting according to the configured mode.
func (cw *CaptureWriter) Write(p []byte) (n int, err error) {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	// If using streaming mode, count tokens as we go
	if cw.Streaming && cw.Counter != nil {
		if err := cw.Counter.AddText(string(p)); err != nil {
			return 0, err
		}
	} else if !cw.Streaming {
		// Legacy mode: store everything in buffer
		if _, err := cw.Buffer.Write(p); err != nil {
			return 0, err
		}
	}

	// Write to the actual destination
	return cw.Writer.Write(p)
}

// CountTokens counts the tokens in the captured content and stores the result
// in the TokenCount field. In streaming mode, this just returns the current count.
func (cw *CaptureWriter) CountTokens() error {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	if cw.Streaming {
		if cw.Counter != nil {
			// Streaming mode: just get the current count
			cw.TokenCount = cw.Counter.TokenCount()
			return nil
		}
		return nil
	}

	// Legacy mode: count all tokens at once
	count, err := CountTokens(cw.Buffer.String())
	if err != nil {
		return err
	}

	cw.TokenCount = count
	return nil
}

// GetTokenCount returns the current token count
func (cw *CaptureWriter) GetTokenCount() int {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	if cw.Streaming && cw.Counter != nil {
		return cw.Counter.TokenCount()
	}

	return cw.TokenCount
}
