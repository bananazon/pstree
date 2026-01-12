package pstree

import (
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// setupTestLogger creates a logger for testing
func setupTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

// TestNewProcessTree tests the creation of a new process tree
func TestNewProcessTree(t *testing.T) {
	// Create test processes
	proc1 := Process{PID: 1, PPID: 0, Command: "init"}
	proc2 := Process{PID: 100, PPID: 1, Command: "proc2"}
	processes := []Process{proc1, proc2}

	// Create display options
	displayOptions := DisplayOptions{
		ShowPIDs: true,
	}

	// Create a process tree
	processTree := NewProcessTree(0, setupTestLogger(), processes, displayOptions)

	// Verify that the tree was created correctly
	assert.Equal(t, 2, len(processTree.Nodes))
	assert.Equal(t, true, processTree.DisplayOptions.ShowPIDs)

	// Verify PID to index mapping
	idx1, ok := processTree.PidToIndexMap[1]
	assert.True(t, ok)
	idx2, ok := processTree.PidToIndexMap[100]
	assert.True(t, ok)

	// Verify process data
	assert.Equal(t, int32(1), processTree.Nodes[idx1].PID)
	assert.Equal(t, "init", processTree.Nodes[idx1].Command)
	assert.Equal(t, int32(100), processTree.Nodes[idx2].PID)
	assert.Equal(t, "proc2", processTree.Nodes[idx2].Command)
}

// TestBuildTreeSimple tests the basic functionality of the BuildTree method with a minimal test case
func TestBuildTreeSimple(t *testing.T) {
	// Create minimal test processes
	processes := []Process{
		{PID: 1, PPID: 0, Command: "init"},
		{PID: 2, PPID: 1, Command: "child"},
	}

	// Create a process tree with minimal configuration
	processTree := &ProcessTree{
		Logger:        setupTestLogger(),
		Nodes:         processes,
		PidToIndexMap: map[int32]int{1: 0, 2: 1},
		IndexToPidMap: map[int]int32{0: 1, 1: 2},
	}

	// Build the tree
	processTree.BuildTree()

	// Verify basic parent-child relationship
	assert.Equal(t, -1, processTree.Nodes[0].Parent) // init has no parent
	assert.Equal(t, 0, processTree.Nodes[1].Parent)  // child's parent is init
	assert.Equal(t, 1, processTree.Nodes[0].Child)   // init's child is child
	assert.Equal(t, -1, processTree.Nodes[1].Child)  // child has no children
}

// TestMarkProcesses tests the MarkProcesses method
func TestMarkProcesses(t *testing.T) {
	logger := setupTestLogger()

	// Create test processes
	proc1 := Process{PID: 1, PPID: 0, Command: "init", Username: "root"}
	proc2 := Process{PID: 100, PPID: 1, Command: "proc2", Username: "user1"}
	proc3 := Process{PID: 200, PPID: 1, Command: "proc3", Username: "user2"}
	proc4 := Process{PID: 300, PPID: 100, Command: "proc4", Username: "user1"}

	processes := []Process{proc1, proc2, proc3, proc4}

	// Test case 1: Show all processes
	displayOptions1 := DisplayOptions{}
	processTree1 := NewProcessTree(0, logger, processes, displayOptions1)
	processTree1.MarkProcesses()

	// Verify that all processes are marked for display
	for i := range processTree1.Nodes {
		assert.True(t, processTree1.Nodes[i].Print)
	}

	// Test case 2: Filter by username
	displayOptions2 := DisplayOptions{
		Usernames: []string{"user1"},
	}
	processTree2 := NewProcessTree(0, logger, processes, displayOptions2)
	processTree2.MarkProcesses()

	// Verify that only processes with username "user1" and their ancestors are marked
	idx1 := processTree2.PidToIndexMap[1]
	idx2 := processTree2.PidToIndexMap[100]
	idx3 := processTree2.PidToIndexMap[200]
	idx4 := processTree2.PidToIndexMap[300]

	assert.True(t, processTree2.Nodes[idx1].Print)  // init is an ancestor of user1's processes
	assert.True(t, processTree2.Nodes[idx2].Print)  // proc2 belongs to user1
	assert.False(t, processTree2.Nodes[idx3].Print) // proc3 belongs to user2
	assert.True(t, processTree2.Nodes[idx4].Print)  // proc4 belongs to user1
}

// TestMarkUIDTransitions tests the MarkUIDTransitions method
func TestMarkUIDTransitions(t *testing.T) {
	// Skip this test since the Process struct doesn't have a UIDTransition field
	// This is a placeholder test that can be implemented when needed
	t.Skip("Skipping TestMarkUIDTransitions as it needs to be implemented properly")
}
