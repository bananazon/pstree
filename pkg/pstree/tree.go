// Package pstree provides functionality for building and displaying process trees.
//
// This file contains the core tree building and display logic, including:
// - Process tree construction and hierarchy management
// - Process marking and filtering
// - Tree visualization with various display options
// - Support for different graphical styles (ASCII, UTF-8, etc.)
//
// NOTE: This file has been identified as being too large and complex.
// Functions like BuildTree, MarkProcesses, and PrintTree could benefit from
// being broken down into smaller, more focused functions.
package pstree

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/bananazon/pstree/util"
	"github.com/giancarlosio/gorainbow"
	"github.com/mattn/go-runewidth"
)

var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

//------------------------------------------------------------------------------
// INITIALIZATION AND TREE CONSTRUCTION
//------------------------------------------------------------------------------
// Functions in this section handle the creation of the process tree structure
// and establishing the hierarchical relationships between processes.

// NewProcessTree creates a new process tree from a slice of processes.
//
// This function initializes a ProcessTree structure, populates it with ProcessNode objects
// created from the provided processes, and builds the hierarchical relationships between them.
// The resulting tree can be used for traversal, filtering, and visualization of the process hierarchy.
//
// Parameters:
//   - logger: Logger instance for debug and informational messages
//   - processes: Slice of Process objects containing the process information
//   - displayOptions: Configuration options controlling how the tree will be displayed
//
// Returns:
//   - A pointer to the newly created ProcessTree
func NewProcessTree(debugLevel int, logger *slog.Logger, processes []Process, displayOptions DisplayOptions) (processTree *ProcessTree) {
	processTree = &ProcessTree{
		AtDepth:        0,
		DebugLevel:     debugLevel,
		Logger:         logger,
		Nodes:          make([]*Process, 0, len(processes)),
		PidToIndexMap:  make(map[int32]int, len(processes)),
		IndexToPidMap:  make(map[int]int32, len(processes)),
		RootPID:        displayOptions.RootPID,
		DisplayOptions: displayOptions,
	}

	// Create nodes
	for idx := range processes {
		proc := &processes[idx]
		proc.Children = []*Process{} // initialize
		proc.Signature = ""

		nodeIdx := len(processTree.Nodes)
		processTree.Nodes = append(processTree.Nodes, proc)

		processTree.PidToIndexMap[proc.PID] = nodeIdx
		processTree.IndexToPidMap[idx] = proc.PID
	}

	// Link children (O(n))
	for _, child := range processTree.Nodes {
		if parentIdx, ok := processTree.PidToIndexMap[child.PPID]; ok {
			parent := processTree.Nodes[parentIdx]
			parent.Children = append(parent.Children, child)
		}
	}
	// IMPORTANT: clear cached signatures first
	for _, node := range processTree.Nodes {
		node.Signature = ""
	}

	// Compute subtree signatures for all root processes
	for _, node := range processTree.Nodes {
		if node.PPID == 1 || node.PID == displayOptions.RootPID {
			computeSignature(node, displayOptions.ShowArguments)
		}
	}

	// Find the children for each process
	// for _, proc := range processTree.Nodes {
	// 	if proc.PPID > 1 { // This means there is a parent PID
	// 		parentIndex := processTree.PidToIndexMap[proc.PPID]
	// 		processTree.Nodes[parentIndex].Children = append(processTree.Nodes[parentIndex].Children, proc)
	// 	}
	// }

	// If PID is not set via --pid, we want to look for PID 1...
	// https://github.com/FredHucht/pstree/blob/main/pstree.c#L558-L587

	// Define the tree characters
	if processTree.DisplayOptions.IBM850Graphics {
		processTree.TreeChars = TreeStyles["pc850"]
	} else if processTree.DisplayOptions.UTF8Graphics {
		processTree.TreeChars = TreeStyles["utf8"]
	} else if processTree.DisplayOptions.VT100Graphics {
		processTree.TreeChars = TreeStyles["vt100"]
	} else {
		processTree.TreeChars = TreeStyles["ascii"]
	}

	// Initialize the color scheme
	// if 8 bit color (8-16) is detected, we will use the ansi8 color scheme
	if processTree.DisplayOptions.ColorCount >= 8 && processTree.DisplayOptions.ColorCount <= 16 {
		processTree.ColorScheme = ColorSchemes["ansi8"]
	} else if processTree.DisplayOptions.ColorCount >= 256 {
		if processTree.DisplayOptions.ColorScheme != "" {
			processTree.ColorScheme = ColorSchemes[processTree.DisplayOptions.ColorScheme]
		} else {
			switch runtime.GOOS {
			case "windows":
				if os.Getenv("PSModulePath") != "" {
					processTree.ColorScheme = ColorSchemes["powershell"]
				} else {
					processTree.ColorScheme = ColorSchemes["windows10"]
				}
			case "linux":
				processTree.ColorScheme = ColorSchemes["linux"]
			case "darwin":
				processTree.ColorScheme = ColorSchemes["darwin"]
			default:
				processTree.ColorScheme = ColorSchemes["xterm"]
			}
		}
	}

	// Initialize colorizer
	if processTree.DisplayOptions.ColorizeOutput || processTree.DisplayOptions.ColorAttr != "" {
		if processTree.DisplayOptions.ColorCount >= 8 && processTree.DisplayOptions.ColorCount <= 16 {
			processTree.Colorizer = Colorizers["8color"]
		} else if processTree.DisplayOptions.ColorCount >= 256 {
			processTree.Colorizer = Colorizers["256color"]
		}
	}

	// Build the tree
	processTree.BuildTree()

	// Mark UID transitions
	processTree.MarkUIDTransitions()

	return processTree
}

