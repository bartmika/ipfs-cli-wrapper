package randomkit

import (
	"crypto/rand"
)

// RandomGenerator is an interface for generating random data.
type RandomGenerator interface {
	Read(p []byte) (n int, err error)
}

// CryptoRandomGenerator uses crypto/rand for generating random data.
type CryptoRandomGenerator struct{}

// Read fills the slice with random bytes using crypto/rand.
func (g *CryptoRandomGenerator) Read(p []byte) (n int, err error) {
	return rand.Read(p)
}

// DefaultGenerator is the default generator used for random data.
var DefaultGenerator RandomGenerator = &CryptoRandomGenerator{}

// String generates a random string of the specified length `n` using the provided random generator.
func String(n int) string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)
	DefaultGenerator.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}
