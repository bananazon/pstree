package pstree

import (
	"testing"
	
	"github.com/stretchr/testify/assert"
)

// TestDisplayOptions tests the DisplayOptions struct and its fields
func TestDisplayOptions(t *testing.T) {
	// Create a DisplayOptions struct with various settings
	options := DisplayOptions{
		ShowPIDs:           true,
		ShowPPIDs:          true,
		ShowOwner:          true,
		OrderBy:            "cpu",
		ShowUIDTransitions: true,
		MaxDepth:           3,
	}
	
	// Verify that the fields are set correctly
	assert.True(t, options.ShowPIDs)
	assert.True(t, options.ShowPPIDs)
	assert.True(t, options.ShowOwner)
	assert.Equal(t, "cpu", options.OrderBy)
	assert.True(t, options.ShowUIDTransitions)
	assert.Equal(t, 3, options.MaxDepth)
}