// BuildTree constructs the hierarchical relationships between processes in the tree.
//
// This method establishes the parent-child relationships between processes by connecting
// each process to its parent based on the PPID (Parent Process ID). It creates a tree structure
// where each node can have one parent and multiple children, with siblings linked in a list.
// The resulting tree structure enables efficient traversal for operations like marking,
// filtering, and visualization.
//
// The method handles cases where a parent process might not exist in the tree (e.g., if the
// parent was not included in the original process list or if it's the process itself).
// BuildTree constructs the hierarchical relationships between processes in the tree.
//
// This method establishes the parent-child relationships between processes by connecting
// each process to its parent based on the PPID (Parent Process ID). It creates a tree structure
// where each node can have one parent and multiple children, with siblings linked in a list.
// The resulting tree structure enables efficient traversal for operations like marking,
// filtering, and visualization.
//
// The method handles cases where a parent process might not exist in the tree (e.g., if the
// parent was not included in the original process list or if it's the process itself).
//
// Refactoring opportunity: This function could be broken down into smaller functions:
// - initializeNodes: Initialize all nodes with default values
// - buildParentChildRelationships: Establish the parent-child connections
func (processTree *ProcessTree) BuildTree() {
	// https://github.com/FredHucht/pstree/blob/main/pstree.c#L635-L652
	processTree.Logger.Debug("Entering processTree.BuildTree()")

	// Initialize all nodes with -1 for Child, Parent, and Sister fields
	for i := range processTree.Nodes {
		processTree.Nodes[i].Child = -1
		processTree.Nodes[i].Parent = -1
		processTree.Nodes[i].Sister = -1
		processTree.Nodes[i].Print = false
	}

	// Build the tree using the PidToIndexMap for O(1) lookups
	for pidIndex := range processTree.Nodes {
		ppid := processTree.Nodes[pidIndex].PPID

		// Look up parent index directly from the map
		ppidIndex, exists := processTree.PidToIndexMap[ppid]

		// Skip if parent doesn't exist or is the process itself
		if !exists || ppidIndex == pidIndex {
			continue
		}

		// Set parent relationship
		processTree.Nodes[pidIndex].Parent = ppidIndex

		// Add as child
		if processTree.Nodes[ppidIndex].Child == -1 {
			// First child
			processTree.Nodes[ppidIndex].Child = pidIndex
		} else {
			// Find the last sibling
			sisterIndex := processTree.Nodes[ppidIndex].Child
			for processTree.Nodes[sisterIndex].Sister != -1 {
				sisterIndex = processTree.Nodes[sisterIndex].Sister
			}
			// Add as sister to the last child
			processTree.Nodes[sisterIndex].Sister = pidIndex
		}
	}
}

//------------------------------------------------------------------------------
// PROCESS MARKING AND FILTERING
//------------------------------------------------------------------------------
// Functions in this section handle the identification and marking of processes
// that should be included in the display, based on various filtering criteria.

// MarkProcesses marks processes that should be displayed based on filtering criteria.
// It applies various filters such as process name pattern matching, username filtering,
// root process exclusion, and PID filtering to determine which processes should be displayed.
// MarkProcesses marks processes that should be displayed based on filtering criteria.
// It applies various filters such as process name pattern matching, username filtering,
// root process exclusion, and PID filtering to determine which processes should be displayed.
//
// Refactoring opportunity: This function could be broken down into smaller functions:
// - applyUsernameFilter: Mark processes matching username criteria
// - applyRootPIDFilter: Mark processes based on root PID
// - applyCommandFilter: Mark processes matching command pattern
// - applyRootExclusionFilter: Apply root user exclusion filter
func (processTree *ProcessTree) MarkProcesses() {
	// https://github.com/FredHucht/pstree/blob/main/pstree.c#L662-L684
	processTree.Logger.Debug("Entering processTree.MarkProcesses()")
	var (
		myPid    int32
		process  Process
		pidIndex int
		showAll  bool
		username string
	)

	if processTree.DisplayOptions.Contains == "" && len(processTree.DisplayOptions.Usernames) == 0 && !processTree.DisplayOptions.ExcludeRoot && processTree.DisplayOptions.RootPID < 1 {
		showAll = true
	}

	for pidIndex = range processTree.Nodes {
		if showAll {
			processTree.Nodes[pidIndex].Print = true
		} else {
			process = *processTree.Nodes[pidIndex]
			if len(processTree.DisplayOptions.Usernames) > 0 {
				for _, username = range processTree.DisplayOptions.Usernames {
					if process.Username == username {
						processTree.markParents(pidIndex)
						processTree.markChildren(pidIndex)
					}
				}
			} else if processTree.Nodes[pidIndex].PID == processTree.DisplayOptions.RootPID {
				// processTree.Logger.Debug("--pid == processTree.DisplayOptions.RootPID")
				if (processTree.DisplayOptions.ExcludeRoot && processTree.Nodes[pidIndex].Username != "root") || (!processTree.DisplayOptions.ExcludeRoot) {
					// processTree.Logger.Debug("(processTree.DisplayOptions.ExcludeRoot && processTree.Nodes[pidIndex].Username != root) || !processTree.DisplayOptions.ExcludeRoot")
					processTree.markParents(pidIndex)
					processTree.markChildren(pidIndex)
				}
			} else if processTree.DisplayOptions.Contains != "" && strings.Contains(process.Command, processTree.DisplayOptions.Contains) && (process.PID != myPid) {
				// processTree.Logger.Debug("processTree.DisplayOptions.Contains is set && process.Command contains processTree.DisplayOptions.Contains && process.PID != myPid")
				if (processTree.DisplayOptions.ExcludeRoot && process.Username != "root") || (!processTree.DisplayOptions.ExcludeRoot) {
					// processTree.Logger.Debug("(processTree.DisplayOptions.ExcludeRoot && process.Username != root) || !processTree.DisplayOptions.ExcludeRoot")
					processTree.markParents(pidIndex)
					processTree.markChildren(pidIndex)
				}
			} else if processTree.DisplayOptions.Contains != "" && !strings.Contains(process.Command, processTree.DisplayOptions.Contains) && (process.PID != myPid) {
				// processTree.Logger.Debug("processTree.DisplayOptions.Contains is set && process.Command does not contain processTree.DisplayOptions.Contains && process.PID != myPid")
			} else if processTree.DisplayOptions.ExcludeRoot && process.Username != "root" {
				// processTree.Logger.Debug("processTree.DisplayOptions.ExcludeRoot && process.Username != root")
				processTree.markParents(pidIndex)
				processTree.markChildren(pidIndex)
			}
		}
	}
}

