package pstree

import (
	"fmt"
	"syscall"

	"github.com/bananazon/pstree/pkg/globals"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

// ProcessArgs sends a function to the provided channel that retrieves command line arguments for a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessArgs(c chan func(proc *process.Process) (args []string, err error)) {
	c <- (func(proc *process.Process) (args []string, err error) {
		args, err = proc.CmdlineSlice()
		return args, err
	})
}

// ProcessBackground sends a function to the provided channel that retrieves true if the process is in the background.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessBackground(c chan func(proc *process.Process) (background bool, err error)) {
	c <- (func(proc *process.Process) (background bool, err error) {
		background, err = proc.Background()
		return background, err
	})
}

// ProcessCommandName sends a function to the provided channel that retrieves the executable path of a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessCommandName(c chan func(proc *process.Process) (string, error)) {
	c <- (func(proc *process.Process) (command string, err error) {
		// First check for exe, which should be the full path to the
		exe, err := proc.Exe()
		if err == nil && exe != "" {
			// Return the full path
			if globals.GetDebugLevel() > 1 {
				globals.GetLogger().Debug(fmt.Sprintf("ProcessCommandName, PID %d (ExeWithContext): %s", proc.Pid, exe))
			}
			return exe, nil
		}

		// Either there was en error or exe was empty so let's try to get the command slice
		cmdLine, err := proc.CmdlineSlice()
		if err == nil && len(cmdLine) > 0 {
			// Return the first element of the command line slice, which is the executable
			if globals.GetDebugLevel() > 1 {
				globals.GetLogger().Debug(fmt.Sprintf("ProcessCommandName, PID %d (CmdlineSliceWithContext): %s", proc.Pid, cmdLine[0]))
			}
			return cmdLine[0], nil
		}

		// Crud, we don't have a command name so let's try to get the command basename
		name, err := proc.Name()
		if err == nil && name != "" {
			// Return name, which is the basename of the command
			if globals.GetDebugLevel() > 1 {
				globals.GetLogger().Debug(fmt.Sprintf("ProcessCommandName, PID %d (NameWithContext): %s", proc.Pid, name))
			}
			return name, nil
		}

		// Well crap, I give up, let's return the PID
		if globals.GetDebugLevel() > 1 {
			globals.GetLogger().Debug(fmt.Sprintf("ProcessCommandName, PID %d (PID): %d", proc.Pid, proc.Pid))
		}
		return fmt.Sprintf("[PID %d]", proc.Pid), nil
	})
}

// ProcessChildren sends a function to the provided channel that retrieves a slice of child processes for a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessChildren(c chan func(proc *process.Process) (children []*process.Process, err error)) {
	c <- (func(proc *process.Process) (children []*process.Process, err error) {
		children, err = proc.Children()
		return children, err
	})
}

// ProcessConnections sends a function to the provided channel that retrieves network connections for a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessConnections(c chan func(proc *process.Process) (connections []net.ConnectionStat, err error)) {
	c <- (func(proc *process.Process) (connections []net.ConnectionStat, err error) {
		connections, err = proc.Connections()
		return connections, err
	})
}

// ProcessCpuAffinity sends a function to the provided channel that retrieves CPU affinty for a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessCpuAffinity(c chan func(proc *process.Process) (cpuAffinity []int32, err error)) {
	c <- (func(proc *process.Process) (cpuAffinity []int32, err error) {
		cpuAffinity, err = proc.CPUAffinity()
		return cpuAffinity, err
	})
}

// ProcessCpuPercent sends a function to the provided channel that retrieves CPU usage percentage for a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessCpuPercent(c chan func(proc *process.Process) (cpuPercent float64, err error)) {
	c <- (func(proc *process.Process) (cpuPercent float64, err error) {
		cpuPercent, err = proc.CPUPercent()
		return cpuPercent, err
	})
}

// ProcessCpuTimes sends a function to the provided channel that retrieves CPU times for a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessCpuTimes(c chan func(proc *process.Process) (cpuTimes *cpu.TimesStat, err error)) {
	c <- (func(proc *process.Process) (cpuTimes *cpu.TimesStat, err error) {
		cpuTimes, err = proc.Times()
		return cpuTimes, err
	})
}

