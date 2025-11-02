package helpers

import (
	"crypto/rand"
	"math/big"
)

const (
	gameIDLength = 4
	gameIDChars  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// GenerateGameID generates a unique 4-character uppercase alphanumeric game ID
func GenerateGameID() string {
	max := big.NewInt(int64(len(gameIDChars)))
	b := make([]byte, gameIDLength)
	for i := range b {
		n, _ := rand.Int(rand.Reader, max)
		b[i] = gameIDChars[n.Int64()]
	}
	return string(b)
}

// IsValidGameID validates that a game ID is the correct format
// Format: 6 characters, uppercase alphanumeric
// Optimized using byte-level checks for better performance
func IsValidGameID(gameID string) bool {
	if len(gameID) != gameIDLength {
		return false
	}
	// Use byte-level checks for ASCII characters (faster than rune iteration)
	for i := 0; i < gameIDLength; i++ {
		c := gameID[i]
		// Check if character is A-Z or 0-9
		if !((c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
}
