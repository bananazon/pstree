package pstree

import (
	"testing"

	"github.com/bananazon/pstree/util"
	"github.com/stretchr/testify/assert"
)

var colorScheme = ColorSchemes["windows10"]

// Helper function for tests that mimics the expected Colorize function
func Colorize(text, color string) string {
	if text == "" {
		return ""
	}

	result := text
	switch color {
	case "red":
		Color256Red(colorScheme, &result)
	case "green":
		Color256Green(colorScheme, &result)
	case "yellow":
		Color256Yellow(colorScheme, &result)
	case "blue":
		Color256Blue(colorScheme, &result)
	case "magenta", "purple":
		Color256Magenta(colorScheme, &result)
	case "cyan":
		Color256Cyan(colorScheme, &result)
	case "white":
		Color256White(colorScheme, &result)
	case "black":
		Color256Black(colorScheme, &result)
	case "bold":
		Color256WhiteBold(colorScheme, &result)
	default:
		return text
	}
	return result
}

func TestHasColorSupport(t *testing.T) {
	// Just verify that the function returns without error
	colorSupport, colorCount := util.HasColorSupport()

	// The result depends on the environment, but we can at least check the types
	assert.IsType(t, true, colorSupport)
	assert.IsType(t, 0, colorCount)

	// Color count should be non-negative
	assert.GreaterOrEqual(t, colorCount, 0)
}

func TestColorize(t *testing.T) {
	// Test colorizing text with various colors
	text := "test"

	// Test with all available colors
	assert.NotEqual(t, text, Colorize(text, "red"))
	assert.NotEqual(t, text, Colorize(text, "green"))
	assert.NotEqual(t, text, Colorize(text, "yellow"))
	assert.NotEqual(t, text, Colorize(text, "blue"))
	assert.NotEqual(t, text, Colorize(text, "magenta"))
	assert.NotEqual(t, text, Colorize(text, "cyan"))
	assert.NotEqual(t, text, Colorize(text, "white"))
	assert.NotEqual(t, text, Colorize(text, "black"))

	// Test with bold
	assert.NotEqual(t, text, Colorize(text, "bold"))

	// Test with an invalid color (should return the original text)
	assert.Equal(t, text, Colorize(text, "invalid"))

	// Test with an empty color (should return the original text)
	assert.Equal(t, text, Colorize(text, ""))

	// Test with an empty text
	assert.Empty(t, Colorize("", "red"))
}
