// Package pstree provides functionality for building and displaying process trees.
//
// This file contains the implementation of compact mode, which groups identical processes
// in the tree display. It helps reduce visual clutter by showing a count indicator for
// multiple identical processes instead of displaying each one individually.
package pstree

import (
	"fmt"
	"path/filepath"
	"strings"
)

//------------------------------------------------------------------------------
// DATA STRUCTURES
//------------------------------------------------------------------------------

// ProcessGroup represents a group of identical processes
type ProcessGroup struct {
	FirstIndex int    // Index of the first process in the group
	Count      int    // Number of identical processes
	Indices    []int  // Indices of all processes in the group
	FullPath   string // Full path of the command
}

//------------------------------------------------------------------------------
// GLOBAL STATE
//------------------------------------------------------------------------------

// processGroups stores information about groups of identical processes
// Key is the parent PID, value is a map of command -> ProcessGroup
var processGroups map[int32]map[string]ProcessGroup

// skipProcesses tracks which processes should be skipped during printing
var skipProcesses map[int]bool

//------------------------------------------------------------------------------
// INITIALIZATION
//------------------------------------------------------------------------------

// InitCompactMode initializes the compact mode by identifying identical processes.
//
// This function analyzes the provided processes slice and groups processes that have
// identical commands and arguments under the same parent. It populates the processGroups
// map with information about these groups and marks processes that should be skipped
// during printing (all except the first process in each group).
//
// This function should be called before printing the tree when compact mode is enabled.
//
// Parameters:
//   - processes: Slice of Process structs to analyze for grouping
func InitCompactMode(processes []Process) {
	var (
		args      []string
		cmd       string
		exists    bool
		group     ProcessGroup
		parentPID int32
	)

	// Initialize the maps
	processGroups = make(map[int32]map[string]ProcessGroup)
	skipProcesses = make(map[int]bool)

	// Group processes with identical commands under the same parent
	for i := range processes {
		// Skip processes that are already part of a group
		if skipProcesses[i] {
			continue
		}

		// Get parent PID
		parentPID = processes[i].PPID

		// Get the command and arguments to create a composite key
		// This ensures processes are only grouped if both command AND arguments match exactly
		cmd = processes[i].Command
		args = processes[i].Args

		// Create a composite key with both command and arguments
		compositeKey := cmd
		if len(args) > 0 {
			compositeKey = fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))
		}

		// Initialize map for this parent if needed
		if _, exists := processGroups[parentPID]; !exists {
			processGroups[parentPID] = make(map[string]ProcessGroup)
		}

		// Use the composite key (command + args) for grouping
		// This ensures that processes are only grouped if both command AND arguments match exactly
		group, exists = processGroups[parentPID][compositeKey]
		if !exists {
			// Create a new group
			group = ProcessGroup{
				FirstIndex: i,
				Count:      1,
				Indices:    []int{i},
				FullPath:   cmd,
			}
		} else {
			// Add to existing group
			group.Count++
			group.Indices = append(group.Indices, i)

			// Mark this process to be skipped during printing
			skipProcesses[i] = true
		}

		// Update the group in the map
		processGroups[parentPID][compositeKey] = group
	}
}

//------------------------------------------------------------------------------
// PROCESS FILTERING
//------------------------------------------------------------------------------

// ShouldSkipProcess returns true if the process should be skipped during printing.
//
// In compact mode, only the first process of each identical group is displayed,
// with a count indicator. This function checks if a process has been marked to
// be skipped during the initialization phase.
//
// Parameters:
//   - processIndex: Index of the process to check
//
// Returns:
//   - true if the process should be skipped, false otherwise
func ShouldSkipProcess(processIndex int) bool {
	return skipProcesses[processIndex]
}

//------------------------------------------------------------------------------
// PROCESS GROUP INFORMATION
//------------------------------------------------------------------------------

// GetProcessCount returns the count of identical processes for the given process.
//
// For processes that are the first in their group, this returns the total number
// of identical processes in that group. For processes that are not the first in
// their group, or are not part of a group, this returns 1.
//
// Parameters:
//   - processes: Slice of Process structs
//   - processIndex: Index of the process to get the count for
//
// Returns:
//   - count: Number of identical processes in the group
//   - isThread: Whether the process group represents threads
func GetProcessCount(processes []Process, processIndex int) int {
	var (
		args         []string
		cmd          string
		parentPID    int32
		compositeKey string
	)

	// Get parent PID and command
	args = processes[processIndex].Args
	cmd = processes[processIndex].Command
	parentPID = processes[processIndex].PPID

	// Create the same composite key used in InitCompactMode
	compositeKey = cmd
	if len(args) > 0 {
		compositeKey = cmd + " " + strings.Join(args, " ")
	}

	// Check if we have a group for this process
	if groups, exists := processGroups[parentPID]; exists {
		// Look up by composite key (command + args)
		if group, exists := groups[compositeKey]; exists && group.FirstIndex == processIndex {
			return group.Count
		}
	}

	// No group or not the first process in the group
	return 1
}

//------------------------------------------------------------------------------
// OUTPUT FORMATTING
//------------------------------------------------------------------------------

// FormatCompactOutput formats the command with count for compact mode.
//
// This function creates a formatted string representation of a process group
// in the style of Linux pstree. For regular processes, the format is "N*[command]",
// and for threads, the format is "N*[{command}]", where N is the count.
//
// Parameters:
//   - command: The command name to format
//   - count: Number of identical processes/threads
//
// Returns:
//   - Formatted string for display, or empty string if threads should be hidden
func FormatCompactOutput(command string, count int) string {
	if count <= 1 {
		return command
	}

	// Format for processes: N*[command]
	return fmt.Sprintf("%d*[%s]", count, filepath.Base(command))
}
