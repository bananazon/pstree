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
// GLOBAL STATE
//------------------------------------------------------------------------------

// processGroups stores information about groups of identical processes
// Key is the parent PID, value is a map of command -> ProcessGroup
var processGroups map[int32]map[string]map[string]ProcessGroup

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
func InitCompactMode(processes []*Process) {
	var (
		// args         []string
		cmd          string
		exists       bool
		group        ProcessGroup
		parentPID    int32
		processOwner string
	)

	// Initialize the maps
	processGroups = make(map[int32]map[string]map[string]ProcessGroup)
	skipProcesses = make(map[int]bool)

	// Group processes with identical commands under the same parent
	for pidIndex := range processes {
		// Skip processes that are already part of a group
		if skipProcesses[pidIndex] {
			continue
		}

		// Get parent PID
		parentPID = processes[pidIndex].PPID

		// Get the process owner
		processOwner = processes[pidIndex].Username
		compositeKey := processes[pidIndex].Signature

		// Initialize map for this parent if needed
		if _, exists := processGroups[parentPID]; !exists {
			processGroups[parentPID] = make(map[string]map[string]ProcessGroup)
		}

		if _, exists = processGroups[parentPID][compositeKey]; !exists {
			processGroups[parentPID][compositeKey] = make(map[string]ProcessGroup)
		}

		// Use the composite key for grouping
		// This ensures that processes are only grouped if both command AND arguments match exactly
		group, exists = processGroups[parentPID][compositeKey][processOwner]
		if !exists {
			// Create a new group
			group = ProcessGroup{
				Count:      1,
				FirstIndex: pidIndex,
				FullPath:   cmd,
				Indices:    []int{pidIndex},
				Owner:      processOwner,
			}
		} else {
			// Add to existing group
			group.Count++
			group.Indices = append(group.Indices, pidIndex)

			// Mark this process to be skipped during printing
			skipProcesses[pidIndex] = true
		}

		// Update the group in the map
		processGroups[parentPID][compositeKey][processOwner] = group
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
func GetProcessCount(processes []*Process, pidIndex int) (int, []int32) {
	var (
		groupPIDs    []int32
		compositeKey string
		parentPID    int32
		processOwner string
	)

	// Get parent PID and command
	parentPID = processes[pidIndex].PPID
	processOwner = processes[pidIndex].Username
	compositeKey = processes[pidIndex].Signature

	// Check if we have a group for this process
	if groups, exists := processGroups[parentPID]; exists {
		// Look up by owner and composite key (command + args)
		if group, exists := groups[compositeKey][processOwner]; exists && group.FirstIndex == pidIndex {
			// Find PIDs for each member of the group
			for i := range group.Indices {
				groupPIDs = append(groupPIDs, processes[group.Indices[i]].PID)
			}
			return group.Count, groupPIDs
		}
	}

	// No group or not the first process in the group
	return 1, []int32{}
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
func FormatCompactOutput(command string, count int, groupPIDs []int32, showPIDs bool) string {
	if count <= 1 {
		return command
	}

	if showPIDs {
		return fmt.Sprintf("%d*[%s] (%s)", count, filepath.Base(command), strings.Join(PIDsToString(groupPIDs), ","))
	} else {
		return fmt.Sprintf("%d*[%s]", count, filepath.Base(command))
	}
}

// PIDsToString converts a slice of process IDs to a slice of their string representations.
//
// This function is used in compact mode when displaying process groups with PIDs.
// Each PID is converted to a string representation that can be joined together
// for display in the process tree.
//
// Parameters:
//   - pids: Slice of int32 process IDs to convert
//
// Returns:
//   - []string: Slice of string representations of the PIDs
func PIDsToString(pids []int32) []string {
	pidStrings := make([]string, len(pids))
	for i, pid := range pids {
		pidStrings[i] = fmt.Sprintf("%d", pid)
	}
	return pidStrings
}
