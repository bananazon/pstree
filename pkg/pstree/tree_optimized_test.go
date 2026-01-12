package pstree

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// setupTestProcessTree creates a simple process tree for testing
func setupTestProcessTree() *ProcessTree {
	// Create test processes
	processes := []Process{
		{PID: 1, PPID: 0, Command: "init"},
		{PID: 2, PPID: 1, Command: "child1"},
		{PID: 3, PPID: 1, Command: "child2"},
		{PID: 4, PPID: 2, Command: "grandchild"},
	}

	// Create a logger that discards output
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create a process tree
	processTree := &ProcessTree{
		Logger:         logger,
		Nodes:          processes,
		PidToIndexMap:  make(map[int32]int),
		IndexToPidMap:  make(map[int]int32),
		DisplayOptions: DisplayOptions{},
	}

	// Initialize PID to index mapping
	for i, proc := range processes {
		processTree.PidToIndexMap[proc.PID] = i
		processTree.IndexToPidMap[i] = proc.PID
	}

	return processTree
}

// TestOptimizedBuildTree tests that the optimized BuildTree function works correctly
func TestOptimizedBuildTree(t *testing.T) {
	// Create a test process tree
	processTree := setupTestProcessTree()

	// Run the optimized BuildTree function
	processTree.OptimizedBuildTree()

	// Verify parent-child relationships
	assert.Equal(t, -1, processTree.Nodes[0].Parent) // init has no parent
	assert.Equal(t, 0, processTree.Nodes[1].Parent)  // child1's parent is init
	assert.Equal(t, 0, processTree.Nodes[2].Parent)  // child2's parent is init
	assert.Equal(t, 1, processTree.Nodes[3].Parent)  // grandchild's parent is child1

	// Verify child relationships
	assert.True(t, processTree.Nodes[0].Child == 1 || processTree.Nodes[0].Child == 2) // init's first child is either child1 or child2
	assert.Equal(t, 3, processTree.Nodes[1].Child)  // child1's child is grandchild
	assert.Equal(t, -1, processTree.Nodes[2].Child) // child2 has no children
	assert.Equal(t, -1, processTree.Nodes[3].Child) // grandchild has no children

	// Verify sibling relationships
	if processTree.Nodes[0].Child == 1 {
		assert.Equal(t, 2, processTree.Nodes[1].Sister) // If init's first child is child1, then child1's sister is child2
	} else {
		assert.Equal(t, 1, processTree.Nodes[2].Sister) // Otherwise, child2's sister is child1
	}
}

// BenchmarkOriginalBuildTree benchmarks the original BuildTree function
func BenchmarkOriginalBuildTree(b *testing.B) {
	for i := 0; i < b.N; i++ {
		processTree := setupTestProcessTree()
		processTree.BuildTree()
	}
}

// BenchmarkOptimizedBuildTree benchmarks the optimized BuildTree function
func BenchmarkOptimizedBuildTree(b *testing.B) {
	for i := 0; i < b.N; i++ {
		processTree := setupTestProcessTree()
		processTree.OptimizedBuildTree()
	}
}

// TestBuildTreeWithTimeout tests the original BuildTree function with a timeout
func TestBuildTreeWithTimeout(t *testing.T) {
	// Create a test process tree
	processTree := setupTestProcessTree()

	// Run the BuildTree function with a timeout
	done := make(chan bool)
	go func() {
		processTree.BuildTree()
		done <- true
	}()

	// Wait for the function to complete or timeout
	select {
	case <-done:
		// Function completed successfully
		t.Log("BuildTree completed successfully")
	case <-time.After(5 * time.Second):
		t.Fatal("BuildTree timed out after 5 seconds")
	}

	// Verify parent-child relationships
	assert.Equal(t, -1, processTree.Nodes[0].Parent) // init has no parent
	assert.Equal(t, 0, processTree.Nodes[1].Parent)  // child1's parent is init
	assert.Equal(t, 0, processTree.Nodes[2].Parent)  // child2's parent is init
	assert.Equal(t, 1, processTree.Nodes[3].Parent)  // grandchild's parent is child1
}
