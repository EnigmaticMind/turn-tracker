package core

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

var defaultNameList = []string{
	"Player", "Gamer", "Champion", "Hero", "Warrior",
	"Explorer", "Adventurer", "Master", "Legend", "Ace",
}

// GenerateRandomColor generates a random hex color
func GenerateRandomColor() string {
	r, _ := rand.Int(rand.Reader, big.NewInt(256))
	g, _ := rand.Int(rand.Reader, big.NewInt(256))
	b, _ := rand.Int(rand.Reader, big.NewInt(256))
	return fmt.Sprintf("#%02X%02X%02X", r.Int64(), g.Int64(), b.Int64())
}

// GenerateRandomDisplayName generates a random display name
func GenerateRandomDisplayName() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(defaultNameList))))
	num, _ := rand.Int(rand.Reader, big.NewInt(9999))
	return fmt.Sprintf("%s%d", defaultNameList[n.Int64()], num.Int64())
}

// InitializeClientProfile sets client's display name and color
// Generates random values if not provided
func InitializeClientProfile(client *Client, displayName, color string) {
	if displayName == "" {
		client.DisplayName = GenerateRandomDisplayName()
	} else {
		client.DisplayName = displayName
	}

	if color == "" {
		client.Color = GenerateRandomColor()
	} else {
		client.Color = color
	}
}
