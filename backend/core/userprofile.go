package core

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"turn-tracker/backend/helpers"
)

const (
	// Display name generation constants
	maxDisplayNameNumber = 9999
	defaultColorIndex    = 0 // Fallback color index if random fails
	defaultNameIndex     = 0 // Fallback name index if random fails
	defaultNameNumber    = 1 // Fallback number if random fails
)

var (
	// Board game themed colors (classic player piece colors)
	// Colors chosen to be distinct and visible on various backgrounds
	boardGameColors = []string{
		"#FF0000", // Red (Risk, Catan)
		"#0066FF", // Blue (Risk, Catan)
		"#00AA00", // Green (Risk, Catan)
		"#FFD700", // Gold/Yellow (Risk, Catan)
		"#FF6600", // Orange (Risk)
		"#9932CC", // Dark Orchid/Purple (Risk, Catan)
		"#8B4513", // Saddle Brown (Catan)
		"#FF1493", // Deep Pink (Catan)
		"#00CED1", // Dark Turquoise/Teal (Catan)
	}

	// Funny/fantasy display names (short to avoid frontend overflow)
	boardGameNames = []string{
		// Obviously fake/funny names
		"Zorp", "Blip", "Flum", "Glork", "Snarf",
		"Blorp", "Quark", "Zing", "Fizz", "Wump",
		// Fantasy names
		"Zara", "Kira", "Luna", "Zeph", "Nyx",
		"Rex", "Vex", "Zed", "Lex", "Jax",
		// Silly/funny
		"Squish", "Bloop", "Plop", "Ding", "Bonk",
		"Zap", "Zoom", "Boom", "Ping", "Pow",
		// More fantasy (removed duplicates: Zara, Rex)
		"Axel", "Rune", "Faye", "Jade", "Sage",
		"Zane", "Mira", "Kai", "Blaze", "Nova",
		// Fun/gamey
		"Pixel", "Byte", "Code", "Data", "Bit",
	}
)

// GenerateRandomColor generates a random board game themed color
// Returns a fallback color if random number generation fails
func GenerateRandomColor() string {
	colorCount := len(boardGameColors)

	idx, err := rand.Int(rand.Reader, big.NewInt(int64(colorCount)))
	if err != nil {
		// Fallback to first color if random fails
		return boardGameColors[defaultColorIndex]
	}

	return boardGameColors[idx.Int64()]
}

// GenerateRandomDisplayName generates a random funny/fantasy themed display name
// Returns a fallback name if random number generation fails
func GenerateRandomDisplayName() string {
	nameCount := len(boardGameNames)

	nameIdx, err := rand.Int(rand.Reader, big.NewInt(int64(nameCount)))
	if err != nil {
		// Fallback to first name if random fails
		nameIdx = big.NewInt(defaultNameIndex)
	}

	num, err := rand.Int(rand.Reader, big.NewInt(maxDisplayNameNumber))
	if err != nil {
		// Fallback to default number if random fails
		num = big.NewInt(defaultNameNumber)
	}

	return fmt.Sprintf("%s%d", boardGameNames[nameIdx.Int64()], num.Int64())
}

// InitializeClientProfile sets client's display name and color
// Generates random funny/fantasy themed values if not provided
func InitializeClientProfile(client *Client, displayName, color string) {
	if displayName == "" || !helpers.IsValidDisplayName(displayName) {
		client.DisplayName = GenerateRandomDisplayName()
	} else {
		client.DisplayName = displayName
	}

	if color == "" || !helpers.IsValidHexColor(color) {
		client.Color = GenerateRandomColor()
	} else {
		client.Color = color
	}
}
