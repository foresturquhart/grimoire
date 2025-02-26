package tokens

import (
	"github.com/tiktoken-go/tokenizer"
)

// CountTokens counts the number of tokens in the provided text using the specified encoder.
// It returns the token count and any error that occurred during counting.
func CountTokens(text string) (int, error) {
	enc, err := tokenizer.Get(tokenizer.O200kBase)
	if err != nil {
		return 0, err
	}

	ids, _, err := enc.Encode(text)
	if err != nil {
		return 0, err
	}

	return len(ids), nil
}
