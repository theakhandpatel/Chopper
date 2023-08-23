package utils

import (
	"math"
	"strings"
)

const alphabet = "123456789abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ"
const base = uint64(len(alphabet))

func EncodeToBase62(num uint64) string {
	encoded := ""
	for num > 0 {
		remainder := num % base
		num = num / base
		encoded = string(alphabet[remainder]) + encoded
	}
	return encoded
}

func DecodeFromBase62(encoded string) uint64 {
	decoded := uint64(0)
	power := len(encoded) - 1
	for _, char := range encoded {
		index := uint64(strings.IndexRune(alphabet, char))
		decoded += index * uint64(math.Pow(float64(base), float64(power)))
		power--
	}
	return decoded
}
