// Package pstree provides functionality for building and displaying process trees.
//
// This file contains the core data structures and type definitions used throughout the package.
// It defines the Process, DisplayOptions, ProcessTree, and TreeChars types that form the foundation
// of the process tree visualization system.
package pstree

import (
	"log/slog"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

const (
	// Reset
	AnsiReset = "\033[0m"

	// Regular Colors
	AnsiBlack   = "\033[30m"
	AnsiRed     = "\033[31m"
	AnsiGreen   = "\033[32m"
	AnsiYellow  = "\033[33m"
	AnsiBlue    = "\033[34m"
	AnsiMagenta = "\033[35m"
	AnsiCyan    = "\033[36m"
	AnsiWhite   = "\033[37m"

	// Bold Colors (technically "bright", but often shown as bold in terminals)
	AnsiBlackBold   = "\033[1;30m"
	AnsiRedBold     = "\033[1;31m"
	AnsiGreenBold   = "\033[1;32m"
	AnsiYellowBold  = "\033[1;33m"
	AnsiBlueBold    = "\033[1;34m"
	AnsiMagentaBold = "\033[1;35m"
	AnsiCyanBold    = "\033[1;36m"
	AnsiWhiteBold   = "\033[1;37m"
)

//------------------------------------------------------------------------------
// CORE DATA STRUCTURES
//------------------------------------------------------------------------------

// Process represents a system process with all its attributes and relationships.
// It combines information gathered from both gopsutil and direct ps command calls
// to provide comprehensive details about the process.
type Process struct {
	// Process age in seconds since creation
	Age int64
	// Command line arguments
	Args []string
	// Background status of the process
	Background bool
	// Index of the first child process in the process tree
	Child int
	// Pointer to a slice of child processes
	Children []*Process
	// Command name (executable name)
	Command string
	// Network connections associated with this process
	Connections []net.ConnectionStat
	// CPU Affinity
	CPUAffinity []int32
	// CPU usage percentage
	CPUPercent float64
	// CPU time statistics
	CPUTimes *cpu.TimesStat
	// Process creation time as Unix timestamp
	CreateTime int64
	// Environment variables
	Environment []string
	// Foreground status of the process
	Foreground bool
	// Group IDs associated with this process
	GIDs []uint32
	// Groups associated with this process
	Groups []uint32
	// Indicates if this process has a different UID from its parent
	HasUIDTransition bool
	// Process hierarchy
	Hierarchy map[int32][]string
	// Indicates if this process is the current process or an ancestor
	IsCurrentOrAncestor bool
	// IO counters associated with this process
	IOCounters *process.IOCountersStat
	// Memory usage information
	MemoryInfo *process.MemoryInfoStat
	// Platform-specific memory usage information
	MemoryInfoEx *process.MemoryInfoExStat
	// Memory usage as percentage of total system memory
	MemoryPercent float32
	// Number of file descriptors
	NumFDs int32
	// Number of context switches
	NumContextSwitches *process.NumCtxSwitchesStat
	// Number of threads
	NumThreads int32
	// Open files
	OpenFiles []process.OpenFilesStat
	// Page faults associated with this process
	PageFaults *process.PageFaultsStat
	// Index of the parent process in the process tree
	Parent int
	// Pointer to the parent process
	ParentProcess *Process
	// UID of the parent process
	ParentUID uint32
	// Username of the parent process
	ParentUsername string
	// Process group ID
	PGID int32
	// Process ID
	PID int32
	// The PID index
	PidIndex int
	// Parent process ID
	PPID int32
	// Whether or not we plan to display this process
	Print bool
	// Resource limits associated with this process
	ResourceLimit []process.RlimitStat
	// Resource limits associated with this process
	ResourceLimitUsage []process.RlimitStat
	// Cached subtree signature
	Signature string // cached subtree signature
	// Index of the next sibling process in the process tree
	Sister int
	// Process status information
	Status []string
	// Threads associated with this process
	Threads map[int32]*cpu.TimesStat
	// User IDs associated with this process
	UIDs []uint32
	// Username of the process owner
	Username string
}

//------------------------------------------------------------------------------
// DISPLAY CONFIGURATION
//------------------------------------------------------------------------------

// DisplayOptions controls how the process tree is displayed, including formatting,
// coloring, and which information is shown for each process.
type DisplayOptions struct {
	// Attribute to color by ("age", "cpu", or "mem")
	ColorAttr string
	// Number of colors to use in rainbow mode
	ColorCount int
	// Whether to colorize the output with predefined colors
	ColorizeOutput bool
	// The system color scheme to use
	ColorScheme string
	// Whether the terminal supports color output
	ColorSupport bool
	// Whether to compact identical processes in the tree
	CompactMode bool
	// String to search for in process names
	Contains string
	// Whether to exclude processes owned by root
	ExcludeRoot bool
	// Whether to hide threads in the output
	HideThreads bool
	// Whether to use IBM850 graphics characters for tree lines
	IBM850Graphics bool
	// Total installed system memory in bytes
	InstalledMemory uint64
	// Maximum depth of the tree to display (0 for unlimited)
	MaxDepth int
	// Sort the results by a number of fields
	OrderBy string
	// Whether to use rainbow colors for output
	RainbowOutput bool
	// Root process PID
	RootPID int32
	// Width of the terminal screen in characters
	ScreenWidth int
	// Whether to show command line arguments
	ShowArguments bool
	// Whether to show CPU usage percentage
	ShowCpuPercent bool
	// Whether to show memory usage
	ShowMemoryUsage bool
	// Whether to show thread count
	ShowNumThreads bool
	// Whether to show process owner
	ShowOwner bool
	// Whether to highlight process group leaders
	ShowPGLs bool
	// Whether to show process group IDs
	ShowPGIDs bool
	// Whether to show process IDs
	ShowPIDs bool
	// Whether to show parent process IDs
	ShowPPIDs bool
	// Whether to show process age
	ShowProcessAge bool
	// Whether to show UID transitions
	ShowUIDTransitions bool
	// Whether to show username transitions
	ShowUserTransitions bool
	// Whether to use UTF-8 graphics characters for tree lines
	UTF8Graphics bool
	// List of usernames to filter by
	Usernames []string
	// Whether to use VT100 graphics characters for tree lines
	VT100Graphics bool
	// Whether to display wide output (not truncated to screen width)
	WideDisplay bool
}

//------------------------------------------------------------------------------
// TREE STRUCTURE
//------------------------------------------------------------------------------

// ProcessTree handles the construction and display of the process tree.
// It maintains the tree structure and provides methods for building,
// manipulating, and displaying the process hierarchy.
type ProcessTree struct {
	// Current depth in the tree during traversal
	AtDepth int
	// Colorizer for applying colors to text
	Colorizer Colorizer
	// Color scheme for applying colors to text
	ColorScheme ColorScheme
	// Enable debugging
	DebugLevel int
	// Display options controlling how the tree is rendered
	DisplayOptions DisplayOptions
	// Map from index in the Nodes array to PID
	IndexToPidMap map[int]int32
	// Logger for debug and informational messages
	Logger *slog.Logger
	// Array of process nodes in the tree
	Nodes []*Process
	// Map from PID to index in the Nodes array for quick lookups
	PidToIndexMap map[int32]int
	// Process groups for grouping identical processes
	ProcessGroups map[int32]map[string]map[string]ProcessGroup
	// PID of the root process for the tree
	RootPID int32
	// Tree characters for drawing the tree
	TreeChars TreeChars
}

// ------------------------------------------------------------------------------
// GROUP OF IDENTICAL PROCESSES
// ------------------------------------------------------------------------------
type ProcessGroup struct {
	// Number of identical processes
	Count int
	// CPU usage of the group
	CPUPercent float64
	// Index of the first process in the group
	FirstIndex int
	// Full path of the command
	FullPath string
	// Indices of all processes in the group
	Indices []int
	// Memory usage of the group
	MemoryUsage uint64
	// Thread count of the group
	NumThreads int32
	// The process owner
	Owner string
	// Subtree signature for comparison
	Signature string
}

//------------------------------------------------------------------------------
// TREE VISUALIZATION
//------------------------------------------------------------------------------

// TreeChars defines the characters used for drawing the tree.
// Different character sets are available for different terminal types and preferences.
type TreeChars struct {
	// Bar represents the vertical bar character (│) used for drawing process tree lines
	Bar string
	// BarC represents the T-junction character (├) used where a process branches off
	BarC string
	// BarL represents the L-junction character (└) used for the last child process in a branch
	BarL string
	// EG represents the End Graphics character sequence for terminating graphic mode
	EG string
	// Init represents the initialization sequence for the terminal
	Init string
	// NPGL represents the character sequence to initialize the graphic set for non-process group leaders
	NPGL string
	// P represents the horizontal line character (─) used for connecting processes
	P string
	// PGL represents the character sequence used to highlight process group leaders
	PGL string
	// S2 represents the secondary process horizontal line character (─) for alternative styling
	S2 string
	// SG represents the Start Graphics character sequence for entering graphic mode
	SG string
}

// TreeStyles defines different graphical styles for tree visualization.
// Each style uses a different set of characters for drawing the tree structure,
// allowing users to choose the style that works best with their terminal.
var TreeStyles = map[string]TreeChars{
	// https://github.com/FredHucht/pstree/blob/main/pstree.c#L192-L207
	"ascii": {
		Bar:  "|",  // B
		BarC: "|",  // C
		BarL: "\\", // L
		EG:   "",   // eg
		Init: "",   // init
		NPGL: "-",  // N
		P:    "-+", // PP
		PGL:  "=",  // G
		S2:   "--", // ss
		SG:   "",   // sg
	},
	"pc850": {
		Bar:  string([]byte{0xB3}),       // B
		BarC: string([]byte{0xC3}),       // C
		BarL: string([]byte{0xB4}),       // L
		EG:   string([]byte{}),           // eg
		Init: string([]byte{}),           // init
		NPGL: string([]byte{0xDA}),       // N
		P:    string([]byte{0xDA, 0xC2}), // PP
		PGL:  "¤",                        // G
		S2:   string([]byte{0xDA, 0xDA}), // ss
		SG:   string([]byte{}),           // sg
	},
	"vt100": {
		Bar:  "\x0Ex\x0F",    // B
		BarC: "\x0Et\x0F",    // C
		BarL: "\x0Em\x0F",    // L
		EG:   "\x0F",         // eg
		Init: "\033(B\033)0", // init
		NPGL: "\x0Eq\x0F",    // N
		P:    "\x0Eqw\x0F",   // PP
		PGL:  "◆",            // G
		S2:   "\x0Eqq\x0F",   // ss
		SG:   "\x0E",         // sg
	},
	"utf8": {
		Bar:  "\342\224\202",             // B
		BarC: "\342\224\234",             // C
		BarL: "\342\224\224",             // L
		EG:   "",                         // eg
		Init: "",                         // init
		NPGL: "\342\224\200",             // N
		P:    "\342\224\200\342\224\254", // PP
		PGL:  "●",                        // G
		S2:   "\342\224\200\342\224\200", // ss
		SG:   "",                         // sg
	},
}

type ColorFunc func(cs ColorScheme, text *string)

type Colorizer struct {
	Age                ColorFunc
	Args               ColorFunc
	Command            ColorFunc
	CompactStr         ColorFunc
	Connector          ColorFunc
	CPU                ColorFunc
	Memory             ColorFunc
	NumThreads         ColorFunc
	Owner              ColorFunc
	OwnerTransition    ColorFunc
	PIDPGID            ColorFunc
	Prefix             ColorFunc
	ProcessAgeLow      ColorFunc
	ProcessAgeMedium   ColorFunc
	ProcessAgeHigh     ColorFunc
	ProcessAgeVeryHigh ColorFunc
	CPULow             ColorFunc
	CPUMedium          ColorFunc
	CPUHigh            ColorFunc
	MemoryLow          ColorFunc
	MemoryMedium       ColorFunc
	MemoryHigh         ColorFunc
	Default            ColorFunc
}

var Colorizers = map[string]Colorizer{
	"8color": {
		Age:                Color8GreenBold,
		Args:               Color8Red,
		Command:            Color8BlueBold,
		CompactStr:         Color8BlackBold,
		Connector:          Color8BlackBold,
		CPU:                Color8YellowBold,
		Memory:             Color8RedBold,
		NumThreads:         Color8WhiteBold,
		Owner:              Color8CyanBold,
		OwnerTransition:    Color8BlackBold,
		PIDPGID:            Color8MagentaBold,
		Prefix:             Color8Green,
		ProcessAgeLow:      Color8Red,
		ProcessAgeMedium:   Color8Yellow,
		ProcessAgeHigh:     Color8Cyan,
		ProcessAgeVeryHigh: Color8Green,
		CPULow:             Color8Green,
		CPUMedium:          Color8Yellow,
		CPUHigh:            Color8Red,
		MemoryLow:          Color8Green,
		MemoryMedium:       Color8Yellow,
		MemoryHigh:         Color8Red,
		Default:            Color8Green,
	},
	"256color": {
		Age:                Color256Green,
		Args:               Color256Red,
		Command:            Color256Blue,
		CompactStr:         Color256BlackBold,
		Connector:          Color256BlackBold,
		CPU:                Color256Yellow,
		Memory:             Color256Orange,
		NumThreads:         Color256White,
		Owner:              Color256Cyan,
		OwnerTransition:    Color256BlackBold,
		PIDPGID:            Color256Magenta,
		Prefix:             Color256Green,
		ProcessAgeLow:      Color256Red,
		ProcessAgeMedium:   Color256Yellow,
		ProcessAgeHigh:     Color256Cyan,
		ProcessAgeVeryHigh: Color256Green,
		CPULow:             Color256Green,
		CPUMedium:          Color256Yellow,
		CPUHigh:            Color256Red,
		MemoryLow:          Color256Green,
		MemoryMedium:       Color256Yellow,
		MemoryHigh:         Color256Red,
		Default:            Color256Green,
	},
}

type ColorMap struct {
	R    int
	G    int
	B    int
	Ansi string
}

type ColorScheme struct {
	Black       ColorMap
	BlackBold   ColorMap
	Blue        ColorMap
	BlueBold    ColorMap
	Cyan        ColorMap
	CyanBold    ColorMap
	Green       ColorMap
	GreenBold   ColorMap
	Orange      ColorMap
	OrangeBold  ColorMap
	Magenta     ColorMap
	MagentaBold ColorMap
	Red         ColorMap
	RedBold     ColorMap
	White       ColorMap
	WhiteBold   ColorMap
	Yellow      ColorMap
	YellowBold  ColorMap
}

// https://en.wikipedia.org/wiki/ANSI_escape_code#Colors
// https://www.ditig.com/256-colors-cheat-sheet
var ColorSchemes map[string]ColorScheme = map[string]ColorScheme{
	"windows10": {
		Black:       ColorMap{R: 12, G: 12, B: 12},
		BlackBold:   ColorMap{R: 118, G: 118, B: 118},
		Blue:        ColorMap{R: 0, G: 255, B: 218},
		BlueBold:    ColorMap{R: 59, G: 120, B: 255},
		Cyan:        ColorMap{R: 58, G: 150, B: 221},
		CyanBold:    ColorMap{R: 97, G: 214, B: 214},
		Green:       ColorMap{R: 19, G: 161, B: 14},
		GreenBold:   ColorMap{R: 22, G: 198, B: 12},
		Magenta:     ColorMap{R: 136, G: 23, B: 152},
		MagentaBold: ColorMap{R: 180, G: 0, B: 158},
		Red:         ColorMap{R: 197, G: 15, B: 31},
		RedBold:     ColorMap{R: 231, G: 72, B: 86},
		White:       ColorMap{R: 204, G: 204, B: 204},
		WhiteBold:   ColorMap{R: 242, G: 242, B: 242},
		Yellow:      ColorMap{R: 193, G: 156, B: 0},
		YellowBold:  ColorMap{R: 249, G: 241, B: 165},
		// Not part of the standard 16 colors
		Orange:     ColorMap{R: 215, G: 95, B: 0},
		OrangeBold: ColorMap{R: 255, G: 135, B: 0},
	},
	"powershell": {
		Black:       ColorMap{R: 0, G: 0, B: 0},
		BlackBold:   ColorMap{R: 128, G: 128, B: 128},
		Blue:        ColorMap{R: 0, G: 0, B: 128},
		BlueBold:    ColorMap{R: 0, G: 0, B: 255},
		Cyan:        ColorMap{R: 0, G: 128, B: 128},
		CyanBold:    ColorMap{R: 0, G: 255, B: 255},
		Green:       ColorMap{R: 0, G: 128, B: 0},
		GreenBold:   ColorMap{R: 0, G: 255, B: 0},
		Magenta:     ColorMap{R: 1, G: 36, B: 86},
		MagentaBold: ColorMap{R: 255, G: 0, B: 255},
		Red:         ColorMap{R: 128, G: 0, B: 0},
		RedBold:     ColorMap{R: 255, G: 0, B: 0},
		White:       ColorMap{R: 192, G: 192, B: 192},
		WhiteBold:   ColorMap{R: 255, G: 255, B: 255},
		Yellow:      ColorMap{R: 237, G: 237, B: 240},
		YellowBold:  ColorMap{R: 255, G: 255, B: 0},
		// Not part of the standard 16 colors
		Orange:     ColorMap{R: 215, G: 95, B: 0},
		OrangeBold: ColorMap{R: 255, G: 135, B: 0},
	},
	"darwin": {
		Black:       ColorMap{R: 0, G: 0, B: 0},
		BlackBold:   ColorMap{R: 102, G: 102, B: 102},
		Blue:        ColorMap{R: 0, G: 0, B: 178},
		BlueBold:    ColorMap{R: 0, G: 0, B: 255},
		Cyan:        ColorMap{R: 0, G: 166, B: 178},
		CyanBold:    ColorMap{R: 0, G: 230, B: 230},
		Green:       ColorMap{R: 0, G: 166, B: 0},
		GreenBold:   ColorMap{R: 0, G: 217, B: 0},
		Magenta:     ColorMap{R: 178, G: 0, B: 178},
		MagentaBold: ColorMap{R: 230, G: 0, B: 230},
		Red:         ColorMap{R: 153, G: 0, B: 0},
		RedBold:     ColorMap{R: 230, G: 0, B: 0},
		White:       ColorMap{R: 191, G: 191, B: 191},
		WhiteBold:   ColorMap{R: 230, G: 230, B: 230},
		Yellow:      ColorMap{R: 153, G: 153, B: 0},
		YellowBold:  ColorMap{R: 230, G: 230, B: 0},
		// Not part of the standard 16 colors
		Orange:     ColorMap{R: 215, G: 95, B: 0},
		OrangeBold: ColorMap{R: 255, G: 135, B: 0},
	},
	"linux": {
		Black:       ColorMap{R: 1, G: 1, B: 1},
		BlackBold:   ColorMap{R: 128, G: 128, B: 128},
		Blue:        ColorMap{R: 0, G: 111, B: 184},
		BlueBold:    ColorMap{R: 0, G: 0, B: 255},
		Cyan:        ColorMap{R: 41, G: 181, B: 233},
		CyanBold:    ColorMap{R: 0, G: 255, B: 255},
		Green:       ColorMap{R: 57, G: 181, B: 74},
		GreenBold:   ColorMap{R: 0, G: 255, B: 0},
		Magenta:     ColorMap{R: 118, G: 38, B: 113},
		MagentaBold: ColorMap{R: 255, G: 0, B: 255},
		Red:         ColorMap{R: 222, G: 56, B: 43},
		RedBold:     ColorMap{R: 255, G: 0, B: 0},
		White:       ColorMap{R: 204, G: 204, B: 204},
		WhiteBold:   ColorMap{R: 255, G: 255, B: 255},
		Yellow:      ColorMap{R: 255, G: 199, B: 6},
		YellowBold:  ColorMap{R: 255, G: 255, B: 0},
		// Not part of the standard 16 colors
		Orange:     ColorMap{R: 215, G: 95, B: 0},
		OrangeBold: ColorMap{R: 255, G: 135, B: 0},
	},
	"xterm": {
		Black:       ColorMap{R: 0, G: 0, B: 0},
		BlackBold:   ColorMap{R: 127, G: 127, B: 127},
		Blue:        ColorMap{R: 0, G: 0, B: 238},
		BlueBold:    ColorMap{R: 92, G: 92, B: 255},
		Cyan:        ColorMap{R: 0, G: 205, B: 205},
		CyanBold:    ColorMap{R: 0, G: 255, B: 255},
		Green:       ColorMap{R: 0, G: 205, B: 0},
		GreenBold:   ColorMap{R: 0, G: 255, B: 0},
		Magenta:     ColorMap{R: 205, G: 0, B: 205},
		MagentaBold: ColorMap{R: 255, G: 0, B: 255},
		Red:         ColorMap{R: 205, G: 0, B: 0},
		RedBold:     ColorMap{R: 255, G: 0, B: 0},
		White:       ColorMap{R: 229, G: 229, B: 229},
		WhiteBold:   ColorMap{R: 255, G: 255, B: 255},
		Yellow:      ColorMap{R: 205, G: 205, B: 0},
		YellowBold:  ColorMap{R: 255, G: 255, B: 0},
		// Not part of the standard 16 colors
		Orange:     ColorMap{R: 215, G: 95, B: 0},
		OrangeBold: ColorMap{R: 255, G: 135, B: 0},
	},
	"ansi8": {
		Black:       ColorMap{Ansi: AnsiBlack},
		BlackBold:   ColorMap{Ansi: AnsiBlackBold},
		Blue:        ColorMap{Ansi: AnsiBlue},
		BlueBold:    ColorMap{Ansi: AnsiBlueBold},
		Cyan:        ColorMap{Ansi: AnsiCyan},
		CyanBold:    ColorMap{Ansi: AnsiCyanBold},
		Green:       ColorMap{Ansi: AnsiGreen},
		GreenBold:   ColorMap{Ansi: AnsiGreenBold},
		Magenta:     ColorMap{Ansi: AnsiMagenta},
		MagentaBold: ColorMap{Ansi: AnsiMagentaBold},
		Red:         ColorMap{Ansi: AnsiRed},
		RedBold:     ColorMap{Ansi: AnsiRedBold},
		White:       ColorMap{Ansi: AnsiWhite},
		WhiteBold:   ColorMap{Ansi: AnsiWhiteBold},
		Yellow:      ColorMap{Ansi: AnsiYellow},
		YellowBold:  ColorMap{Ansi: AnsiYellowBold},
	},
}
