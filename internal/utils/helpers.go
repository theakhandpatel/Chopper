package utils

import (
	"crypto/md5"
	"encoding/binary"
)

// generates a short URL based on the given long URL.
func Shorten(longURL string) string {
	uid := generateUniqueIntIdentifier(longURL) // Generate a unique integer identifier based on the long URL.
	shortURL := EncodeToBase62(uint64(uid))     // Convert the identifier to a Base62 encoded string.
	return shortURL
}

// generates a unique 32-bit integer identifier from a URL .
func generateUniqueIntIdentifier(url string) uint32 {
	hash := md5.Sum([]byte(url))
	return binary.BigEndian.Uint32(hash[:4]) // Convert the first 4 bytes of the hash to a 32-bit integer.
}
