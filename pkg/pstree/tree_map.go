// Package pstree provides functionality for building and displaying process trees.
//
// This file contains an alternative implementation of the process tree using a map-based
// hierarchical structure, which is more intuitive and easier to maintain than the array-based
// approach. It's designed to work alongside the existing implementation while providing
// a path for gradual refactoring.
package pstree

import (
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/bananazon/pstree/util"
	"github.com/giancarlosio/gorainbow"
	"github.com/mattn/go-runewidth"
)

// ProcessNode represents a node in the hierarchical process tree map
type ProcessNode struct {
	Children map[int32]*ProcessNode // Child processes mapped by PID
	Depth    int                    // Depth in the tree (0 for root nodes)
	Print    bool
	Process  Process // The process data
}

type ProcessMap struct {
	DisplayOptions DisplayOptions
	Logger         *slog.Logger
	Nodes          map[int32]*ProcessNode
	TreeChars      TreeChars
	ColorScheme    ColorScheme
	Colorizer      Colorizer
}

// NewProcessMap creates a new process tree map from a slice of processes.
//
// This function initializes a ProcessMap structure, configures the tree characters based on display options,
// builds the process tree hierarchy, and marks UID transitions between processes.
//
// Parameters:
//   - logger: Logger instance for debug and informational messages
//   - processes: Slice of Process objects containing the process information
//   - displayOptions: Configuration options controlling how the tree will be displayed
//
// Returns:
//   - A pointer to the newly created ProcessMap
func NewProcessMap(logger *slog.Logger, processes []Process, displayOptions DisplayOptions) *ProcessMap {
	logger.Debug("Entering pstreeNewProcessMap()")

	processMap := &ProcessMap{
		Logger:         logger,
		Nodes:          make(map[int32]*ProcessNode),
		DisplayOptions: displayOptions,
	}

	if processMap.DisplayOptions.IBM850Graphics {
		processMap.TreeChars = TreeStyles["pc850"]
	} else if processMap.DisplayOptions.UTF8Graphics {
		processMap.TreeChars = TreeStyles["utf8"]
	} else if processMap.DisplayOptions.VT100Graphics {
		processMap.TreeChars = TreeStyles["vt100"]
	} else {
		processMap.TreeChars = TreeStyles["ascii"]
	}

	// Initialize the color scheme
	if processMap.DisplayOptions.ColorScheme != "" {
		processMap.ColorScheme = ColorSchemes[processMap.DisplayOptions.ColorScheme]
	} else {
		switch runtime.GOOS {
		case "windows":
			if os.Getenv("PSModulePath") != "" {
				processMap.ColorScheme = ColorSchemes["powershell"]
			} else {
				processMap.ColorScheme = ColorSchemes["windows10"]
			}
		case "linux":
			processMap.ColorScheme = ColorSchemes["linux"]
		case "darwin":
			processMap.ColorScheme = ColorSchemes["darwin"]
		default:
			processMap.ColorScheme = ColorSchemes["xterm"]
		}
	}

	// Initialize colorizer
	if processMap.DisplayOptions.ColorizeOutput || processMap.DisplayOptions.ColorAttr != "" {
		if processMap.DisplayOptions.ColorCount >= 8 && processMap.DisplayOptions.ColorCount <= 16 {
			processMap.Colorizer = Colorizers["8color"]
		} else if processMap.DisplayOptions.ColorCount >= 256 {
			processMap.Colorizer = Colorizers["256color"]
		}
	}

	// Build the tree
	processMap.BuildTree(processes)

	// Mark UID transitions
	processMap.markUIDTransitions()

	return processMap
}

//------------------------------------------------------------------------------
// TREE CONSTRUCTION AND STRUCTURE
//------------------------------------------------------------------------------
// Functions in this section handle the creation of the process tree structure
// and establishing the hierarchical relationships between processes.

// BuildTree constructs the hierarchical relationships between processes in the tree.
//
// This method creates nodes for all processes, establishes parent-child relationships,
// identifies root nodes, and calculates the depth of each node in the tree.
//
// Parameters:
//   - processes: Slice of Process objects to build the tree from
func (processMap *ProcessMap) BuildTree(processes []Process) {
	// First pass: Create nodes for all processes
	for i := range processes {
		processMap.Nodes[processes[i].PID] = &ProcessNode{
			Children: make(map[int32]*ProcessNode),
			Depth:    0, // Will be calculated in second pass
			Print:    false,
			Process:  processes[i],
		}
	}

	// Second pass: Build the hierarchy
	rootNodes := make(map[int32]*ProcessNode)
	for pid, node := range processMap.Nodes {
		ppid := node.Process.PPID

		// If parent exists in our map, add this as a child
		if parentNode, exists := processMap.Nodes[ppid]; exists {
			parentNode.Children[pid] = node
		} else {
			// No parent found or parent is self, this is a root node
			rootNodes[pid] = node
		}
	}

	// Special case: If we have only one root node (PID 1), use that
	// Otherwise, find the real root processes (usually just PID 1)
	if len(rootNodes) > 1 {
		// On Unix systems, PID 1 is the init process
		if node, exists := processMap.Nodes[1]; exists {
			// Keep only PID 1 as root
			newRootNodes := make(map[int32]*ProcessNode)
			newRootNodes[1] = node
			rootNodes = newRootNodes
		}
	}

	// Third pass: Calculate depth
	for _, node := range rootNodes {
		processMap.calculateDepth(node, 0)
	}

	processMap.Nodes = rootNodes
}

