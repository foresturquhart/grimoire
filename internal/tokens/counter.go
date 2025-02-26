package tokens

import (
	"sync"

	"github.com/tiktoken-go/tokenizer"
)

// encoderCache provides a singleton instance of the encoder to avoid repeated initialization
var (
	encoder     tokenizer.Codec
	encoderOnce sync.Once
	encoderErr  error
)

// getEncoder returns a cached encoder instance to avoid repeated initialization costs
func getEncoder() (tokenizer.Codec, error) {
	encoderOnce.Do(func() {
		encoder, encoderErr = tokenizer.Get(tokenizer.O200kBase)
	})
	return encoder, encoderErr
}

// CountTokens counts the number of tokens in the provided text using the specified encoder.
// It returns the token count and any error that occurred during counting.
func CountTokens(text string) (int, error) {
	enc, err := getEncoder()
	if err != nil {
		return 0, err
	}

	count, err := enc.Count(text)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// StreamingTokenCounter maintains an incremental token count for streaming content
type StreamingTokenCounter struct {
	enc        tokenizer.Codec
	tokenCount int
	mu         sync.Mutex
}

// NewStreamingCounter creates a new streaming token counter
func NewStreamingCounter() (*StreamingTokenCounter, error) {
	enc, err := getEncoder()
	if err != nil {
		return nil, err
	}

	return &StreamingTokenCounter{
		enc:        enc,
		tokenCount: 0,
	}, nil
}

// AddText adds text to the counter and updates the token count
func (c *StreamingTokenCounter) AddText(text string) error {
	if text == "" {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	count, err := c.enc.Count(text)
	if err != nil {
		return err
	}

	c.tokenCount += count
	return nil
}

// TokenCount returns the current token count
func (c *StreamingTokenCounter) TokenCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.tokenCount
}
