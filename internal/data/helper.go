package data

import (
	"crypto/md5"
	"fmt"
	"io"
)

func shorten(longURL string) string {
	hash := md5.New()
	io.WriteString(hash, longURL)
	hashBytes := hash.Sum(nil)

	hashString := fmt.Sprintf("%x", hashBytes)
	return hashString[:6]
}
