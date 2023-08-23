package utils

import (
	"crypto/md5"
	"encoding/binary"
)

func Shorten(longURL string) string {
	uid := generateUniqueIntIdentifier(longURL)
	shortURL := EncodeToBase62((uint64(uid)))
	return shortURL
}

func generateUniqueIntIdentifier(url string) uint32 {
	hash := md5.Sum([]byte(url))
	return binary.BigEndian.Uint32(hash[:4])
}