// DropUnmarked removes processes that are not marked for display from the process tree.
// It modifies the process tree structure to maintain proper parent-child relationships
// while excluding processes that should not be displayed.
// DropUnmarked removes processes that are not marked for display from the process tree.
// It modifies the process tree structure to maintain proper parent-child relationships
// while excluding processes that should not be displayed.
//
// Refactoring opportunity: This function could be split into:
// - dropUnmarkedChildren: Remove unmarked children from each node
// - dropUnmarkedSiblings: Remove unmarked siblings from each node
func (processTree *ProcessTree) DropUnmarked() {
	// https://github.com/FredHucht/pstree/blob/main/pstree.c#L706-L717
	processTree.Logger.Debug("Entering processTree.DropUnmarked()")
	var (
		childPidIndex  int
		pidIndex       int
		sisterPidIndex int
	)

	for pidIndex = range processTree.Nodes {
		if processTree.Nodes[pidIndex].Print {
			// Drop children that won't print
			childPidIndex = processTree.Nodes[pidIndex].Child
			for childPidIndex != -1 && !processTree.Nodes[childPidIndex].Print {
				childPidIndex = processTree.Nodes[childPidIndex].Sister
			}
			processTree.Nodes[pidIndex].Child = childPidIndex

			// Drop sisters that won't print
			sisterPidIndex = processTree.Nodes[pidIndex].Sister
			for sisterPidIndex != -1 && !processTree.Nodes[sisterPidIndex].Print {
				sisterPidIndex = processTree.Nodes[sisterPidIndex].Sister
			}
			processTree.Nodes[pidIndex].Sister = sisterPidIndex
		}
	}
}

//------------------------------------------------------------------------------
// PROCESS ATTRIBUTE MARKING
//------------------------------------------------------------------------------
// Functions in this section handle marking special attributes of processes,
// such as UID transitions and highlighting the current process.

// MarkUIDTransitions identifies and marks processes where the user ID changes from the parent process.
// This function compares the UIDs of each process with its parent and sets HasUIDTransition=true
// when a transition is detected. It also stores the parent UID for display purposes.
//
// Parameters:
//   - logger: Logger instance for debug information
//   - processes: Pointer to a slice of Process structs
func (processTree *ProcessTree) MarkUIDTransitions() {
	var (
		pidIndex  int
		ppidIndex int
	)

	processTree.Logger.Debug("Marking UID transitions between processes - START")

	for pidIndex = range processTree.Nodes {
		// Skip the root process (which has no parent)
		if processTree.Nodes[pidIndex].Parent == -1 {
			continue
		}

		// Get parent index
		ppidIndex = processTree.Nodes[pidIndex].Parent

		// Compare UIDs between process and its parent
		if len(processTree.Nodes[pidIndex].UIDs) > 0 && len(processTree.Nodes[ppidIndex].UIDs) > 0 {
			// Store parent UID regardless of transition
			processTree.Nodes[pidIndex].ParentUID = processTree.Nodes[ppidIndex].UIDs[0]
			processTree.Nodes[pidIndex].ParentUsername = processTree.Nodes[ppidIndex].Username

			// Compare the first UID (effective UID
			if processTree.Nodes[pidIndex].UIDs[0] != processTree.Nodes[ppidIndex].UIDs[0] {
				if processTree.DebugLevel > 1 {
					processTree.Logger.Debug(fmt.Sprintf("UID transition detected: Process %d (UID %d) has different UID from parent %d (UID %d)",
						processTree.Nodes[pidIndex].PID, processTree.Nodes[pidIndex].UIDs[0],
						processTree.Nodes[ppidIndex].PID, processTree.Nodes[ppidIndex].UIDs[0]))
				}
				processTree.Nodes[pidIndex].HasUIDTransition = true
			}
		} else {
			// Fallback to username comparison if UIDs are not available
			if processTree.Nodes[pidIndex].Username != processTree.Nodes[ppidIndex].Username {
				if processTree.DebugLevel > 1 {
					processTree.Logger.Debug(fmt.Sprintf("Username transition detected: Process %d (%s) has different username from parent %d (%s)",
						processTree.Nodes[pidIndex].PID, processTree.Nodes[pidIndex].Username,
						processTree.Nodes[ppidIndex].PID, processTree.Nodes[ppidIndex].Username))
				}
				processTree.Nodes[pidIndex].HasUIDTransition = true
			}
		}
	}
}

// MarkCurrentAndAncestors marks the current process and all its ancestors.
// This function identifies the current process by its PID and marks it and all
// its ancestors with IsCurrentOrAncestor=true for highlighting in the display.
//
// Parameters:
//   - currentPid: The PID of the current process to highlight
func (processTree *ProcessTree) MarkCurrentAndAncestors(currentPid int32) {
	if currentPid <= 0 {
		return
	}

	processTree.Logger.Debug(fmt.Sprintf("Marking current process %d and its ancestors", currentPid))
	var (
		currentIndex int
		parentIndex  int
	)

	// Find the current process index
	currentIndex = processTree.getPidIndex(currentPid)
	if currentIndex == -1 {
		processTree.Logger.Debug(fmt.Sprintf("Current process %d not found", currentPid))
		return
	}

	// Mark the current process
	processTree.Nodes[currentIndex].IsCurrentOrAncestor = true

	// Mark all ancestors

	parentIndex = processTree.Nodes[currentIndex].Parent
	for parentIndex != -1 {
		processTree.Logger.Debug(fmt.Sprintf("Marking pid %d as ancestor of current process", processTree.IndexToPidMap[parentIndex]))
		processTree.Nodes[parentIndex].IsCurrentOrAncestor = true
		parentIndex = processTree.Nodes[parentIndex].Parent
	}
}

//------------------------------------------------------------------------------
// DEBUGGING UTILITIES
//------------------------------------------------------------------------------
// Functions in this section provide debugging capabilities for the process tree.

// ShowPrintable logs all processes that are marked for display.
//
// This method is primarily used for debugging purposes. It iterates through all nodes
// in the process tree and logs detailed information about each process that has been
// marked for display (Print=true). The output includes all fields of the Process struct.
//
// The method uses pretty-printing to format the output in a readable way.
func (processTree *ProcessTree) ShowPrintable() {
	for i := range processTree.Nodes {
		if processTree.Nodes[i].Print {
			fmt.Printf("PID %d is printable\n", processTree.IndexToPidMap[i])
		}
	}
}

//------------------------------------------------------------------------------
// TREE VISUALIZATION AND FORMATTING
//------------------------------------------------------------------------------
// Functions in this section handle the visual representation of the process tree,
// including line formatting, tree branch construction, and display options.

