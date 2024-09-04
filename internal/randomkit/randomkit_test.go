package randomkit_test

import (
	"testing"

	"github.com/bartmika/ipfs-cli-wrapper/internal/randomkit"
)

// TestStringLength checks if the generated string has the correct length.
func TestStringLength(t *testing.T) {
	length := 10
	result := randomkit.String(length)

	if len(result) != length {
		t.Errorf("Expected length %d, but got %d", length, len(result))
	}
}

// TestStringCharacters checks if the generated string contains only alphanumeric characters.
func TestStringCharacters(t *testing.T) {
	length := 50
	result := randomkit.String(length)
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	for _, char := range result {
		if !containsRune(alphanum, char) {
			t.Errorf("Generated string contains invalid character: %v", char)
		}
	}
}

// TestStringUniqueness checks if two generated strings of the same length are different.
func TestStringUniqueness(t *testing.T) {
	length := 15
	str1 := randomkit.String(length)
	str2 := randomkit.String(length)

	if str1 == str2 {
		t.Errorf("Expected different strings but got identical ones: %s and %s", str1, str2)
	}
}

// containsRune is a helper function to check if a rune is in a string.
func containsRune(str string, char rune) bool {
	for _, c := range str {
		if c == char {
			return true
		}
	}
	return false
}