// calculateDepth recursively sets the depth of a node and all its children.
//
// Parameters:
//   - node: The node to calculate depth for
//   - depth: The current depth value
func (processMap *ProcessMap) calculateDepth(node *ProcessNode, depth int) {
	node.Depth = depth
	for _, child := range node.Children {
		processMap.calculateDepth(child, depth+1)
	}
}

//------------------------------------------------------------------------------
// PROCESS MARKING AND FILTERING
//------------------------------------------------------------------------------
// Functions in this section handle the identification and marking of processes
// that should be included in the display, based on various filtering criteria.

// FindPrintable marks processes that should be displayed based on filtering criteria.
// It applies various filters such as process name pattern matching, username filtering,
// root process exclusion, and PID filtering to determine which processes should be displayed.
func (processMap *ProcessMap) FindPrintable() {
	processMap.Logger.Debug("Entering processMap.FindPrintable()")
	var (
		markedPIDs []int32
		myPid      int32
		node       *ProcessNode
		pid        int32
		pids       []int32
		username   string
		showAll    bool
	)

	if processMap.DisplayOptions.Contains == "" && len(processMap.DisplayOptions.Usernames) == 0 && !processMap.DisplayOptions.ExcludeRoot && processMap.DisplayOptions.RootPID < 1 {
		showAll = true
	}

	pids = make([]int32, 0, len(processMap.Nodes))
	for pid = range processMap.Nodes {
		pids = append(pids, pid)
	}
	slices.Sort(pids)

	// Inner recursive function
	var findNestedPrintable func(node *ProcessNode, markedPIDs *[]int32)
	findNestedPrintable = func(node *ProcessNode, markedPIDs *[]int32) {
		if showAll {
			node.Print = true
		} else {
			// Junk goes here
			if len(processMap.DisplayOptions.Usernames) > 0 {
				for _, username = range processMap.DisplayOptions.Usernames {
					if node.Process.Username == username {
						processMap.findParentsAndChildren(node.Process.PID, markedPIDs)
					}
				}
			} else if node.Process.PID == processMap.DisplayOptions.RootPID {
				processMap.Logger.Debug("--pid == processMap.DisplayOptions.RootPID")
				if (processMap.DisplayOptions.ExcludeRoot && node.Process.Username != "root") || (!processMap.DisplayOptions.ExcludeRoot) {
					processMap.Logger.Debug("(processMap.DisplayOptions.ExcludeRoot && node.Process.Username != root) || !processMap.DisplayOptions.ExcludeRoot")
					processMap.findParentsAndChildren(node.Process.PID, markedPIDs)
				}
			} else if processMap.DisplayOptions.Contains != "" && strings.Contains(node.Process.Command, processMap.DisplayOptions.Contains) && (node.Process.PID != myPid) {
				processMap.Logger.Debug("processMap.DisplayOptions.Contains is set && node.Process.Command contains processMap.DisplayOptions.Contains && node.Process.PID != myPid")
				if (processMap.DisplayOptions.ExcludeRoot && node.Process.Username != "root") || (!processMap.DisplayOptions.ExcludeRoot) {
					processMap.Logger.Debug("(processMap.DisplayOptions.ExcludeRoot && node.Process.Username != root) || !processMap.DisplayOptions.ExcludeRoot")
					processMap.findParentsAndChildren(node.Process.PID, markedPIDs)
				}
			} else if processMap.DisplayOptions.Contains != "" && !strings.Contains(node.Process.Command, processMap.DisplayOptions.Contains) && (node.Process.PID != myPid) {
				processMap.Logger.Debug("processMap.DisplayOptions.Contains is set && node.Process.Command does not contain processMap.DisplayOptions.Contains && node.Process.PID != myPid")
			} else if processMap.DisplayOptions.ExcludeRoot && node.Process.Username != "root" {
				processMap.Logger.Debug("processMap.DisplayOptions.ExcludeRoot && node.Process.Username != root")
				processMap.findParentsAndChildren(node.Process.PID, markedPIDs)
			}
		}

		childPIDs := make([]int32, 0, len(node.Children))
		for pid = range node.Children {
			childPIDs = append(childPIDs, pid)
		}
		slices.Sort(childPIDs)

		for _, pid = range childPIDs {
			findNestedPrintable(node.Children[pid], markedPIDs)
		}
	}

	// Start traversal
	for _, pid = range pids {
		node = processMap.Nodes[pid]
		findNestedPrintable(node, &markedPIDs)
	}

	slices.Sort(markedPIDs)

	processMap.markPrintable(markedPIDs)
}