// buildLinePrefix constructs the tree visualization prefix for a process node in the tree display.
// It creates the branch connectors (├, └, etc.) that show the hierarchical relationship between processes.
//
// Parameters:
//   - head: The accumulated prefix string from parent levels
//   - pidIndex: Index of the current process in the Nodes array
//
// Returns a formatted string containing tree branch characters that represent the process's position in the hierarchy.
// buildLinePrefix constructs the tree visualization prefix for a process node in the tree display.
// It creates the branch connectors (├, └, etc.) that show the hierarchical relationship between processes.
//
// Parameters:
//   - head: The accumulated prefix string from parent levels
//   - pidIndex: Index of the current process in the Nodes array
//
// Returns:
//   - A formatted string containing tree branch characters that represent the process's position in the hierarchy
//
// Refactoring opportunity: This function is complex and could be broken down into:
// - determineNodePosition: Determine if node is last child, has siblings, etc.
// - selectBranchCharacters: Select appropriate branch characters based on position
// - formatPrefix: Format the final prefix string
func (processTree *ProcessTree) buildLinePrefix(head string, pidIndex int) string {
	processTree.Logger.Debug(fmt.Sprintf("processTree.buildLinePrefix(head=\"%s\", pidIndex=%d, atDepth=%d)", head, pidIndex, processTree.AtDepth))

	// Create a strings.Builder with an estimated capacity
	// This helps avoid reallocations as the builder grows
	var builder strings.Builder

	// Pre-allocate capacity based on expected size
	// This is an optimization to avoid reallocations
	// You can adjust the capacity based on typical usage patterns
	builder.Grow(len(head) + 50) // Estimate based on typical usage

	// Append initialization sequences
	builder.WriteString(processTree.TreeChars.Init)
	builder.WriteString(processTree.TreeChars.SG)
	builder.WriteString(head)

	if processTree.Nodes[pidIndex].PID == 1 {
		// This is a worakround
		builder.WriteString(processTree.TreeChars.P)
		if processTree.DisplayOptions.ShowPGLs {
			builder.WriteString(processTree.TreeChars.PGL)
		} else {
			builder.WriteString(processTree.TreeChars.NPGL)
		}
		builder.WriteString(processTree.TreeChars.EG)
		return builder.String()
	}

	if head == "" {
		return ""
	} else {
		// Check if this process has a visible sibling
		hasVisibleSibling := false
		sibling := processTree.Nodes[pidIndex].Sister

		// In compact mode, we need to check if all siblings are going to be skipped
		if processTree.DisplayOptions.CompactMode {
			for sibling != -1 {
				if !ShouldSkipProcess(sibling) {
					hasVisibleSibling = true
					break
				}
				sibling = processTree.Nodes[sibling].Sister
			}
		} else {
			// In normal mode, just check if there's a sibling
			hasVisibleSibling = (sibling != -1)
		}

		if hasVisibleSibling {
			builder.WriteString(processTree.TreeChars.BarC) // T-connector for processes with visible siblings
		} else {
			builder.WriteString(processTree.TreeChars.BarL) // L-connector for processes without visible siblings (last child)
		}
	}

	if processTree.Nodes[pidIndex].Child != -1 && processTree.AtDepth < processTree.DisplayOptions.MaxDepth {
		builder.WriteString(processTree.TreeChars.P)
	} else {
		builder.WriteString(processTree.TreeChars.S2)
	}

	if processTree.Nodes[pidIndex].PID == processTree.Nodes[pidIndex].PGID {
		if !processTree.DisplayOptions.ShowPGLs {
			builder.WriteString(processTree.TreeChars.NPGL)
		} else {
			builder.WriteString(processTree.TreeChars.PGL)
		}
	} else {
		builder.WriteString(processTree.TreeChars.NPGL)
	}
	builder.WriteString(processTree.TreeChars.EG)

	// Return the completed string
	return builder.String()
}

