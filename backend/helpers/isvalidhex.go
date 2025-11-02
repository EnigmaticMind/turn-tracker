package helpers

const (
	colorLength = 7 // #RRGGBB format
)

// IsValidHexColor validates that a color is in #RRGGBB format
func IsValidHexColor(color string) bool {
	if len(color) != colorLength {
		return false
	}
	if color[0] != '#' {
		return false
	}
	for i := 1; i < len(color); i++ {
		c := color[i]
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
