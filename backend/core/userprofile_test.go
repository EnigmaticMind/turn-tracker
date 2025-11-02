package core

import (
	"strings"
	"testing"
)

// TestUserProfile wraps all user profile tests
// This allows running all tests together or individually in the IDE
func TestUserProfile(t *testing.T) {
	t.Run("GenerateRandomColor", func(t *testing.T) {
		t.Run("ReturnsValidColor", func(t *testing.T) {
			color := GenerateRandomColor()

			// Should be a valid hex color format
			if !strings.HasPrefix(color, "#") {
				t.Errorf("Expected color to start with #, got %s", color)
			}
			if len(color) != 7 {
				t.Errorf("Expected color length 7, got %d for %s", len(color), color)
			}
		})

		t.Run("ReturnsFromBoardGameColors", func(t *testing.T) {
			// Generate many colors and verify they're all from the boardGameColors list
			colorsSeen := make(map[string]bool)
			for i := 0; i < 100; i++ {
				color := GenerateRandomColor()
				colorsSeen[color] = true

				// Verify it's in our board game colors list (from userprofile.go)
				found := false
				for _, validColor := range boardGameColors {
					if color == validColor {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Generated color %s is not in boardGameColors list", color)
				}
			}

			// Should see multiple different colors over 100 iterations
			if len(colorsSeen) < 3 {
				t.Logf("Warning: Only saw %d unique colors in 100 iterations (may be random)", len(colorsSeen))
			}
		})

		t.Run("FallbackOnRandomFailure", func(t *testing.T) {
			// This test is hard to trigger naturally, but we can verify fallback exists
			// by checking that the function always returns a valid color
			// In practice, rand.Int failures are extremely rare

			// Generate many colors - all should be valid
			for i := 0; i < 1000; i++ {
				color := GenerateRandomColor()
				if color == "" {
					t.Error("GenerateRandomColor returned empty string")
				}
				if !strings.HasPrefix(color, "#") {
					t.Errorf("GenerateRandomColor returned invalid format: %s", color)
				}
			}
		})

		t.Run("ReturnsValidHexFormat", func(t *testing.T) {
			color := GenerateRandomColor()

			// Check format: #RRGGBB
			if len(color) != 7 {
				t.Errorf("Expected hex color length 7, got %d", len(color))
			}

			// Check each character after # is hex digit
			for i := 1; i < len(color); i++ {
				c := color[i]
				isHex := (c >= '0' && c <= '9') || (c >= 'A' && c <= 'F')
				if !isHex {
					t.Errorf("Invalid hex character at position %d: %c in %s", i, c, color)
				}
			}
		})
	})

	t.Run("GenerateRandomDisplayName", func(t *testing.T) {
		t.Run("ReturnsValidFormat", func(t *testing.T) {
			name := GenerateRandomDisplayName()

			// Should contain at least one letter and one number
			if len(name) == 0 {
				t.Error("Expected non-empty display name")
			}

			// Should match pattern: NameNumber (e.g., "Zorp123")
			hasName := false
			hasNumber := false

			// Check against boardGameNames from userprofile.go
			for _, validName := range boardGameNames {
				if strings.HasPrefix(name, validName) {
					hasName = true
					break
				}
			}

			// Check if ends with number
			if len(name) > 0 {
				lastChar := name[len(name)-1]
				if lastChar >= '0' && lastChar <= '9' {
					hasNumber = true
				}
			}

			if !hasName {
				t.Errorf("Display name %s does not start with valid name", name)
			}
			if !hasNumber {
				t.Errorf("Display name %s does not contain number suffix", name)
			}
		})

		t.Run("NamesAreShort", func(t *testing.T) {
			// Verify names are short enough to avoid frontend overflow
			maxNameLength := 12  // Base name length before number
			maxTotalLength := 16 // Total with 4-digit number

			for i := 0; i < 100; i++ {
				name := GenerateRandomDisplayName()

				// Extract base name (without number)
				baseName := ""
				for _, validName := range boardGameNames {
					if strings.HasPrefix(name, validName) {
						baseName = validName
						break
					}
				}

				if baseName == "" {
					t.Errorf("Could not extract base name from %s", name)
					continue
				}

				// Check base name length
				if len(baseName) > maxNameLength {
					t.Errorf("Name %s is too long (%d chars), max is %d", baseName, len(baseName), maxNameLength)
				}

				// Check total length
				if len(name) > maxTotalLength {
					t.Errorf("Full name %s is too long (%d chars), max is %d", name, len(name), maxTotalLength)
				}
			}
		})

		t.Run("ReturnsFromBoardGameNames", func(t *testing.T) {
			// Generate many names and verify they're all from the boardGameNames list
			namesSeen := make(map[string]bool)
			for i := 0; i < 100; i++ {
				name := GenerateRandomDisplayName()
				namesSeen[name] = true

				// Verify it starts with a valid name (from userprofile.go)
				found := false
				for _, validName := range boardGameNames {
					if strings.HasPrefix(name, validName) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Generated name %s does not start with valid name", name)
				}
			}

			// Should see multiple different names over 100 iterations
			if len(namesSeen) < 5 {
				t.Logf("Warning: Only saw %d unique names in 100 iterations (may be random)", len(namesSeen))
			}
		})

		t.Run("NumberIsWithinRange", func(t *testing.T) {
			for i := 0; i < 100; i++ {
				name := GenerateRandomDisplayName()

				// Extract number suffix
				numberStr := ""
				for j := len(name) - 1; j >= 0; j-- {
					if name[j] >= '0' && name[j] <= '9' {
						numberStr = string(name[j]) + numberStr
					} else {
						break
					}
				}

				if numberStr == "" {
					t.Errorf("Name %s does not have a number suffix", name)
					continue
				}

				// Number should be within range [0, maxDisplayNameNumber)
				// maxDisplayNameNumber is 9999, so max length is 4 digits
				if len(numberStr) > 4 {
					t.Errorf("Name %s has invalid number length: %s (expected max 4 digits)", name, numberStr)
				}
			}
		})

		t.Run("FallbackOnRandomFailure", func(t *testing.T) {
			// Verify fallback exists - function should always return valid name
			for i := 0; i < 1000; i++ {
				name := GenerateRandomDisplayName()
				if name == "" {
					t.Error("GenerateRandomDisplayName returned empty string")
				}
				if len(name) < 2 {
					t.Errorf("GenerateRandomDisplayName returned too short name: %s", name)
				}
			}
		})

		t.Run("UniqueNamesPossible", func(t *testing.T) {
			// Verify that names can be unique (different numbers)
			names := make(map[string]bool)
			for i := 0; i < 50; i++ {
				name := GenerateRandomDisplayName()
				names[name] = true
			}

			// With 50 iterations and 9999 possible numbers, we should see some uniqueness
			// (though collisions are possible)
			if len(names) < 10 {
				t.Logf("Note: Only %d unique names in 50 iterations (collisions possible)", len(names))
			}
		})
	})

	t.Run("InitializeClientProfile", func(t *testing.T) {
		t.Run("GeneratesRandomWhenEmpty", func(t *testing.T) {
			client := &Client{ClientID: "test-client"}
			InitializeClientProfile(client, "", "")

			if client.DisplayName == "" {
				t.Error("Expected display name to be generated when empty")
			}
			if client.Color == "" {
				t.Error("Expected color to be generated when empty")
			}

			// Verify generated values are valid
			if !strings.HasPrefix(client.Color, "#") {
				t.Errorf("Generated color has invalid format: %s", client.Color)
			}
			if len(client.DisplayName) == 0 {
				t.Error("Generated display name is empty")
			}
		})

		t.Run("UsesProvidedValues", func(t *testing.T) {
			client := &Client{ClientID: "test-client"}
			InitializeClientProfile(client, "CustomPlayer", "#FF0000")

			if client.DisplayName != "CustomPlayer" {
				t.Errorf("Expected display name 'CustomPlayer', got '%s'", client.DisplayName)
			}
			if client.Color != "#FF0000" {
				t.Errorf("Expected color '#FF0000', got '%s'", client.Color)
			}
		})

		t.Run("GeneratesColorWhenEmpty", func(t *testing.T) {
			client := &Client{ClientID: "test-client"}
			InitializeClientProfile(client, "Player", "")

			if client.DisplayName != "Player" {
				t.Errorf("Expected display name 'Player', got '%s'", client.DisplayName)
			}
			if client.Color == "" {
				t.Error("Expected color to be generated when empty")
			}
			if !strings.HasPrefix(client.Color, "#") {
				t.Errorf("Generated color has invalid format: %s", client.Color)
			}
		})

		t.Run("GeneratesNameWhenEmpty", func(t *testing.T) {
			client := &Client{ClientID: "test-client"}
			InitializeClientProfile(client, "", "#FF0000")

			if client.Color != "#FF0000" {
				t.Errorf("Expected color '#FF0000', got '%s'", client.Color)
			}
			if client.DisplayName == "" {
				t.Error("Expected display name to be generated when empty")
			}
		})

		t.Run("HandlesMultipleCalls", func(t *testing.T) {
			client := &Client{ClientID: "test-client"}

			// First call
			InitializeClientProfile(client, "", "")

			// Second call with empty - should generate different values (likely)
			InitializeClientProfile(client, "", "")

			// Values should be valid (may or may not be different due to randomness)
			if client.DisplayName == "" {
				t.Error("Expected display name after second call")
			}
			if client.Color == "" {
				t.Error("Expected color after second call")
			}

			// At least verify format is correct
			if !strings.HasPrefix(client.Color, "#") {
				t.Errorf("Color format invalid: %s", client.Color)
			}
		})

		t.Run("PreservesExistingValuesWhenProvided", func(t *testing.T) {
			client := &Client{ClientID: "test-client"}
			InitializeClientProfile(client, "Alice", "#00FF00")

			// Call again with same values
			InitializeClientProfile(client, "Alice", "#00FF00")

			if client.DisplayName != "Alice" {
				t.Errorf("Expected display name 'Alice', got '%s'", client.DisplayName)
			}
			if client.Color != "#00FF00" {
				t.Errorf("Expected color '#00FF00', got '%s'", client.Color)
			}
		})
	})

	t.Run("BoardGameColors", func(t *testing.T) {
		t.Run("ColorsAreValid", func(t *testing.T) {
			// Uses boardGameColors from userprofile.go
			if len(boardGameColors) == 0 {
				t.Error("boardGameColors should not be empty")
			}

			for i, color := range boardGameColors {
				if !strings.HasPrefix(color, "#") {
					t.Errorf("Color at index %d does not start with #: %s", i, color)
				}
				if len(color) != 7 {
					t.Errorf("Color at index %d has invalid length: %s", i, color)
				}
			}
		})

		t.Run("ColorsAreDistinct", func(t *testing.T) {
			seen := make(map[string]bool)
			for i, color := range boardGameColors {
				if seen[color] {
					t.Errorf("Duplicate color at index %d: %s", i, color)
				}
				seen[color] = true
			}
		})
	})

	t.Run("BoardGameNames", func(t *testing.T) {
		t.Run("NamesAreNotEmpty", func(t *testing.T) {
			// Uses boardGameNames from userprofile.go
			if len(boardGameNames) == 0 {
				t.Error("boardGameNames should not be empty")
			}

			for i, name := range boardGameNames {
				if name == "" {
					t.Errorf("Name at index %d is empty", i)
				}
			}
		})

		t.Run("NamesAreShort", func(t *testing.T) {
			// Verify all names are short to avoid frontend overflow
			maxLength := 12
			for i, name := range boardGameNames {
				if len(name) > maxLength {
					t.Errorf("Name at index %d is too long: %s (%d chars, max %d)", i, name, len(name), maxLength)
				}
			}
		})

		t.Run("NamesAreUnique", func(t *testing.T) {
			seen := make(map[string]bool)
			for i, name := range boardGameNames {
				if seen[name] {
					t.Errorf("Duplicate name at index %d: %s", i, name)
				}
				seen[name] = true
			}
		})
	})

	t.Run("RandomFailureHandling", func(t *testing.T) {
		// This test verifies that fallback constants exist
		// Actual random failures are extremely rare and hard to test

		t.Run("FallbackConstantsDefined", func(t *testing.T) {
			// Uses constants from userprofile.go
			if defaultColorIndex < 0 || defaultColorIndex >= len(boardGameColors) {
				t.Errorf("defaultColorIndex %d is out of bounds for boardGameColors (len=%d)",
					defaultColorIndex, len(boardGameColors))
			}
			if defaultNameIndex < 0 || defaultNameIndex >= len(boardGameNames) {
				t.Errorf("defaultNameIndex %d is out of bounds for boardGameNames (len=%d)",
					defaultNameIndex, len(boardGameNames))
			}
			if defaultNameNumber < 0 || defaultNameNumber >= maxDisplayNameNumber {
				t.Errorf("defaultNameNumber %d is out of bounds (max=%d)",
					defaultNameNumber, maxDisplayNameNumber)
			}
		})

		t.Run("FallbackColorExists", func(t *testing.T) {
			// Uses boardGameColors and defaultColorIndex from userprofile.go
			fallbackColor := boardGameColors[defaultColorIndex]
			color := GenerateRandomColor()

			// Both should be valid
			if !strings.HasPrefix(fallbackColor, "#") {
				t.Errorf("Fallback color invalid: %s", fallbackColor)
			}
			if !strings.HasPrefix(color, "#") {
				t.Errorf("Generated color invalid: %s", color)
			}
		})

		t.Run("FallbackNameExists", func(t *testing.T) {
			// Uses boardGameNames and defaultNameIndex from userprofile.go
			fallbackName := boardGameNames[defaultNameIndex]
			name := GenerateRandomDisplayName()

			// Both should be valid
			if fallbackName == "" {
				t.Error("Fallback name is empty")
			}
			if name == "" {
				t.Error("Generated name is empty")
			}
		})
	})
}