// buildLineItem constructs a complete formatted line for a process in the tree display.
// It combines the tree structure prefix with various process information based on display options.
//
// Parameters:
//   - head: The accumulated prefix string from parent levels
//   - pidIndex: Index of the current process in the Nodes array
//
// Returns a fully formatted string containing the process information with appropriate formatting and coloring.
// The string includes elements such as tree structure, process IDs, resource usage, and command information
// based on the configured display options.
// buildLineItem constructs a complete formatted line for a process in the tree display.
// It combines the tree structure prefix with various process information based on display options.
//
// Parameters:
//   - head: The accumulated prefix string from parent levels
//   - pidIndex: Index of the current process in the Nodes array
//
// Returns:
//   - A fully formatted string containing the process information with appropriate formatting and coloring.
//     The string includes elements such as tree structure, process IDs, resource usage, and command information
//     based on the configured display options.
//
// Refactoring opportunity: This function is very large and could be broken down into:
// - formatProcessIDs: Format PID, PPID, PGID information
// - formatResourceUsage: Format CPU, memory, thread information
// - formatCommandInfo: Format command and arguments
// - formatOwnerInfo: Format username and UID transition information
func (processTree *ProcessTree) buildLineItem(head string, pidIndex int) string {
	processTree.Logger.Debug(fmt.Sprintf("processTree.buildLineItem(head=\"%s\", pidIndex=%d, atDepth=%d)", head, pidIndex, processTree.AtDepth))
	var (
		ageString       string
		args            string
		commandStr      string
		compactStr      string
		connector       string
		cpuPercent      string
		linePrefix      string
		memoryUsage     string
		owner           string
		ownerTransition string
		pgidString      string
		pidPgidSlice    []string
		pidPgidString   string
		pidString       string
		ppidString      string
		threads         string
	)

	// Create a strings.Builder with an estimated capacity
	// This helps avoid reallocations as the builder grows
	var builder strings.Builder

	// Pre-allocate capacity based on expected size
	// This is an optimization to avoid reallocations
	// You can adjust the capacity based on typical usage patterns
	builder.Grow(len(head) + 260) // Estimate based on typical usage

	linePrefix = processTree.buildLinePrefix(head, pidIndex)
	processTree.colorizeField("prefix", &linePrefix, pidIndex)

	builder.WriteString(linePrefix)
	builder.WriteString(" ")

	if processTree.DisplayOptions.ShowPIDs {
		pidString = util.Int32toStr(processTree.Nodes[pidIndex].PID)
		pidPgidSlice = append(pidPgidSlice, pidString)
	}

	if processTree.DisplayOptions.ShowPPIDs {
		ppidString = util.Int32toStr(processTree.Nodes[pidIndex].PPID)
		pidPgidSlice = append(pidPgidSlice, ppidString)
	}

	if processTree.DisplayOptions.ShowPGIDs {
		pgidString = util.Int32toStr(processTree.Nodes[pidIndex].PGID)
		pidPgidSlice = append(pidPgidSlice, pgidString)
	}

	if len(pidPgidSlice) > 0 {
		pidPgidString = fmt.Sprintf("(%s)", strings.Join(pidPgidSlice, ","))
		processTree.colorizeField("pidPgid", &pidPgidString, pidIndex)
		builder.WriteString(pidPgidString)
		builder.WriteString(" ")
	}

	if processTree.DisplayOptions.ShowOwner {
		owner = processTree.Nodes[pidIndex].Username
		processTree.colorizeField("owner", &owner, pidIndex)
		builder.WriteString(owner)
		builder.WriteString(" ")
	}

	if processTree.DisplayOptions.ShowProcessAge {
		duration := util.FindDuration(processTree.Nodes[pidIndex].Age)
		ageSlice := []string{}
		ageSlice = append(ageSlice, fmt.Sprintf("%02d", duration.Days))
		ageSlice = append(ageSlice, fmt.Sprintf("%02d", duration.Hours))
		ageSlice = append(ageSlice, fmt.Sprintf("%02d", duration.Minutes))
		ageSlice = append(ageSlice, fmt.Sprintf("%02d", duration.Seconds))
		ageString = fmt.Sprintf(
			"(%s)",
			strings.Join(ageSlice, ":"),
		)
		processTree.colorizeField("age", &ageString, pidIndex)
		builder.WriteString(ageString)
		builder.WriteString(" ")
	}

	if processTree.DisplayOptions.ShowCpuPercent {
		cpuPercent = fmt.Sprintf("(c:%.2f%%)", processTree.Nodes[pidIndex].CPUPercent)
		processTree.colorizeField("cpu", &cpuPercent, pidIndex)
		builder.WriteString(cpuPercent)
		builder.WriteString(" ")
	}

	if processTree.DisplayOptions.ShowMemoryUsage {
		memoryUsage = fmt.Sprintf("(m:%s)", util.ByteConverter(processTree.Nodes[pidIndex].MemoryInfo.RSS))
		processTree.colorizeField("memory", &memoryUsage, pidIndex)
		builder.WriteString(memoryUsage)
		builder.WriteString(" ")
	}

	if processTree.DisplayOptions.ShowNumThreads {
		// Always show thread count, even when showing compact format
		threads = fmt.Sprintf("(t:%d)", processTree.Nodes[pidIndex].NumThreads)
		processTree.colorizeField("threads", &threads, pidIndex)
		builder.WriteString(threads)
		builder.WriteString(" ")
	}

	if processTree.DisplayOptions.ShowUIDTransitions && processTree.Nodes[pidIndex].HasUIDTransition {
		// Add UID transition notation {parentUID→currentUID}
		if len(processTree.Nodes[pidIndex].UIDs) > 0 {
			ownerTransition = fmt.Sprintf("(%d→%d)", processTree.Nodes[pidIndex].ParentUID, processTree.Nodes[pidIndex].UIDs[0])
		}
	} else if processTree.DisplayOptions.ShowUserTransitions && processTree.Nodes[pidIndex].HasUIDTransition {
		// Add user transition notation {parentUser→currentUser}
		if processTree.Nodes[pidIndex].ParentUsername != "" {
			ownerTransition = fmt.Sprintf("(%s→%s)", processTree.Nodes[pidIndex].ParentUsername, processTree.Nodes[pidIndex].Username)
		}
	}

	if ownerTransition != "" {
		processTree.colorizeField("ownerTransition", &ownerTransition, pidIndex)
		builder.WriteString(ownerTransition)
		builder.WriteString(" ")
	}

	// Get the command - use full path when compact mode is disabled
	commandStr = processTree.Nodes[pidIndex].Command

	// In compact mode, format the command with count for the first process in a group
	if processTree.DisplayOptions.CompactMode {
		// Get the count of identical processes
		count, groupPIDs := GetProcessCount(processTree.Nodes, pidIndex)

		// If there are multiple identical processes, format with count
		if count > 1 {
			// Format in Linux pstree style
			compactStr = FormatCompactOutput(commandStr, count, groupPIDs, processTree.DisplayOptions.ShowPIDs)

			if compactStr != "" {
				// Create the connector string
				connector = "───"

				// Colorize the connector and compact format indicator in green if color support is available
				if processTree.DisplayOptions.ColorSupport {
					// The util.ColorGreen function modifies the string in place via pointer
					processTree.colorizeField("connector", &connector, pidIndex)
					// builder.WriteString(connector)
					processTree.colorizeField("compactStr", &compactStr, pidIndex)
					// builder.WriteString(compactStr)
				}

				// In Linux pstree, the format is just the count and brackets, not repeating the command
				commandStr = fmt.Sprintf("%s%s%s", commandStr, connector, compactStr)
			}
		}
	}

	// For threads in non-compact mode, wrap the command in curly braces
	// if !processTree.DisplayOptions.CompactMode && isThread && processTree.Nodes[pidIndex].NumThreads == 0 {
	// 	// This is likely a thread of a process
	// 	commandStr = fmt.Sprintf("{%s}", commandStr)
	// }

	processTree.colorizeField("command", &commandStr, pidIndex)
	builder.WriteString(commandStr)
	builder.WriteString(" ")

	if processTree.DisplayOptions.ShowArguments {
		if len(processTree.Nodes[pidIndex].Args) > 0 {
			args = strings.Join(processTree.Nodes[pidIndex].Args, " ")
			processTree.colorizeField("args", &args, pidIndex)
			builder.WriteString(args)
			builder.WriteString(" ")
		}
	}

	return builder.String()
}

// buildNewHead constructs a new head string for child processes based on the current process's position.
//
// Parameters:
//   - head: The accumulated prefix string from parent levels
//   - pidIndex: Index of the current process in the Nodes array
//
// Returns a string to be used as the head for child processes, including appropriate vertical bars
// or spaces based on whether the current process has visible siblings.
// buildNewHead constructs a new head string for child processes based on the current process's position.
//
// Parameters:
//   - head: The accumulated prefix string from parent levels
//   - pidIndex: Index of the current process in the Nodes array
//
// Returns:
//   - A string to be used as the head for child processes, including appropriate vertical bars
//     or spaces based on whether the current process has visible siblings.
func (processTree *ProcessTree) buildNewHead(head string, pidIndex int) string {
	newHead := fmt.Sprintf("%s%s ", head,
		func() string {
			if head == "" {
				return ""
			}
			// In compact mode, we need to check if any visible siblings exist
			if processTree.DisplayOptions.CompactMode {
				sibling := processTree.Nodes[pidIndex].Sister
				for sibling != -1 {
					if !ShouldSkipProcess(sibling) {
						return processTree.TreeChars.Bar // Only add vertical bar if there's a visible sibling
					}
					sibling = processTree.Nodes[sibling].Sister
				}
				return " " // No visible siblings
			} else {
				// In normal mode, just check if there's a sibling
				if processTree.Nodes[pidIndex].Sister != -1 {
					return processTree.TreeChars.Bar
				}
				return " "
			}
		}(),
	)
	return newHead
}

