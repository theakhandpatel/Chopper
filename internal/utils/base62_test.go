package utils

import (
	"testing"
)

func TestBase62EncodingAndDecoding(t *testing.T) {
	tests := []struct {
		input uint64
	}{
		{0},
		{1},
		{123},
		{987654},
		{uint64(base)},
	}

	for _, test := range tests {
		encoded := EncodeToBase62(test.input)
		decoded := DecodeFromBase62(encoded)

		if decoded != test.input {
			t.Errorf("Input: %d, Encoded: %s, Decoded: %d, Expected: %d", test.input, encoded, decoded, test.input)
		}
	}
}
