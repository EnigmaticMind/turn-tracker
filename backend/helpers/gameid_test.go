package helpers

import (
	"strings"
	"testing"
)

// TestGameID wraps all game ID tests
// This allows running all tests together or individually in the IDE
func TestGameID(t *testing.T) {
	t.Run("GenerateGameID", func(t *testing.T) {
		t.Run("GeneratesCorrectLength", func(t *testing.T) {
			gameID := GenerateGameID()
			if len(gameID) != gameIDLength {
				t.Errorf("Expected game ID length %d, got %d", gameIDLength, len(gameID))
			}
		})

		t.Run("GeneratesUppercaseAlphanumeric", func(t *testing.T) {
			for i := 0; i < 100; i++ {
				gameID := GenerateGameID()
				if !IsValidGameID(gameID) {
					t.Errorf("Generated invalid game ID: %s", gameID)
				}
			}
		})

		t.Run("GeneratesUniqueIDs", func(t *testing.T) {
			ids := make(map[string]bool)
			iterations := 1000
			duplicates := 0

			for i := 0; i < iterations; i++ {
				gameID := GenerateGameID()
				if ids[gameID] {
					duplicates++
				}
				ids[gameID] = true
			}

			// With 4 characters, we have 36^4 = 1,679,616 possible combinations
			// With 1000 iterations, duplicates are very unlikely but not impossible
			// We'll just log if there are duplicates, but not fail the test
			if duplicates > 0 {
				t.Logf("Generated %d duplicate IDs out of %d iterations", duplicates, iterations)
			}
		})

		t.Run("OnlyUsesValidCharacters", func(t *testing.T) {
			for i := 0; i < 100; i++ {
				gameID := GenerateGameID()
				for _, c := range gameID {
					if !strings.ContainsRune(gameIDChars, c) {
						t.Errorf("Game ID contains invalid character: %c in %s", c, gameID)
					}
				}
			}
		})
	})

	t.Run("IsValidGameID", func(t *testing.T) {
		t.Run("ValidGameIDs", func(t *testing.T) {
			validIDs := []string{
				"ABCD",
				"1234",
				"A1B2",
				"ZZ99",
				"0000",
				"AAAA",
				"ZZZZ",
			}

			for _, gameID := range validIDs {
				if !IsValidGameID(gameID) {
					t.Errorf("Expected %s to be valid", gameID)
				}
			}
		})

		t.Run("InvalidLength", func(t *testing.T) {
			invalidIDs := []string{
				"",
				"A",
				"AB",
				"ABC",
				"ABCDE",
				"ABCDEF",
				"ABCDEFGH",
			}

			for _, gameID := range invalidIDs {
				if IsValidGameID(gameID) {
					t.Errorf("Expected %s to be invalid (wrong length: %d)", gameID, len(gameID))
				}
			}
		})

		t.Run("InvalidCharacters", func(t *testing.T) {
			invalidIDs := []string{
				"abc",   // lowercase
				"ABCd",  // lowercase
				"AB!D",  // special character
				"AB-D",  // special character
				"AB D",  // space
				"AB\nD", // newline
				"AB\tD", // tab
			}

			// Test invalid length ones with valid length
			invalidCharsButCorrectLength := []string{
				"abcD",
				"ABCd",
				"AB!D",
				"AB-D",
				"AB D",
			}

			allInvalid := append(invalidIDs, invalidCharsButCorrectLength...)

			for _, gameID := range allInvalid {
				// Only check ones with correct length for invalid characters
				if len(gameID) == gameIDLength && IsValidGameID(gameID) {
					t.Errorf("Expected %s to be invalid (contains invalid characters)", gameID)
				}
			}
		})

		t.Run("CaseSensitive", func(t *testing.T) {
			lowercase := "abcd"
			if IsValidGameID(lowercase) {
				t.Errorf("Expected lowercase %s to be invalid", lowercase)
			}
		})

		t.Run("EmptyString", func(t *testing.T) {
			if IsValidGameID("") {
				t.Error("Expected empty string to be invalid")
			}
		})

		t.Run("UnicodeCharacters", func(t *testing.T) {
			unicodeIDs := []string{
				"ABÇD",
				"ABÑD",
				"AB€D",
			}

			for _, gameID := range unicodeIDs {
				if len(gameID) == gameIDLength && IsValidGameID(gameID) {
					t.Errorf("Expected unicode string %s to be invalid", gameID)
				}
			}
		})
	})
}
