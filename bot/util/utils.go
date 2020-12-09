package util

import (
	"crypto/rand"
	"math/big"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// RandString get random string containing just letters and numbers
// todo search better crypto/ function to generate such string
func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))

		b[i] = letters[num.Int64()]
	}

	return string(b)
}