// findParentsAndChildren identifies all parents and children of a process with the given PID.
// The identified PIDs are added to the markedPIDs slice for later marking as printable.
//
// Parameters:
//   - pid: The process ID to find parents and children for
//   - markedPIDs: Pointer to a slice that will be populated with the PIDs to mark
func (processMap *ProcessMap) findParentsAndChildren(pid int32, markedPIDs *[]int32) {
	parentPIDs := []int32{}
	processMap.FindAllParents(pid, &parentPIDs)
	for _, pid := range parentPIDs {
		if !slices.Contains(*markedPIDs, pid) {
			*(markedPIDs) = append(*(markedPIDs), pid)
		}
	}
	childPIDs := []int32{}
	processMap.FindAllChildren(pid, &childPIDs)
	for _, pid := range childPIDs {
		if !slices.Contains(*markedPIDs, pid) {
			*(markedPIDs) = append(*(markedPIDs), pid)
		}
	}
}

// markPrintable sets the Print flag to true for all processes whose PIDs are in the markedPIDs slice.
// This makes them visible in the final tree display.
//
// Parameters:
//   - markedPIDs: Slice of process IDs to mark as printable
func (processMap *ProcessMap) markPrintable(markedPIDs []int32) {
	processMap.Logger.Debug("Entering processMap.markAllProcesses()")
	var (
		childPIDs []int32
		node      *ProcessNode
		pid       int32
		rootPIDs  []int32
	)

	rootPIDs = make([]int32, 0, len(processMap.Nodes))
	for pid = range processMap.Nodes {
		rootPIDs = append(rootPIDs, pid)
	}
	slices.Sort(rootPIDs)

	var markNestedPrintable func(node *ProcessNode)
	markNestedPrintable = func(node *ProcessNode) {
		if node.Process.PID == 1 && slices.Contains(markedPIDs, node.Process.PID) {
			processMap.Logger.Debug(fmt.Sprintf("Marking PID %d as printable", node.Process.PID))
			node.Print = true
		} else {

			if slices.Contains(markedPIDs, node.Process.PID) {
				processMap.Logger.Debug(fmt.Sprintf("Marking PID %d as printable", node.Process.PID))
				node.Print = true
			}
		}

		childPIDs = make([]int32, 0, len(node.Children))
		for pid = range node.Children {
			childPIDs = append(childPIDs, pid)
		}
		slices.Sort(childPIDs)

		for _, pid = range childPIDs {
			markNestedPrintable(node.Children[pid])
		}
	}

	// Start traversal
	for _, pid = range rootPIDs {
		node = processMap.Nodes[pid]
		markNestedPrintable(node)
	}
}

//------------------------------------------------------------------------------
// PROCESS ATTRIBUTE MARKING
//------------------------------------------------------------------------------
// Functions in this section handle marking special attributes of processes,
// such as UID transitions between parent and child processes.

// markUIDTransitions identifies and marks processes where the user ID changes from the parent process.
// This function compares the UIDs of each process with its parent and sets HasUIDTransition=true
// when a transition is detected. It also stores the parent UID for display purposes.
//
// Parameters:
//   - logger: Logger instance for debug information
//   - processes: Pointer to a slice of Process structs
func (processMap *ProcessMap) markUIDTransitions() {
	var (
		node *ProcessNode
		pid  int32
		pids []int32
	)

	processMap.Logger.Debug("Marking UID transitions between processes - START")

	pids = make([]int32, 0, len(processMap.Nodes))
	for pid = range processMap.Nodes {
		pids = append(pids, pid)
	}
	slices.Sort(pids)

	// Inner recursive function
	var markNestedUIDTransitions func(node *ProcessNode)
	markNestedUIDTransitions = func(node *ProcessNode) {
		var (
			childPIDs  []int32
			parentNode *ProcessNode
			pid        int32
		)

		if node.Process.PID > 1 {
			parentNode = processMap.FindProcess(node.Process.PPID)

			if len(node.Process.UIDs) > 0 && len(parentNode.Process.UIDs) > 0 {
				node.Process.ParentUID = parentNode.Process.UIDs[0]
				node.Process.ParentUsername = parentNode.Process.Username

				if node.Process.UIDs[0] != parentNode.Process.UIDs[0] {
					// processMap.Logger.Debug(fmt.Sprintf("UID transition detected: Process %d (UID %d) has different UID from parent %d (UID %d)",
					// 	node.Process.PID, node.Process.UIDs[0],
					// 	node.Process.PPID, parentNode.Process.UIDs[0]))
					node.Process.HasUIDTransition = true
				}
			} else {
				if node.Process.Username != parentNode.Process.Username {
					// processMap.Logger.Debug(fmt.Sprintf("Username transition detected: Process %d (%s) has different username from parent %d (%s)",
					// 	node.Process.PID, node.Process.Username,
					// 	node.Process.PPID, parentNode.Process.Username))
					node.Process.HasUIDTransition = true
				}
			}
		}

		childPIDs = make([]int32, 0, len(node.Children))
		for pid = range node.Children {
			childPIDs = append(childPIDs, pid)
		}
		slices.Sort(childPIDs)

		for _, pid = range childPIDs {
			markNestedUIDTransitions(node.Children[pid])
		}
	}

	// Start traversal
	for _, pid = range pids {
		node = processMap.Nodes[pid]
		markNestedUIDTransitions(node)
	}
}

