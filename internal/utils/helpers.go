package utils

import (
	"github.com/jaevor/go-nanoid"
)

var shortCodeof8Generator func() string

func init() {
	var err error
	shortCodeof8Generator, err = nanoid.Standard(8)
	if err != nil {
		panic(err)
	}
}

// generates a short URL based on the given long URL.
func GetShortCode(length int) string {

	shortCode := shortCodeof8Generator()
	// Crop the shortCode to the specified length.
	if length > 0 && length < len(shortCode) {
		shortCode = shortCode[:length]
	}

	return shortCode
}
