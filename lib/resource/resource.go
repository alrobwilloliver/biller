package resource

import (
	"crypto/rand"
	"encoding/binary"
	"regexp"
)

var resourceIDRegexp = regexp.MustCompile("^[a-z]([a-z0-9-]{0,61}[a-z0-9])?$")

func ValidResourceID(in string) bool {
	return resourceIDRegexp.Match([]byte(in))
}

var nanoIDRegexp = regexp.MustCompile("^[a-z][0-9a-z-]{10}[0-9a-z]$")

func ValidNanoID(in string) bool {
	return nanoIDRegexp.Match([]byte(in))
}

// defaultAlphabet is the alphabet used for Nano ID characters.
var startAlphabet = []rune("abcdefghijklmnopqrstuvwxyz")
var midAlphabet = []rune("-0123456789abcdefghijklmnopqrstuvwxyz")
var endAlphabet = []rune("0123456789abcdefghijklmnopqrstuvwxyz")

// NewNanoID generates secure URL-friendly unique ID.
// Accepts required parameter - length of the ID to be generated.
func NewNanoID(size int) (string, error) {
	bytes := make([]byte, size*4)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	id := make([]rune, size)
	for i := 0; i < size; i++ {
		alphabet := startAlphabet
		if i > 0 && i < size-1 {
			alphabet = midAlphabet
		} else if i == size-1 {
			alphabet = endAlphabet
		}
		n := binary.LittleEndian.Uint32(bytes[i*4:])
		id[i] = alphabet[n%uint32(len(alphabet))]
	}
	return string(id), nil
}