//------------------------------------------------------------------------------
// TREE TRAVERSAL AND DISPLAY
//------------------------------------------------------------------------------
// Functions in this section handle the recursive traversal of the process tree
// and the display of processes with their relationships.

// PrintTree recursively prints a process tree with customizable formatting options.
//
// This function displays a process and all its children in a tree-like structure,
// with various display options such as process age, CPU usage, memory usage, etc.
// The tree is formatted using different graphical styles based on the display options.
//
// Parameters:
//   - pidIndex: Index of the current process to print
//   - head: String representing the indentation and tree structure for the current line
//
// PrintTree recursively prints a process tree with customizable formatting options.
//
// This function displays a process and all its children in a tree-like structure,
// with various display options such as process age, CPU usage, memory usage, etc.
// The tree is formatted using different graphical styles based on the display options.
//
// Parameters:
//   - pidIndex: Index of the current process to print
//   - head: String representing the indentation and tree structure for the current line
//
// Refactoring opportunity: This function could be split into:
// - printCurrentNode: Print just the current node
// - printChildNodes: Handle the recursive printing of child nodes
func (processTree *ProcessTree) PrintTree(pidIndex int, head string) {
	processTree.Logger.Debug(fmt.Sprintf("Entering processTree.PrintTree() with %d nodes", len(processTree.Nodes)))
	processTree.Logger.Debug(fmt.Sprintf("processTree.PrintTree(pidIndex=%d, head=\"%s\", atDepth=%d)", pidIndex, head, processTree.AtDepth))
	// https://github.com/FredHucht/pstree/blob/main/pstree.c#L721-L777
	// Skip if we've reached the maximum depth
	if processTree.DisplayOptions.MaxDepth > 0 && processTree.AtDepth > processTree.DisplayOptions.MaxDepth {
		processTree.Logger.Debug(fmt.Sprintf("Skipping process %d at depth %d (max depth %d)", processTree.Nodes[pidIndex].PID, processTree.AtDepth, processTree.DisplayOptions.MaxDepth))
		return
	}

	// Initialize compact mode if enabled and at the root level
	if processTree.AtDepth == 0 {
		// Always initialize compact mode to identify duplicates
		// But we'll respect the CompactMode flag when displaying
		processTree.Logger.Debug("Initializing compact mode")
		InitCompactMode(processTree.Nodes, &processTree.DisplayOptions)
	}

	// Skip this process if it's been marked as a duplicate in compact mode
	// Only skip if compact mode is actually enabled
	if processTree.DisplayOptions.CompactMode && ShouldSkipProcess(pidIndex) {
		processTree.Logger.Debug(fmt.Sprintf("Skipping process %d in compact mode", processTree.Nodes[pidIndex].PID))
		return
	}

	var (
		line    string
		newHead string
	)

	if processTree.AtDepth > processTree.DisplayOptions.MaxDepth {
		processTree.Logger.Debug(fmt.Sprintf("Skipping process %d at depth %d (max depth %d)", processTree.Nodes[pidIndex].PID, processTree.AtDepth, processTree.DisplayOptions.MaxDepth))
		return
	}

	if head == "" && !processTree.Nodes[pidIndex].Print {
		processTree.Logger.Debug(fmt.Sprintf("Skipping process %d because head is empty and Print is false", processTree.Nodes[pidIndex].PID))
		return
	}

	line = processTree.buildLineItem(head, pidIndex)

	if !processTree.DisplayOptions.WideDisplay {
		if len(line) > processTree.DisplayOptions.ScreenWidth {
			if processTree.DisplayOptions.RainbowOutput {
				line = processTree.truncateANSI(gorainbow.Rainbow(line))
			} else {
				line = processTree.truncateANSI(line)
			}
		} else {
			if processTree.DisplayOptions.RainbowOutput {
				line = gorainbow.Rainbow(line)
			}
		}
	} else {
		if processTree.DisplayOptions.RainbowOutput {
			line = gorainbow.Rainbow(line)
		}
	}

	newHead = processTree.buildNewHead(head, pidIndex)

	processTree.Logger.Debug(fmt.Sprintf("processTree.PrintTree(): printing line for node.PID=%d, head=\"%s\"", processTree.Nodes[pidIndex].PID, head))
	fmt.Fprintln(os.Stdout, line)

	// Iterate over children and determine sibling status
	childme := processTree.Nodes[pidIndex].Child
	for childme != -1 {
		nextChild := processTree.Nodes[childme].Sister
		processTree.AtDepth++
		processTree.PrintTree(childme, newHead)
		processTree.AtDepth--
		childme = nextChild
	}
}

//------------------------------------------------------------------------------
// TREE TRAVERSAL HELPERS
//------------------------------------------------------------------------------
// Helper functions for traversing the process tree in different directions.

// markParents marks all parent processes of a given process as printable.
// This function recursively traverses up the process tree, marking each parent
// process with Print=true until it reaches the root process (or a process with no parent).
//
// Parameters:
//   - pidIndex: Index of the process whose parents should be marked
//
// markParents marks all parent processes of a given process as printable.
// This function recursively traverses up the process tree, marking each parent
// process with Print=true until it reaches the root process (or a process with no parent).
//
// Parameters:
//   - pidIndex: Index of the process whose parents should be marked
func (processTree *ProcessTree) markParents(pidIndex int) {
	processTree.Logger.Debug(fmt.Sprintf("Entering markParents(), pidIndex=%d, pid=%d", pidIndex, processTree.IndexToPidMap[pidIndex]))
	var (
		ppidIndex int
	)

	ppidIndex = processTree.Nodes[pidIndex].Parent
	processTree.Logger.Debug(fmt.Sprintf("Marking %d as a parent of %d", processTree.IndexToPidMap[ppidIndex], processTree.IndexToPidMap[pidIndex]))
	for ppidIndex != -1 {
		processTree.Logger.Debug(fmt.Sprintf("Marking PID %d's Print attribute as true", processTree.IndexToPidMap[ppidIndex]))
		processTree.Nodes[ppidIndex].Print = true
		ppidIndex = processTree.Nodes[ppidIndex].Parent
	}
}

