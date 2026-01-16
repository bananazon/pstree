package pstree

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitCompactMode(t *testing.T) {
	// Create test processes with identical commands under the same parent
	proc1 := Process{PID: 1, PPID: 0, Command: "init"}
	proc2 := Process{PID: 100, PPID: 1, Command: "bash", Args: []string{"arg1"}}
	proc3 := Process{PID: 200, PPID: 1, Command: "bash", Args: []string{"arg1"}}
	proc4 := Process{PID: 300, PPID: 1, Command: "bash", Args: []string{"arg2"}}
	proc5 := Process{PID: 400, PPID: 2, Command: "bash", Args: []string{"arg1"}}

	processes := []Process{proc1, proc2, proc3, proc4, proc5}
	processTree := NewProcessTree(0, setupTestLogger(), processes, DisplayOptions{})

	// Initialize compact mode
	processTree.InitCompactMode()

	// Verify that identical processes are grouped correctly
	// proc2 and proc3 should be grouped (same command, same args, same parent)
	assert.False(t, ShouldSkipProcess(0)) // proc1 should not be skipped
	assert.False(t, ShouldSkipProcess(1)) // proc2 should not be skipped (first in group)
	assert.True(t, ShouldSkipProcess(2))  // proc3 should be skipped (duplicate of proc2)
	assert.False(t, ShouldSkipProcess(3)) // proc4 should not be skipped (different args)
	assert.False(t, ShouldSkipProcess(4)) // proc5 should not be skipped (different parent)

	// Test GetProcessCount for the first process in a group
	// count, isThread := GetProcessCount(processes, 1)
	// assert.Equal(t, 2, count) // proc2 has one duplicate (proc3)
	// assert.False(t, isThread) // proc2 is not a thread

	// Test GetProcessCount for a process that is not in a group
	// count, isThread = GetProcessCount(processes, 0)
	// assert.Equal(t, 1, count) // proc1 has no duplicates
	// assert.False(t, isThread) // proc1 is not a thread
}

func TestShouldSkipProcess(t *testing.T) {
	// Create test processes
	proc1 := Process{PID: 1, PPID: 0, Command: "init"}
	proc2 := Process{PID: 100, PPID: 1, Command: "bash"}
	proc3 := Process{PID: 200, PPID: 1, Command: "bash"}

	processes := []Process{proc1, proc2, proc3}
	processTree := NewProcessTree(0, setupTestLogger(), processes, DisplayOptions{})

	// Initialize compact mode
	processTree.InitCompactMode()

	// Verify that ShouldSkipProcess returns the correct values
	assert.False(t, ShouldSkipProcess(0))
	assert.False(t, ShouldSkipProcess(1))
	assert.True(t, ShouldSkipProcess(2))

	// Test with an index that doesn't exist in the skipProcesses map
	assert.False(t, ShouldSkipProcess(999))
}

func TestGetProcessCount(t *testing.T) {
	// Create test processes with identical commands
	proc1 := Process{PID: 1, PPID: 0, Command: "init"}
	proc2 := Process{PID: 100, PPID: 1, Command: "bash"}
	proc3 := Process{PID: 200, PPID: 1, Command: "bash"}
	proc4 := Process{PID: 300, PPID: 1, Command: "bash"}

	processes := []Process{proc1, proc2, proc3, proc4}
	processTree := NewProcessTree(0, setupTestLogger(), processes, DisplayOptions{})

	// Initialize compact mode
	processTree.InitCompactMode()

	// Test GetProcessCount for the first process in a group
	count, _, _, _, _ := processTree.GetProcessCount(1)
	assert.Equal(t, 3, count) // proc2 has two duplicates (proc3 and proc4)
	// assert.False(t, isThread) // proc2 is not a thread

	// Test GetProcessCount for a process that should be skipped
	count, _, _, _, _ = processTree.GetProcessCount(2)
	assert.Equal(t, 1, count) // Not the first in group, so count is 1
	// assert.False(t, isThread)

	// Test with a process that has threads
	proc5 := Process{PID: 500, PPID: 1, Command: "chrome", NumThreads: 5}
	proc6 := Process{PID: 600, PPID: 1, Command: "chrome", NumThreads: 5}

	processes2 := []Process{proc1, proc5, proc6}
	processTree2 := NewProcessTree(0, setupTestLogger(), processes2, DisplayOptions{})

	// Initialize compact mode
	processTree2.InitCompactMode()

	// Test GetProcessCount for a process with threads
	count, _, _, _, _ = processTree.GetProcessCount(1)
	assert.Equal(t, 2, count) // proc5 has one duplicate (proc6)
	// assert.True(t, isThread)  // proc5 is a thread
}

func TestFormatCompactOutput(t *testing.T) {
	// Test formatting for regular processes
	output := FormatCompactOutput("bash", 3, []int32{}, false)
	assert.Equal(t, "3*[bash]", output)

	// Test formatting for a single process (no compaction)
	output = FormatCompactOutput("bash", 1, []int32{}, false)
	assert.Equal(t, "bash", output)

	// Test formatting for threads
	output = FormatCompactOutput("chrome", 4, []int32{}, false)
	assert.Equal(t, "4*[{chrome}]", output)

	// Test formatting for threads when threads are hidden
	output = FormatCompactOutput("chrome", 4, []int32{}, false)
	assert.Equal(t, "", output)

	// Test formatting for processes with full paths
	output = FormatCompactOutput("/usr/bin/bash", 2, []int32{}, false)
	assert.Equal(t, "2*[bash]", output)
}