//------------------------------------------------------------------------------
// TREE TRAVERSAL AND SEARCH
//------------------------------------------------------------------------------
// Functions in this section handle finding and traversing nodes in the process tree.

// FindProcess locates a process node with the specified PID in the tree.
//
// Parameters:
//   - targetPID: The PID of the process to find
//
// Returns:
//   - A pointer to the ProcessNode if found, nil otherwise
func (processMap *ProcessMap) FindProcess(targetPID int32) *ProcessNode {
	// Check if it's a root node
	if node, exists := processMap.Nodes[targetPID]; exists {
		return node
	}

	// Inner recursive function
	var findProcessRecursive func(node *ProcessNode, targetPID int32) *ProcessNode
	findProcessRecursive = func(node *ProcessNode, targetPID int32) *ProcessNode {
		if node.Process.PID == targetPID {
			return node
		}

		for _, child := range node.Children {
			if found := findProcessRecursive(child, targetPID); found != nil {
				return found
			}
		}
		return nil
	}

	// Search through all nodes recursively
	for _, rootNode := range processMap.Nodes {
		if found := findProcessRecursive(rootNode, targetPID); found != nil {
			return found
		}
	}
	return nil
}

func (processMap *ProcessMap) ShowPrintable() {
	processMap.Logger.Debug("Entering processMap.ShowPrintable()")
	var (
		childPIDs []int32
		node      *ProcessNode
		pid       int32
		rootPIDs  []int32
	)

	rootPIDs = make([]int32, 0, len(processMap.Nodes))
	for pid = range processMap.Nodes {
		rootPIDs = append(rootPIDs, pid)
	}
	slices.Sort(rootPIDs)

	var showNestedPrintable func(node *ProcessNode)
	showNestedPrintable = func(node *ProcessNode) {
		if node.Print {
			fmt.Printf("PID %d is printable\n", node.Process.PID)
		}

		childPIDs = make([]int32, 0, len(node.Children))
		for pid = range node.Children {
			childPIDs = append(childPIDs, pid)
		}
		slices.Sort(childPIDs)

		for _, pid = range childPIDs {
			showNestedPrintable(node.Children[pid])
		}
	}

	// Start traversal
	for _, pid = range rootPIDs {
		node = processMap.Nodes[pid]
		showNestedPrintable(node)
	}
}

//------------------------------------------------------------------------------
// TREE VISUALIZATION AND DISPLAY
//------------------------------------------------------------------------------
// Functions in this section handle the visual representation of the process tree,
// including printing the tree and formatting the output.

// PrintTree prints the process tree with indentation based on depth
// Each line shows the PID and process name, indented according to its depth in the tree
func (processMap *ProcessMap) PrintTree() {
	var (
		node *ProcessNode
		pid  int32
		pids []int32
	)

	processMap.Logger.Debug("Entering processMap.PrintTree()")
	var printNodeSimple func(node *ProcessNode, head string)
	printNodeSimple = func(node *ProcessNode, head string) {
		processMap.Logger.Debug(fmt.Sprintf("processMap.printNodeSimple(): node.PID=%d, head=\"%s\"", node.Process.PID, head))
		var (
			childPIDs []int32
			lineItem  string
			newHead   string
			pid       int32
		)

		// Begin attempt to clone tree-based logic
		if processMap.DisplayOptions.MaxDepth > 0 && node.Depth > processMap.DisplayOptions.MaxDepth {
			return
		}

		if head == "" && !node.Print {
			return
		}
		// End attempt to clone tree-based logic

		if node.Print {
			lineItem = processMap.buildLineItem(node, head)

			if !processMap.DisplayOptions.WideDisplay {
				if len(lineItem) > processMap.DisplayOptions.ScreenWidth {
					if processMap.DisplayOptions.RainbowOutput {
						lineItem = processMap.truncateANSI(gorainbow.Rainbow(lineItem))
					} else {
						lineItem = processMap.truncateANSI(lineItem)
					}
				} else if processMap.DisplayOptions.RainbowOutput {
					lineItem = gorainbow.Rainbow(lineItem)
				}
			} else if processMap.DisplayOptions.RainbowOutput {
				lineItem = gorainbow.Rainbow(lineItem)
			}

			processMap.Logger.Debug(fmt.Sprintf("processMap.printNodeSimple(): printing line for node.PID=%d, head=\"%s\"", node.Process.PID, head))
			fmt.Fprintln(os.Stdout, lineItem)

			newHead = processMap.buildNewHead(head, node)
		}

		// Process children
		childPIDs = make([]int32, 0, len(node.Children))
		for pid = range node.Children {
			childPIDs = append(childPIDs, pid)
		}

		slices.Sort(childPIDs)
		for _, pid = range childPIDs {
			printNodeSimple(node.Children[pid], newHead)
		}
	}

	// Sort root PIDs for consistent output
	pids = make([]int32, 0, len(processMap.Nodes))
	for pid = range processMap.Nodes {
		pids = append(pids, pid)
	}
	slices.Sort(pids)

	// Print each root node
	for _, pid = range pids {
		node = processMap.Nodes[pid]
		printNodeSimple(node, "")
	}
}