// markChildren marks a process and all its child processes as printable.
// This function recursively traverses down the process tree, marking each child
// process with Print=true, and continues with any sibling processes.
//
// Parameters:
//   - pidIndex: Index of the process whose children should be marked
//
// markChildren marks a process and all its child processes as printable.
// This function recursively traverses down the process tree, marking each child
// process with Print=true, and continues with any sibling processes.
//
// Parameters:
//   - pidIndex: Index of the process whose children should be marked
func (processTree *ProcessTree) markChildren(pidIndex int) {
	processTree.Logger.Debug(fmt.Sprintf("Entering markChildren(), pidIndex=%d, pid=%d", pidIndex, processTree.IndexToPidMap[pidIndex]))
	var (
		childPidIndex int
	)

	processTree.Logger.Debug(fmt.Sprintf("Marking PID %d's Print attribute as true", processTree.IndexToPidMap[pidIndex]))
	processTree.Nodes[pidIndex].Print = true
	childPidIndex = processTree.Nodes[pidIndex].Child
	for childPidIndex != -1 {
		processTree.markChildren(childPidIndex)
		childPidIndex = processTree.Nodes[childPidIndex].Sister
	}
}

//------------------------------------------------------------------------------
// DISPLAY FORMATTING AND STYLING
//------------------------------------------------------------------------------
// Functions in this section handle the visual styling of the process tree,
// including colorization, width calculation, and text truncation.

// ColorizeField applies appropriate color formatting to a specific field in the process tree output.
//
// This method enhances the visual representation of the process tree by applying colors
// to different elements based on the current display options. It supports two main coloring modes:
//
// 1. Standard colorization (--colorize flag): Each field type gets a predefined color
//
//   - Username: Blue
//
//   - Command: Blue
//
//   - Arguments: Red
//
//   - PID/PGID: Purple
//
//   - CPU: Yellow
//
//   - Memory: Orange
//
//   - Age: Bold Green
//
//   - Threads: Bold White
//
//   - Tree characters: Green
//
//     2. Attribute-based colorization (--color flag): Colors are applied based on process attributes
//     like CPU or memory usage, with thresholds determining the color (green/yellow/red)
//
// The colors help to quickly identify important information in the tree, such as high
// resource usage processes or specific elements of interest.
//
// Parameters:
//   - fieldName: String identifying which field is being colored (e.g., "cpu", "memory", "command")
//   - value: Pointer to the string value to be colored (modified in-place)
//   - pidIndex: Index of the process to be colored
//
// The method uses the hybrid approach data (combining gopsutil with direct ps command calls)
// when applying attribute-based coloring, ensuring accurate thresholds for CPU and memory usage.
// colorizeField applies appropriate color formatting to a specific field in the process tree output.
//
// This method enhances the visual representation of the process tree by applying colors
// to different elements based on the current display options. It supports two main coloring modes:
//
//  1. Standard colorization (--colorize flag): Each field type gets a predefined color
//  2. Attribute-based colorization (--color flag): Colors are applied based on process attributes
//     like CPU or memory usage, with thresholds determining the color (green/yellow/red)
//
// Parameters:
//   - fieldName: String identifying which field is being colored (e.g., "cpu", "memory", "command")
//   - value: Pointer to the string value to be colored (modified in-place)
//   - pidIndex: Index of the process to be colored
//
// Refactoring opportunity: This function could be split into:
// - applyStandardColors: Apply standard color scheme
// - applyAttributeBasedColors: Apply colors based on attribute thresholds
func (processTree *ProcessTree) colorizeField(fieldName string, value *string, pidIndex int) {
	var (
		process *Process
	)
	// Only apply colors if the terminal supports them
	if processTree.DisplayOptions.ColorSupport {
		// Standard colorization mode (--colorize flag)
		if processTree.DisplayOptions.ColorizeOutput {
			// Apply specific colors based on the field type
			switch fieldName {
			case "age":
				processTree.Colorizer.Age(processTree.ColorScheme, value)
			case "args":
				processTree.Colorizer.Args(processTree.ColorScheme, value)
			case "connector":
				processTree.Colorizer.Connector(processTree.ColorScheme, value)
			case "command":
				processTree.Colorizer.Command(processTree.ColorScheme, value)
			case "compactStr":
				processTree.Colorizer.CompactStr(processTree.ColorScheme, value)
			case "cpu":
				processTree.Colorizer.CPU(processTree.ColorScheme, value)
			case "memory":
				processTree.Colorizer.Memory(processTree.ColorScheme, value)
			case "owner":
				processTree.Colorizer.Owner(processTree.ColorScheme, value)
			case "ownerTransition":
				processTree.Colorizer.OwnerTransition(processTree.ColorScheme, value)
			case "pidPgid":
				processTree.Colorizer.PIDPGID(processTree.ColorScheme, value)
			case "prefix":
				processTree.Colorizer.Prefix(processTree.ColorScheme, value)
			case "threads":
				processTree.Colorizer.NumThreads(processTree.ColorScheme, value)
			}
		} else if processTree.DisplayOptions.ColorAttr != "" {
			// Attribute-based colorization mode (--color flag)
			// Don't apply attribute-based coloring to the tree prefix
			if fieldName != "prefix" {
				process = processTree.Nodes[pidIndex]
				switch processTree.DisplayOptions.ColorAttr {
				case "age":
					// Ensure process age is shown when coloring by age
					processTree.DisplayOptions.ShowProcessAge = true

					// Apply color based on process age thresholds in seconds
					if process.Age < 60 {
						// Low age (< 1 minute)
						processTree.Colorizer.ProcessAgeLow(processTree.ColorScheme, value)
					} else if process.Age >= 60 && process.Age < 3600 {
						// Medium age (< 1 hour)
						processTree.Colorizer.ProcessAgeMedium(processTree.ColorScheme, value)
					} else if process.Age >= 3600 && process.Age < 86400 {
						// High age (> 1 hour and < 1 day)
						processTree.Colorizer.ProcessAgeHigh(processTree.ColorScheme, value)
					} else if process.Age >= 86400 {
						// Very high age (> 1 day)
						processTree.Colorizer.ProcessAgeVeryHigh(processTree.ColorScheme, value)
					}
				case "cpu":
					// Ensure CPU percentage is shown when coloring by CPU
					processTree.DisplayOptions.ShowCpuPercent = true

					// Apply color based on CPU usage thresholds in percentage
					if process.CPUPercent < 5 {
						// Low CPU usage (< 5%)
						processTree.Colorizer.CPULow(processTree.ColorScheme, value)
					} else if process.CPUPercent >= 5 && process.CPUPercent < 15 {
						// Medium CPU usage (5-15%)
						processTree.Colorizer.CPUMedium(processTree.ColorScheme, value)
					} else if process.CPUPercent >= 15 {
						// High CPU usage (> 15%)
						processTree.Colorizer.CPUHigh(processTree.ColorScheme, value)
					}
				case "mem":
					// Ensure memory usage is shown when coloring by memory
					processTree.DisplayOptions.ShowMemoryUsage = true

					// Calculate memory usage as percentage of total system memory
					percent := (process.MemoryInfo.RSS / processTree.DisplayOptions.InstalledMemory) * 100

					// Apply color based on memory usage thresholds in percentage
					if percent < 10 {
						// Low memory usage (< 10%)
						processTree.Colorizer.MemoryLow(processTree.ColorScheme, value)
					} else if percent >= 10 && percent < 20 {
						// Medium memory usage (10-20%)
						processTree.Colorizer.MemoryMedium(processTree.ColorScheme, value)
					} else if percent >= 20 {
						// High memory usage (> 20%)
						processTree.Colorizer.MemoryHigh(processTree.ColorScheme, value)
					}
				}
			} else {
				processTree.Colorizer.Default(processTree.ColorScheme, value)
			}
		}
	}
}

