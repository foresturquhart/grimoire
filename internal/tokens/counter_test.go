package tokens

import (
	"bytes"
	"testing"
)

func TestCountTokensAndStreaming(t *testing.T) {
	part1 := "hello "
	part2 := "world"
	total := part1 + part2

	expected, err := CountTokens(total)
	if err != nil {
		t.Fatalf("CountTokens returned error: %v", err)
	}
	if expected == 0 {
		t.Fatalf("expected token count > 0")
	}

	sc, err := NewStreamingCounter()
	if err != nil {
		t.Fatalf("NewStreamingCounter returned error: %v", err)
	}

	if err := sc.AddText(part1); err != nil {
		t.Fatalf("AddText returned error: %v", err)
	}
	if err := sc.AddText(part2); err != nil {
		t.Fatalf("AddText returned error: %v", err)
	}

	if got := sc.TokenCount(); got != expected {
		t.Errorf("streaming token count = %d, want %d", got, expected)
	}
}

func TestCaptureWriterCounts(t *testing.T) {
	var dest bytes.Buffer
	cw, err := NewCaptureWriter(&dest, nil)
	if err != nil {
		t.Fatalf("NewCaptureWriter returned error: %v", err)
	}

	input := "foo bar"
	if _, err := cw.Write([]byte(input)); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	if err := cw.CountTokens(); err != nil {
		t.Fatalf("CountTokens returned error: %v", err)
	}

	expected, err := CountTokens(input)
	if err != nil {
		t.Fatalf("CountTokens returned error: %v", err)
	}

	if cw.GetTokenCount() != expected {
		t.Errorf("token count = %d, want %d", cw.GetTokenCount(), expected)
	}

	if dest.String() != input {
		t.Errorf("writer content = %q, want %q", dest.String(), input)
	}
}

func TestCountTokensEmpty(t *testing.T) {
	count, err := CountTokens("")
	if err != nil {
		t.Fatalf("CountTokens returned error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 tokens, got %d", count)
	}
}