// buildLinePrefix constructs the tree visualization prefix for a process node in the tree display.
// It creates the branch connectors (├, └, etc.) that show the hierarchical relationship between processes.
//
// Parameters:
//   - node: The current process node
//   - head: The accumulated prefix string from parent levels
//
// Returns:
//   - A formatted string containing tree branch characters that represent the process's position in the hierarchy
func (processMap *ProcessMap) buildLinePrefix(node *ProcessNode, head string) string {
	processMap.Logger.Debug(fmt.Sprintf("processMap.buildLinePrefix(head=\"%s\", pid=%d, depth=%d)", head, node.Process.PID, node.Depth))

	// Create a strings.Builder with an estimated capacity
	// This helps avoid reallocations as the builder grows
	var builder strings.Builder

	// Pre-allocate capacity based on expected size
	// This is an optimization to avoid reallocations
	// You can adjust the capacity based on typical usage patterns
	builder.Grow(len(head) + 50) // Estimate based on typical usage

	// Append initialization sequences
	builder.WriteString(processMap.TreeChars.Init)
	builder.WriteString(processMap.TreeChars.SG)
	builder.WriteString(head)

	if node.Process.PID == 1 {
		// This is a worakround
		builder.WriteString(processMap.TreeChars.P)
		if processMap.DisplayOptions.ShowPGLs {
			builder.WriteString(processMap.TreeChars.PGL)
		} else {
			builder.WriteString(processMap.TreeChars.NPGL)
		}
		builder.WriteString(processMap.TreeChars.EG)
		return builder.String()
	}

	if head == "" {
		return ""
	} else {
		if processMap.IsLastChild(node) {
			builder.WriteString(processMap.TreeChars.BarL)
		} else if processMap.HasVisibleSiblings(node) {
			builder.WriteString(processMap.TreeChars.BarC)
		} else if processMap.IsLastSibling(node) {
			builder.WriteString(processMap.TreeChars.BarL)
		}
		// if processMap.HasVisibleSiblings(node) {
		// 	builder.WriteString(processMap.TreeChars.BarC)
		// } else if processMap.IsLastChild(node) {
		// 	builder.WriteString(processMap.TreeChars.BarL)
		// }
	}

	// if head == "" {
	// 	return ""
	// } else {
	// 	if processMap.DisplayOptions.CompactMode {
	// 		// Do compact mode junk here
	// 	} else {
	// 		if processMap.IsLastChild(node) {
	// 			builder.WriteString(processMap.TreeChars.BarL)
	// 		} else if processMap.HasVisibleSiblings(node) {
	// 			builder.WriteString(processMap.TreeChars.BarC)
	// 		}
	// 	}
	// }

	// if len(node.Children) > 0 {
	// 	builder.WriteString(processMap.TreeChars.P)
	// } else {
	// 	builder.WriteString(processMap.TreeChars.S2)
	// }

	// try to fix for level
	if len(node.Children) > 0 && node.Depth < processMap.DisplayOptions.MaxDepth {
		builder.WriteString(processMap.TreeChars.P)
	} else {
		builder.WriteString(processMap.TreeChars.S2)
	}

	if node.Process.PID == node.Process.PGID {
		if processMap.DisplayOptions.ShowPGLs {
			builder.WriteString(processMap.TreeChars.PGL)
		} else {
			builder.WriteString(processMap.TreeChars.NPGL)
		}
	} else {
		builder.WriteString(processMap.TreeChars.NPGL)
	}
	builder.WriteString(processMap.TreeChars.EG)

	// Return the completed string
	return builder.String()
}