// ProcessCreateTime sends a function to the provided channel that retrieves the creation time of a process.
// This function is designed to be used with goroutines to gather process information concurrently.
// The creation time is converted from milliseconds to seconds before being returned.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessCreateTime(c chan func(proc *process.Process) (createTime int64, err error)) {
	c <- (func(proc *process.Process) (createTime int64, err error) {
		createTime, err = proc.CreateTime()
		return createTime / 1000, err
	})
}

// ProcessEnvironment sends a function to the provided channel that retrieves environment variables for a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessEnvironment(c chan func(proc *process.Process) (environment []string, err error)) {
	c <- (func(proc *process.Process) (environment []string, err error) {
		environment, err = proc.Environ()
		return environment, err
	})
}

// ProcessForeground sends a function to the provided channel that retrieves true if the process is in the foreground.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessForeground(c chan func(proc *process.Process) (foreground bool, err error)) {
	c <- (func(proc *process.Process) (foreground bool, err error) {
		foreground, err = proc.Foreground()
		return foreground, err
	})
}

// ProcessGIDs sends a function to the provided channel that retrieves group IDs for a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessGIDs(c chan func(proc *process.Process) (gids []uint32, err error)) {
	c <- (func(proc *process.Process) (gids []uint32, err error) {
		gids, err = proc.Gids()
		return gids, err
	})
}

// ProcessGroups sends a function to the provided channel that retrieves supplementary group IDs for a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessGroups(c chan func(proc *process.Process) (groups []uint32, err error)) {
	c <- (func(proc *process.Process) (groups []uint32, err error) {
		groups, err = proc.Groups()
		return groups, err
	})
}

// ProcessIOCounters sends a function to the provided channel that retrieves IO counters for a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessIOCounters(c chan func(proc *process.Process) (ioCounters *process.IOCountersStat, err error)) {
	c <- (func(proc *process.Process) (ioCounters *process.IOCountersStat, err error) {
		ioCounters, err = proc.IOCounters()
		return ioCounters, err
	})
}

// ProcessMemoryInfo sends a function to the provided channel that retrieves memory usage statistics for a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessMemoryInfo(c chan func(proc *process.Process) (memoryInfo *process.MemoryInfoStat, err error)) {
	c <- (func(proc *process.Process) (memoryInfo *process.MemoryInfoStat, err error) {
		memoryInfo, err = proc.MemoryInfo()
		return memoryInfo, err
	})
}

// ProcessMemoryInfoEx sends a function to the provided channel that retrieves platform-spememory usage statistics for a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessMemoryInfoEx(c chan func(proc *process.Process) (memoryInfoEx *process.MemoryInfoExStat, err error)) {
	c <- (func(proc *process.Process) (memoryInfoEx *process.MemoryInfoExStat, err error) {
		memoryInfoEx, err = proc.MemoryInfoEx()
		return memoryInfoEx, err
	})
}

// ProcessMemoryPercent sends a function to the provided channel that retrieves memory usage percentage for a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessMemoryPercent(c chan func(proc *process.Process) (memoryPercent float32, err error)) {
	c <- (func(proc *process.Process) (memoryPercent float32, err error) {
		memoryPercent, err = proc.MemoryPercent()
		return memoryPercent, err
	})
}

// ProcessNumCtxSwitches sends a function to the provided channel that retrieves the number of context switches for process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessNumCtxSwitches(c chan func(proc *process.Process) (numContextSwitches *process.NumCtxSwitchesStat, err error)) {
	c <- (func(proc *process.Process) (numContextSwitches *process.NumCtxSwitchesStat, err error) {
		numContextSwitches, err = proc.NumCtxSwitches()
		return numContextSwitches, err
	})
}

// ProcessNumFDs sends a function to the provided channel that retrieves the number of file descriptors used by a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessNumFDs(c chan func(proc *process.Process) (numFDs int32, err error)) {
	c <- (func(proc *process.Process) (numFDs int32, err error) {
		numFDs, err = proc.NumFDs()
		return numFDs, err
	})
}

// ProcessNumThreads sends a function to the provided channel that retrieves the number of threads used by a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessNumThreads(c chan func(proc *process.Process) (numThreads int32, err error)) {
	c <- (func(proc *process.Process) (numThreads int32, err error) {
		numThreads, err = proc.NumThreads()
		return numThreads, err
	})
}

