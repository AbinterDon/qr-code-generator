package token

import (
	"testing"
)

func TestGenerate_Length(t *testing.T) {
	tok := Generate("https://example.com")
	if len(tok) != tokenLength {
		t.Errorf("Generate() length = %d; want %d", len(tok), tokenLength)
	}
}

func TestGenerate_OnlyBase62Chars(t *testing.T) {
	tok := Generate("https://example.com")
	for _, c := range tok {
		if !isBase62(byte(c)) {
			t.Errorf("Generate() contains non-base62 character: %c", c)
		}
	}
}

func TestGenerate_Nondeterministic(t *testing.T) {
	url := "https://example.com"
	tokens := make(map[string]struct{})

	for range 100 {
		tok := Generate(url)
		tokens[tok] = struct{}{}
	}

	if len(tokens) < 90 {
		t.Errorf("Generate() appears deterministic: only %d unique tokens in 100 calls", len(tokens))
	}
}

func isBase62(c byte) bool {
	for i := range base62Chars {
		if base62Chars[i] == c {
			return true
		}
	}
	return false
}