// buildNewHead constructs a new head string for child processes based on the current process's position.
//
// Parameters:
//   - head: The accumulated prefix string from parent levels
//   - node: The current process node
//
// Returns:
//   - A string to be used as the head for child processes, including appropriate vertical bars
//     or spaces based on whether the current process has visible siblings
func (processMap *ProcessMap) buildNewHead(head string, node *ProcessNode) string {
	newHead := fmt.Sprintf("%s%s ", head,
		func() string {
			if head == "" {
				return ""
			}

			if processMap.IsLastSibling(node) {
				// processMap.Logger.Debug(fmt.Sprintf("pid=%d (1) is returning \"%s\"", node.Process.PID, " "))
				return " "
			}
			// processMap.Logger.Debug(fmt.Sprintf("pid=%d (2) is returning \"%s\"", node.Process.PID, processMap.TreeChars.Bar))
			return processMap.TreeChars.Bar
		}(),
	)
	processMap.Logger.Debug(fmt.Sprintf("pid=%d head=\"%s\" newHead=\"%s\"", node.Process.PID, head, newHead))

	return newHead
}

// buildLineItem constructs a complete formatted line for a process in the tree display.
// It combines the tree structure prefix with various process information based on display options.
//
// Parameters:
//   - node: The process node to format
//   - head: The accumulated prefix string from parent levels
//
// Returns:
//   - A fully formatted string containing the process information with appropriate formatting
//     including elements such as tree structure, process IDs, resource usage, and command information
func (processMap *ProcessMap) buildLineItem(node *ProcessNode, head string) string {
	var (
		ageString       string
		args            string
		commandStr      string
		cpuPercent      string
		linePrefix      string
		memoryUsage     string
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

	linePrefix = processMap.buildLinePrefix(node, head)
	// processMap.colorizeField("prefix", &linePrefix, &node.Process)

	builder.WriteString(linePrefix)
	builder.WriteString(" ")

	if processMap.DisplayOptions.ShowOwner {
		owner := node.Process.Username
		processMap.colorizeField("owner", &owner, &node.Process)
		builder.WriteString(owner)
		builder.WriteString(" ")
	}

	if processMap.DisplayOptions.ShowPPIDs {
		ppidString = util.Int32toStr(node.Process.PPID)
		pidPgidSlice = append(pidPgidSlice, ppidString)
	}

	if processMap.DisplayOptions.ShowPIDs {
		pidString = util.Int32toStr(node.Process.PID)
		pidPgidSlice = append(pidPgidSlice, pidString)
	}

	if processMap.DisplayOptions.ShowPGIDs {
		pgidString = util.Int32toStr(node.Process.PGID)
		pidPgidSlice = append(pidPgidSlice, pgidString)
	}

	if len(pidPgidSlice) > 0 {
		pidPgidString = fmt.Sprintf("(%s)", strings.Join(pidPgidSlice, ","))
		processMap.colorizeField("pidPgid", &pidPgidString, &node.Process)
		builder.WriteString(pidPgidString)
		builder.WriteString(" ")
	}

	if processMap.DisplayOptions.ShowProcessAge {
		duration := util.FindDuration(node.Process.Age)
		ageSlice := []string{}
		ageSlice = append(ageSlice, fmt.Sprintf("%02d", duration.Days))
		ageSlice = append(ageSlice, fmt.Sprintf("%02d", duration.Hours))
		ageSlice = append(ageSlice, fmt.Sprintf("%02d", duration.Minutes))
		ageSlice = append(ageSlice, fmt.Sprintf("%02d", duration.Seconds))
		ageString = fmt.Sprintf(
			"(%s)",
			strings.Join(ageSlice, ":"),
		)
		processMap.colorizeField("age", &ageString, &node.Process)
		builder.WriteString(ageString)
		builder.WriteString(" ")
	}

	if processMap.DisplayOptions.ShowCpuPercent {
		cpuPercent = fmt.Sprintf("(c:%.2f%%)", node.Process.CPUPercent)
		processMap.colorizeField("cpu", &cpuPercent, &node.Process)
		builder.WriteString(cpuPercent)
		builder.WriteString(" ")
	}

	if processMap.DisplayOptions.ShowMemoryUsage {
		memoryUsage = fmt.Sprintf("(m:%s)", util.ByteConverter(node.Process.MemoryInfo.RSS))
		processMap.colorizeField("memory", &memoryUsage, &node.Process)
		builder.WriteString(memoryUsage)
		builder.WriteString(" ")
	}

	if processMap.DisplayOptions.ShowNumThreads {
		// Always show thread count, even when showing compact format
		threads = fmt.Sprintf("(t:%d)", node.Process.NumThreads)
		processMap.colorizeField("threads", &threads, &node.Process)
		builder.WriteString(threads)
		builder.WriteString(" ")
	}

	if processMap.DisplayOptions.ShowUIDTransitions && node.Process.HasUIDTransition {
		// Add UID transition notation {parentUID→currentUID}
		if len(node.Process.UIDs) > 0 {
			ownerTransition = fmt.Sprintf("(%d→%d)", node.Process.ParentUID, node.Process.UIDs[0])
		}
	} else if processMap.DisplayOptions.ShowUserTransitions && node.Process.HasUIDTransition {
		// Add user transition notation {parentUser→currentUser}
		if node.Process.ParentUsername != "" {
			ownerTransition = fmt.Sprintf("(%s→%s)", node.Process.ParentUsername, node.Process.Username)
		}
	}

	if ownerTransition != "" {
		processMap.colorizeField("ownerTransition", &ownerTransition, &node.Process)
		builder.WriteString(ownerTransition)
		builder.WriteString(" ")
	}

	commandStr = node.Process.Command
	processMap.colorizeField("command", &commandStr, &node.Process)
	builder.WriteString(commandStr)
	builder.WriteString(" ")

	if processMap.DisplayOptions.ShowArguments {
		if len(node.Process.Args) > 0 {
			args = strings.Join(node.Process.Args, " ")
			processMap.colorizeField("args", &args, &node.Process)
			builder.WriteString(args)
			builder.WriteString(" ")
		}
	}

	return builder.String()
}

//------------------------------------------------------------------------------
// TREE STRUCTURE HELPERS
//------------------------------------------------------------------------------
// Functions in this section provide utility methods for determining the
// relationships between nodes in the tree structure.

// HasVisibleSiblings determines if a process has visible siblings
// This is useful for drawing the correct branch characters in the tree
func (processMap *ProcessMap) HasVisibleSiblings(node *ProcessNode) bool {
	// Find the parent of this node
	parentPID := node.Process.PPID
	parent := processMap.FindProcess(parentPID)

	// If no parent found or parent is the node itself, it has no siblings
	if parent == nil || parent.Process.PID == node.Process.PID {
		return false
	}

	// Count visible siblings (children of the same parent)
	visibleSiblings := 0
	for _, child := range parent.Children {
		// Skip self when counting siblings
		if child.Process.PID != node.Process.PID {
			visibleSiblings++
		}
	}

	return visibleSiblings > 0
}

// IsLastChild determines if a node is the last child of its parent
// This is useful for drawing the correct branch characters in the tree
func (processMap *ProcessMap) IsLastChild(node *ProcessNode) bool {
	// Find the parent of this node
	parentPID := node.Process.PPID
	parent := processMap.FindProcess(parentPID)

	// If no parent found or parent is the node itself, it's not a child
	if parent == nil || parent.Process.PID == node.Process.PID {
		return false
	}

	// Get all children of the parent
	childPIDs := make([]int32, 0, len(parent.Children))
	for pid := range parent.Children {
		childPIDs = append(childPIDs, pid)
	}

	// Sort by PID for consistent comparison
	slices.Sort(childPIDs)

	// Check if this node is the last child
	return len(childPIDs) > 0 && childPIDs[len(childPIDs)-1] == node.Process.PID
}

// IsLastSibling determines if a node is the last sibling among its parent's children
func (processMap *ProcessMap) IsLastSibling(node *ProcessNode) bool {
	// Find the parent of this node
	parentPID := node.Process.PPID
	parent := processMap.FindProcess(parentPID)

	// If no parent found or parent is the node itself, it's not a sibling
	if parent == nil || parent.Process.PID == node.Process.PID {
		return false
	}

	// Get all children of the parent
	childPIDs := make([]int32, 0, len(parent.Children))
	for pid := range parent.Children {
		childPIDs = append(childPIDs, pid)
	}

	// Sort by PID for consistent comparison
	slices.Sort(childPIDs)

	// Check if this node is the last child
	return len(childPIDs) > 0 && childPIDs[len(childPIDs)-1] == node.Process.PID
}

// FindAllParents identifies all parent processes of a given PID and adds them to the parentPIDs slice.
// This function recursively traverses up the process tree to find all ancestors.
//
// Parameters:
//   - pid: The process ID to find parents for
//   - parentPIDs: Pointer to a slice that will be populated with parent PIDs
func (processMap *ProcessMap) FindAllParents(pid int32, parentPIDs *[]int32) {
	node := processMap.FindProcess(pid)
	if node != nil {
		if !slices.Contains(*parentPIDs, node.Process.PID) {
			*(parentPIDs) = append(*(parentPIDs), pid)
		}
		for node.Process.PPID > 0 {
			if !slices.Contains(*parentPIDs, node.Process.PPID) {
				*(parentPIDs) = append(*(parentPIDs), node.Process.PPID)
			}
			node = processMap.FindProcess(node.Process.PPID)
		}
	}
}

// FindAllChildren identifies all child processes of a given PID and adds them to the childPIDs slice.
// This function recursively traverses down the process tree to find all descendants.
//
// Parameters:
//   - pid: The process ID to find children for
//   - childPIDs: Pointer to a slice that will be populated with child PIDs
func (processMap *ProcessMap) FindAllChildren(pid int32, childPIDs *[]int32) {
	node := processMap.FindProcess(pid)
	if node != nil {
		if !slices.Contains(*childPIDs, pid) {
			*(childPIDs) = append(*(childPIDs), pid)
		}
		for _, childNode := range node.Children {
			if !slices.Contains(*childPIDs, childNode.Process.PID) {
				*(childPIDs) = append(*(childPIDs), childNode.Process.PID)
			}
			processMap.FindAllChildren(childNode.Process.PID, childPIDs)
		}
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
func (processMap *ProcessMap) colorizeField(fieldName string, value *string, process *Process) {
	// Only apply colors if the terminal supports them
	if processMap.DisplayOptions.ColorSupport {
		// Standard colorization mode (--colorize flag)
		if processMap.DisplayOptions.ColorizeOutput {
			// Apply specific colors based on the field type
			switch fieldName {
			case "age":
				processMap.Colorizer.Age(processMap.ColorScheme, value)
			case "args":
				processMap.Colorizer.Args(processMap.ColorScheme, value)
			case "command":
				processMap.Colorizer.Command(processMap.ColorScheme, value)
			case "compactStr":
				processMap.Colorizer.CompactStr(processMap.ColorScheme, value)
			case "connector":
				processMap.Colorizer.Connector(processMap.ColorScheme, value)
			case "cpu":
				processMap.Colorizer.CPU(processMap.ColorScheme, value)
			case "memory":
				processMap.Colorizer.Memory(processMap.ColorScheme, value)
			case "owner":
				processMap.Colorizer.Owner(processMap.ColorScheme, value)
			case "ownerTransition":
				processMap.Colorizer.OwnerTransition(processMap.ColorScheme, value)
			case "pidPgid":
				processMap.Colorizer.PIDPGID(processMap.ColorScheme, value)
			case "prefix":
				processMap.Colorizer.Prefix(processMap.ColorScheme, value)
			case "threads":
				processMap.Colorizer.NumThreads(processMap.ColorScheme, value)
			}
		} else if processMap.DisplayOptions.ColorAttr != "" {
			// Attribute-based colorization mode (--color flag)
			// Don't apply attribute-based coloring to the tree prefix
			if fieldName != "prefix" {
				switch processMap.DisplayOptions.ColorAttr {
				case "age":
					// Ensure process age is shown when coloring by age
					processMap.DisplayOptions.ShowProcessAge = true

					// Apply color based on process age thresholds in seconds
					if process.Age < 60 {
						// Low age (< 1 minute)
						processMap.Colorizer.ProcessAgeLow(processMap.ColorScheme, value)
					} else if process.Age >= 60 && process.Age < 3600 {
						// Medium age (< 1 hour)
						processMap.Colorizer.ProcessAgeMedium(processMap.ColorScheme, value)
					} else if process.Age >= 3600 && process.Age < 86400 {
						// High age (> 1 hour and < 1 day)
						processMap.Colorizer.ProcessAgeHigh(processMap.ColorScheme, value)
					} else if process.Age >= 86400 {
						// Very high age (> 1 day)
						processMap.Colorizer.ProcessAgeVeryHigh(processMap.ColorScheme, value)
					}
				case "cpu":
					// Ensure CPU percentage is shown when coloring by CPU
					processMap.DisplayOptions.ShowCpuPercent = true

					// Apply color based on CPU usage thresholds in percentage
					if process.CPUPercent < 5 {
						// Low CPU usage (< 5%)
						processMap.Colorizer.CPULow(processMap.ColorScheme, value)
					} else if process.CPUPercent >= 5 && process.CPUPercent < 15 {
						// Medium CPU usage (5-15%)
						processMap.Colorizer.CPUMedium(processMap.ColorScheme, value)
					} else if process.CPUPercent >= 15 {
						// High CPU usage (> 15%)
						processMap.Colorizer.CPUHigh(processMap.ColorScheme, value)
					}
				case "mem":
					// Ensure memory usage is shown when coloring by memory
					processMap.DisplayOptions.ShowMemoryUsage = true

					// Calculate memory usage as percentage of total system memory
					percent := (process.MemoryInfo.RSS / processMap.DisplayOptions.InstalledMemory) * 100

					// Apply color based on memory usage thresholds in percentage
					if percent < 10 {
						// Low memory usage (< 10%)
						processMap.Colorizer.MemoryLow(processMap.ColorScheme, value)
					} else if percent >= 10 && percent < 20 {
						// Medium memory usage (10-20%)
						processMap.Colorizer.MemoryLow(processMap.ColorScheme, value)
					} else if percent >= 20 {
						// High memory usage (> 20%)
						processMap.Colorizer.MemoryHigh(processMap.ColorScheme, value)
					}
				}
			} else {
				processMap.Colorizer.Default(processMap.ColorScheme, value)
			}
		}
	}
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
func (processMap *ProcessMap) visibleWidth(input string) int {
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
func (processMap *ProcessMap) truncateANSI(input string) string {
	dots := "..."

	if processMap.DisplayOptions.ScreenWidth <= 3 {
		return dots
	}

	// First, check actual display width
	if processMap.visibleWidth(input) <= processMap.DisplayOptions.ScreenWidth {
		return input // No truncation needed
	}

	targetWidth := processMap.DisplayOptions.ScreenWidth - len(dots)
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