// ProcessOpenFiles sends a function to the provided channel that retrieves a slice of OpenFiles used by a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessOpenFiles(c chan func(proc *process.Process) (openFilesStat []process.OpenFilesStat, err error)) {
	c <- (func(proc *process.Process) (openFilesStat []process.OpenFilesStat, err error) {
		openFilesStat, err = proc.OpenFiles()
		return openFilesStat, err
	})
}

// ProcessPageFaults sends a function to the provided channel that retrieves pagefaults for a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessPageFaults(c chan func(proc *process.Process) (pageFaults *process.PageFaultsStat, err error)) {
	c <- (func(proc *process.Process) (pageFaults *process.PageFaultsStat, err error) {
		pageFaults, err = proc.PageFaults()
		return pageFaults, err
	})
}

// ProcessParent sends a function to the provided channel that retrieves the parent process of a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessParent(c chan func(proc *process.Process) (parent *process.Process, err error)) {
	c <- (func(proc *process.Process) (parent *process.Process, err error) {
		parent, err = proc.Parent()
		return parent, err
	})
}

// ProcessPGID sends a function to the provided channel that retrieves the process group ID of a process.
// This function is designed to be used with goroutines to gather process information concurrently.
// Unlike other functions, this one uses syscall.Getpgid directly instead of a context-aware method.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessPGID(c chan func(proc *process.Process) (pgid int, err error)) {
	c <- (func(proc *process.Process) (pgid int, err error) {
		pgid, err = syscall.Getpgid(int(proc.Pid))
		return pgid, err
	})
}

// ProcessPPID sends a function to the provided channel that retrieves the parent process ID of a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessPPID(c chan func(proc *process.Process) (ppid int32, err error)) {
	c <- (func(proc *process.Process) (ppid int32, err error) {
		ppid, err = proc.Ppid()
		return ppid, err
	})
}

// ProcessResourceLimit sends a function to the provided channel that retrieves resource limits of a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessResourceLimit(c chan func(proc *process.Process) (resourceLimit []process.RlimitStat, err error)) {
	c <- (func(proc *process.Process) (resourceLimit []process.RlimitStat, err error) {
		resourceLimit, err = proc.Rlimit()
		return resourceLimit, err
	})
}

// ProcessResourceLimitUsage sends a function to the provided channel that retrieves resource limits of a process.
// If gatherUsed is true, the currently used value will be gathered and added to the resulting RlimitStat.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessResourceLimitUsage(c chan func(proc *process.Process) (resourceLimitUsage []process.RlimitStat, err error)) {
	c <- (func(proc *process.Process) (resourceLimitUsage []process.RlimitStat, err error) {
		resourceLimitUsage, err = proc.RlimitUsage(true)
		return resourceLimitUsage, err
	})
}

// ProcessStatus sends a function to the provided channel that retrieves the status of a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessStatus(c chan func(proc *process.Process) (status []string, err error)) {
	c <- (func(proc *process.Process) (status []string, err error) {
		status, err = proc.Status()
		return status, err
	})
}

// ProcessThreads sends a function to the provided channel that retrieves the threads of a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessThreads(c chan func(proc *process.Process) (threads map[int32]*cpu.TimesStat, err error)) {
	c <- (func(proc *process.Process) (threads map[int32]*cpu.TimesStat, err error) {
		threads, err = proc.Threads()
		return threads, err
	})
}

// ProcessUsername sends a function to the provided channel that retrieves the username of the process owner.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessUsername(c chan func(proc *process.Process) (username string, err error)) {
	c <- (func(proc *process.Process) (username string, err error) {
		username, err = proc.Username()
		return username, err
	})
}

// ProcessUIDs sends a function to the provided channel that retrieves user IDs for a process.
// This function is designed to be used with goroutines to gather process information concurrently.
//
// Parameters:
//   - c: Channel to send the function through
func ProcessUIDs(c chan func(proc *process.Process) (uids []uint32, err error)) {
	c <- (func(proc *process.Process) (uids []uint32, err error) {
		uids, err = proc.Uids()
		return uids, err
	})
}