//------------------------------------------------------------------------------
// UTILITY FUNCTIONS
//------------------------------------------------------------------------------
// General utility functions used throughout the process tree implementation.

// getPidIndex finds the index of a process with the specified PID in the processes slice.
//
// Parameters:
//   - pid: The PID to search for
//
// Returns:
//   - The index of the process with the specified PID, or -1 if not found
//
// getPidIndex finds the index of a process with the specified PID in the processes slice.
//
// Parameters:
//   - pid: The PID to search for
//
// Returns:
//   - The index of the process with the specified PID, or -1 if not found
func (processTree *ProcessTree) getPidIndex(pid int32) int {
	var (
		pidIndex int
	)

	for pidIndex = range processTree.Nodes {
		if processTree.Nodes[pidIndex].PID == pid {
			return pidIndex
		}
	}
	return -1
}

// visibleWidth calculates the display width of a string containing ANSI escape sequences.
// It ignores ANSI escape sequences and counts only the visible characters' width.
// The function properly handles multi-byte Unicode characters and characters with
// different display widths (like CJK characters that take up 2 columns).
//
// Parameters:
//   - input: The string to calculate the width for, which may contain ANSI escape sequences
//
// Returns:
//
//	The display width of the string, excluding ANSI escape sequences
//
// visibleWidth calculates the display width of a string containing ANSI escape sequences.
// It ignores ANSI escape sequences and counts only the visible characters' width.
// The function properly handles multi-byte Unicode characters and characters with
// different display widths (like CJK characters that take up 2 columns).
//
// Parameters:
//   - input: The string to calculate the width for, which may contain ANSI escape sequences
//
// Returns:
//   - The display width of the string, excluding ANSI escape sequences
func (processTree *ProcessTree) visibleWidth(input string) int {
	width := 0
	for len(input) > 0 {
		if loc := ansiEscape.FindStringIndex(input); loc != nil && loc[0] == 0 {
			// Skip ANSI
			input = input[loc[1]:]
			continue
		}
		r, size := utf8.DecodeRuneInString(input)
		width += runewidth.RuneWidth(r)
		input = input[size:]
	}
	return width
}

// TruncateANSI truncates a string containing ANSI escape sequences to fit within a specified screen width.
// It preserves ANSI color and formatting codes while only counting visible characters toward the width limit.
//
// Parameters:
//   - logger: A structured logger for debug output
//   - input: The string to truncate, which may contain ANSI escape sequences
//   - screenWidth: The maximum width (in visible characters) the output string should occupy
//
// The function handles multi-byte Unicode characters correctly by using utf8.DecodeRuneInString
// and accounts for characters with different display widths using the runewidth package.
// If truncation occurs, "..." is appended to the result.
//
// Returns:
//
//	A string that fits within screenWidth, with ANSI sequences preserved.
//
// truncateANSI truncates a string containing ANSI escape sequences to fit within a specified screen width.
// It preserves ANSI color and formatting codes while only counting visible characters toward the width limit.
//
// Parameters:
//   - input: The string to truncate, which may contain ANSI escape sequences
//
// The function handles multi-byte Unicode characters correctly by using utf8.DecodeRuneInString
// and accounts for characters with different display widths using the runewidth package.
// If truncation occurs, "..." is appended to the result.
//
// Returns:
//   - A string that fits within screenWidth, with ANSI sequences preserved.
func (processTree *ProcessTree) truncateANSI(input string) string {
	dots := "..."

	if processTree.DisplayOptions.ScreenWidth <= 3 {
		return dots
	}

	// First, check actual display width
	if processTree.visibleWidth(input) <= processTree.DisplayOptions.ScreenWidth {
		return input // No truncation needed
	}

	targetWidth := processTree.DisplayOptions.ScreenWidth - len(dots)
	var output strings.Builder
	width := 0

	for len(input) > 0 {
		if loc := ansiEscape.FindStringIndex(input); loc != nil && loc[0] == 0 {
			esc := input[loc[0]:loc[1]]
			output.WriteString(esc)
			input = input[loc[1]:]
			continue
		}

		r, size := utf8.DecodeRuneInString(input)
		rw := runewidth.RuneWidth(r)

		if width+rw > targetWidth {
			break
		}

		output.WriteRune(r)
		width += rw
		input = input[size:]
	}

	output.WriteString(dots)
	return output.String() + "\x1b[0m" // Prevent ANSI bleed
}

// computeSignature recursively generates a unique signature for a process subtree.
// This includes the command, args, and all child subtrees.
// Signatures are cached in Process.Signature.
func computeSignature(p *Process, showArguments bool) string {
	if p.Signature != "" {
		return p.Signature
	}

	childSigs := make([]string, 0, len(p.Children))
	for _, c := range p.Children {
		childSigs = append(childSigs, computeSignature(c, showArguments))
	}

	// Ignore child order (matches pstree)
	sort.Strings(childSigs)

	self := p.Command
	if showArguments && len(p.Args) > 0 {
		self += " " + strings.Join(p.Args, " ")
	}

	// Include owner (matches Linux pstree)
	self += "|" + p.Username

	p.Signature = self + "(" + strings.Join(childSigs, ",") + ")"
	return p.Signature
}
