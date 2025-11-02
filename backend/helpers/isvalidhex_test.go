package helpers

import "testing"

// TestIsValidHexColor wraps all hex color validation tests
func TestIsValidHexColor(t *testing.T) {
	tests := []struct {
		name  string
		color string
		want  bool
	}{
		// Valid colors - uppercase
		{"ValidUppercase", "#FF5733", true},
		{"ValidUppercaseShort", "#ABCDEF", true},
		{"ValidUppercaseNumbers", "#123456", true},
		{"ValidUppercaseMixed", "#FF00AA", true},

		// Valid colors - lowercase
		{"ValidLowercase", "#ff5733", true},
		{"ValidLowercaseShort", "#abcdef", true},
		{"ValidLowercaseNumbers", "#123456", true},
		{"ValidLowercaseMixed", "#ff00aa", true},
		{"ValidLowercaseMixedCase", "#5177c2", true}, // User's example

		// Valid colors - mixed case
		{"ValidMixedCase", "#Ff5733", true},
		{"ValidMixedCase2", "#aAbBcC", true},
		{"ValidMixedCase3", "#1a2B3c", true},

		// Invalid - wrong length
		{"TooShort", "#FF573", false},
		{"TooLong", "#FF57333", false},
		{"TooShort2", "#12345", false},
		{"TooLong2", "#1234567", false},
		{"Empty", "", false},
		{"JustHash", "#", false},

		// Invalid - missing hash
		{"NoHash", "FF5733", false},
		{"NoHashLowercase", "ff5733", false},

		// Invalid - wrong characters
		{"InvalidCharG", "#FF573G", false},
		{"InvalidCharLowercaseG", "#ff573g", false},
		{"InvalidCharZ", "#ABCDEZ", false},
		{"InvalidCharSpecial", "#FF573!", false},
		{"InvalidCharSpace", "#FF57 3", false},
		{"InvalidCharAt", "#FF5@33", false},
		{"InvalidCharX", "#FF573X", false},

		// Invalid - edge cases
		{"OnlyNumbers", "#123456", true}, // Valid, just numbers
		{"OnlyLetters", "#ABCDEF", true}, // Valid, just letters
		{"AllZeros", "#000000", true},    // Valid
		{"AllF", "#FFFFFF", true},        // Valid
		{"Allf", "#ffffff", true},        // Valid lowercase

		// Invalid format
		{"StartsWithSpace", " #FF5733", false},
		{"EndsWithSpace", "#FF5733 ", false},
		{"HasSpace", "#FF5 33", false},
		{"HashInMiddle", "FF#5733", false},
		{"MultipleHash", "##FF5733", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidHexColor(tt.color); got != tt.want {
				t.Errorf("IsValidHexColor(%q) = %v, want %v", tt.color, got, tt.want)
			}
		})
	}
}
