package password

import (
	"crypto/rand"
	"math/big"
	"os"
)

func GenerateSecurePassword(length int, includeSymbols bool) (string, error) {
	letters := os.Getenv("LETTERS")
	if letters == "" {
		letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	}

	symbols := os.Getenv("SYMBOLS")
	if symbols == "" {
		symbols = "!@#$%^&*()-_=+[]{}|;:,.<>?/~`"
	}

	charCycle := letters
	if includeSymbols {
		charCycle += symbols
	}

	b := make([]byte, length)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charCycle))))
		if err != nil {
			return "", err
		}
		b[i] = charCycle[num.Int64()]
	}
	return string(b), nil
}
