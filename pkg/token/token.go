package token

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand/v2"
)

const (
	base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	tokenLength = 8
)

// Generate creates a unique token from the given URL.
// A random nonce is added to avoid deterministic output for the same URL.
func Generate(url string) string {
	nonce := rand.Uint64()
	input := fmt.Sprintf("%s:%d", url, nonce)

	hash := sha256.Sum256([]byte(input))
	num := binary.BigEndian.Uint64(hash[:8])

	return encodeBase62(num, tokenLength)
}

func encodeBase62(num uint64, length int) string {
	result := make([]byte, length)
	base := uint64(len(base62Chars))

	for i := length - 1; i >= 0; i-- {
		result[i] = base62Chars[num%base]
		num /= base
	}

	return string(result)
}
