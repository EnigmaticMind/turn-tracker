package helpers

const (
	maxDisplayNameLength = 50
)

// IsValidDisplayName validates that a display name is not empty and not too long
func IsValidDisplayName(displayName string) bool {
	return len(displayName) > 0 && len(displayName) <= maxDisplayNameLength
}
