// Package pstree provides functionality for building and displaying process trees.
//
// This file contains core process handling functions including process collection,
// sorting, and data transformation. It serves as the foundation for the process tree
// visualization by gathering and organizing the raw process data.
package pstree

import (
	"errors"
	"fmt"
	"log"
	"sort"

	"github.com/bananazon/pstree/util"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

//------------------------------------------------------------------------------
// PROCESS SORTING FUNCTIONS
//------------------------------------------------------------------------------
// Functions in this section handle sorting processes by various attributes.

// SortByPid sorts a slice of process.Process pointers by their PID in ascending order.
//
// Parameters:
//   - procs: Slice of process pointers to be sorted
//
// Returns:
//   - Sorted slice of process pointers
func SortByPid(procs []*process.Process) []*process.Process {
	sort.Slice(procs, func(i, j int) bool {
		return procs[i].Pid < procs[j].Pid // Ascending order
	})
	return procs
}

//------------------------------------------------------------------------------
// PROCESS LOOKUP FUNCTIONS
//------------------------------------------------------------------------------
// Functions in this section handle finding processes by specific attributes.

// GetProcessByPid finds and returns a process with the specified PID from the processes slice.
//
// Parameters:
//   - processes: Pointer to a slice of Process structs
//   - pid: The PID of the process to find
//
// Returns:
//   - The Process struct for the specified PID
//   - An error if the process with the given PID was not found
func GetProcessByPid(processes *[]Process, pid int32) (proc Process, err error) {
	for i := range *processes {
		if (*processes)[i].PID == pid {
			return (*processes)[i], nil
		}
	}
	errorMessage := fmt.Sprintf("the process with the PID %d was not found", pid)
	return Process{}, errors.New(errorMessage)
}

// SortProcsByAge sorts the processes slice by process age in ascending order.
//
// Parameters:
//   - processes: Pointer to a slice of Process structs to be sorted
func SortProcsByAge(processes *[]Process) {
	sort.Slice(*processes, func(i, j int) bool {
		return (*processes)[i].Age < (*processes)[j].Age
	})
}

// SortProcsByCpu sorts the processes slice by CPU usage percentage in ascending order.
//
// Parameters:
//   - processes: Pointer to a slice of Process structs to be sorted
func SortProcsByCpu(processes *[]Process) {
	sort.Slice(*processes, func(i, j int) bool {
		return (*processes)[i].CPUPercent < (*processes)[j].CPUPercent
	})
}

// SortProcsByMemory sorts the processes slice by memory usage (RSS) in ascending order.
//
// Parameters:
//   - processes: Pointer to a slice of Process structs to be sorted
func SortProcsByMemory(processes *[]Process) {
	sort.Slice(*processes, func(i, j int) bool {
		return float64((*processes)[i].MemoryInfo.RSS) < float64((*processes)[j].MemoryInfo.RSS)
	})
}

// SortProcsByUsername sorts the processes slice by username in ascending alphabetical order.
//
// Parameters:
//   - processes: Pointer to a slice of Process structs to be sorted
func SortProcsByUsername(processes *[]Process) {
	sort.Slice(*processes, func(i, j int) bool {
		return (*processes)[i].Username < (*processes)[j].Username
	})
}

// SortProcsByPid sorts the processes slice by PID in ascending order.
//
// Parameters:
//   - processes: Pointer to a slice of Process structs to be sorted
func SortProcsByPid(processes *[]Process) {
	sort.Slice(*processes, func(i, j int) bool {
		return (*processes)[i].PID < (*processes)[j].PID
	})
}

// SortProcsByNumThreads sorts the processes slice by the number of threads in ascending order.
//
// Parameters:
//   - processes: Pointer to a slice of Process structs to be sorted
func SortProcsByNumThreads(processes *[]Process) {
	sort.Slice(*processes, func(i, j int) bool {
		return (*processes)[i].NumThreads < (*processes)[j].NumThreads
	})
}

//------------------------------------------------------------------------------
// PROCESS DATA COLLECTION
//------------------------------------------------------------------------------
// Functions in this section handle gathering detailed process information.

// GenerateProcess creates a Process struct from a process.Process pointer.
// It collects various process attributes using goroutines and channels for concurrent execution
// to improve performance when gathering process information.
//
// Parameters:
//   - proc: Pointer to a process.Process struct from which to generate the Process
//
// Returns:
//   - A new Process struct populated with information from the input process
func GenerateProcess(proc *process.Process, miniOptions DisplayOptions) Process {
	var (
		args               []string
		background         bool
		children           []*process.Process
		command            string
		connections        []net.ConnectionStat
		cpuAffinity        []int32
		cpuPercent         float64
		cpuTimes           *cpu.TimesStat
		createTime         int64
		environment        []string
		err                error
		foreground         bool
		gids               []uint32
		groups             []uint32
		ioCounters         *process.IOCountersStat
		pageFaults         *process.PageFaultsStat
		pgid               int
		pid                int32
		ppid               int32
		memoryInfo         *process.MemoryInfoStat
		memoryInfoEx       *process.MemoryInfoExStat
		memoryPercent      float32
		numContextSwitches *process.NumCtxSwitchesStat
		numFDs             int32
		numThreads         int32
		openFiles          []process.OpenFilesStat
		resourceLimit      []process.RlimitStat
		resourceLimitUsage []process.RlimitStat
		status             []string
		threads            map[int32]*cpu.TimesStat
		uids               []uint32
		username           string
	)

	/*
	 * PID and Command are required fields
	 */
	pid = proc.Pid

	// We need to get the arguments so identical processes are grouped, even if arguments are not displayed
	argsChannel := make(chan func(proc *process.Process) (args []string, err error))
	go ProcessArgs(argsChannel)
	argsOut, err := (<-argsChannel)(proc)
	if err != nil {
		args = []string{}
	} else {
		args = argsOut
	}

	commandNameChannel := make(chan func(proc *process.Process) (string, error))
	go ProcessCommandName(commandNameChannel)
	commandOut, err := (<-commandNameChannel)(proc)
	if err != nil {
		command = "?"
	} else {
		command = commandOut
	}

	ppidChannel := make(chan func(proc *process.Process) (ppid int32, err error))
	go ProcessPPID(ppidChannel)
	ppidOut, err := (<-ppidChannel)(proc)
	if err != nil {
		ppid = -1
	} else {
		ppid = ppidOut
	}

	/*
	 * Only gather these if they're requested
	 */
	// This is very expensive so we'll ignore it for now
	// backgroundChannel := make(chan func(proc *process.Process) (background bool, err error))
	// go ProcessBackground(backgroundChannel)
	// backgroundOut, err := (<-backgroundChannel)(proc)
	// if err != nil {
	// 	background = false
	// } else {
	// 	background = backgroundOut
	// }

	// This is very expensive so we'll ignore it for now
	// childrenChannel := make(chan func(proc *process.Process) (children []*process.Process, err error))
	// go ProcessChildren(childrenChannel)
	// childrenOut, err := (<-childrenChannel)(proc)
	// if err != nil {
	// 	children = []*process.Process{}
	// } else {
	// 	children = childrenOut
	// }

	// This is very expensive so we'll ignore it for now
	// connectionsChannel := make(chan func(proc *process.Process) (connections []net.ConnectionStat, err error))
	// go ProcessConnections(connectionsChannel)
	// connectionsOut, err := (<-connectionsChannel)(proc)
	// if err != nil {
	// 	connections = []net.ConnectionStat{}
	// } else {
	// 	connections = connectionsOut
	// }

	// Not in use
	// cpuAffintyChannel := make(chan func(proc *process.Process) (affinity []int32, err error))
	// go ProcessCpuAffinity(cpuAffintyChannel)
	// cpuAffinityOut, err := (<-cpuAffintyChannel)(proc)
	// if err != nil {
	// 	cpuAffinity = []int32{}
	// } else {
	// 	cpuAffinity = cpuAffinityOut
	// }

	if miniOptions.ShowCpuPercent || miniOptions.OrderBy == "cpu" || miniOptions.ColorAttr == "cpu" {
		cpuPercentChannel := make(chan func(proc *process.Process) (cpuPercent float64, err error))
		go ProcessCpuPercent(cpuPercentChannel)
		cpuPercentOut, err := (<-cpuPercentChannel)(proc)
		if err != nil {
			cpuPercent = -1
		} else {
			cpuPercent = cpuPercentOut
		}
	}

	// Not in use
	// cpuTimesChannel := make(chan func(proc *process.Process) (cpuTimes *cpu.TimesStat, err error))
	// go ProcessCpuTimes(cpuTimesChannel)
	// cpuTimesOut, err := (<-cpuTimesChannel)(proc)
	// if err != nil {
	// 	cpuTimes = &cpu.TimesStat{}
	// } else {
	// 	cpuTimes = cpuTimesOut
	// }

	if miniOptions.ShowProcessAge || miniOptions.OrderBy == "age" || miniOptions.ColorAttr == "age" {
		createTimeChannel := make(chan func(proc *process.Process) (createTime int64, err error))
		go ProcessCreateTime(createTimeChannel)
		createTimeOut, err := (<-createTimeChannel)(proc)
		if err != nil {
			createTime = -1
		} else {
			createTime = createTimeOut
		}
	}

	// Not in use
	// environmentChannel := make(chan func(proc *process.Process) (environment []string, err error))
	// go ProcessEnvironment(environmentChannel)
	// environmentOut, err := (<-environmentChannel)(proc)
	// if err != nil {
	// 	environment = []string{}
	// } else {
	// 	environment = environmentOut
	// }

	// This is very expensive so we'll ignore it for now
	// foregroundChannel := make(chan func(proc *process.Process) (foreground bool, err error))
	// go ProcessForeground(foregroundChannel)
	// foregroundOut, err := (<-foregroundChannel)(proc)
	// if err != nil {
	// 	foreground = false
	// } else {
	// 	foreground = foregroundOut
	// }

	gidsChannel := make(chan func(proc *process.Process) (gids []uint32, err error))
	go ProcessGIDs(gidsChannel)
	gidsOut, err := (<-gidsChannel)(proc)
	if err != nil {
		gids = []uint32{}
	} else {
		gids = gidsOut
	}

	groupsChannel := make(chan func(proc *process.Process) (groups []uint32, err error))
	go ProcessGroups(groupsChannel)
	groupsOut, err := (<-groupsChannel)(proc)
	if err != nil {
		groups = []uint32{}
	} else {
		groups = groupsOut
	}

	// Not in use
	// ioCountersChannel := make(chan func(proc *process.Process) (ioCounters *process.IOCountersStat, err error))
	// go ProcessIOCounters(ioCountersChannel)
	// ioCountersOut, err := (<-ioCountersChannel)(proc)
	// if err != nil {
	// 	ioCounters = &process.IOCountersStat{}
	// } else {
	// 	ioCounters = ioCountersOut
	// }

	if miniOptions.ShowMemoryUsage || miniOptions.OrderBy == "mem" || miniOptions.ColorAttr == "mem" {
		memoryInfoChannel := make(chan func(proc *process.Process) (memoryInfo *process.MemoryInfoStat, err error))
		go ProcessMemoryInfo(memoryInfoChannel)
		memoryInfoOut, err := (<-memoryInfoChannel)(proc)
		if err != nil {
			memoryInfo = &process.MemoryInfoStat{}
		} else {
			memoryInfo = memoryInfoOut
		}

		memoryInfoExChannel := make(chan func(proc *process.Process) (memoryInfoEx *process.MemoryInfoExStat, err error))
		go ProcessMemoryInfoEx(memoryInfoExChannel)
		memoryInfoExOut, err := (<-memoryInfoExChannel)(proc)
		if err != nil {
			memoryInfoEx = &process.MemoryInfoExStat{}
		} else {
			memoryInfoEx = memoryInfoExOut
		}

		memoryPercentChannel := make(chan func(proc *process.Process) (memoryPercent float32, err error))
		go ProcessMemoryPercent(memoryPercentChannel)
		memoryPercentOut, err := (<-memoryPercentChannel)(proc)
		if err != nil {
			memoryPercent = -1.0
		} else {
			memoryPercent = memoryPercentOut
		}
	}

	numCtxSwitchesChannel := make(chan func(proc *process.Process) (numContextSwitches *process.NumCtxSwitchesStat, err error))
	go ProcessNumCtxSwitches(numCtxSwitchesChannel)
	numContextSwitchesOut, err := (<-numCtxSwitchesChannel)(proc)
	if err != nil {
		numContextSwitches = &process.NumCtxSwitchesStat{}
	} else {
		numContextSwitches = numContextSwitchesOut
	}

	// Not in use
	// numFDsChannel := make(chan func(proc *process.Process) (numFDs int32, err error))
	// go ProcessNumFDs(numFDsChannel)
	// numFDsOut, err := (<-numFDsChannel)(proc)
	// if err != nil {
	// 	numFDs = -1
	// } else {
	// 	numFDs = numFDsOut
	// }

	if miniOptions.ShowNumThreads || miniOptions.OrderBy == "threads" {
		numThreadsChannel := make(chan func(proc *process.Process) (numThreads int32, err error))
		go ProcessNumThreads(numThreadsChannel)
		numThreadsOut, err := (<-numThreadsChannel)(proc)
		if err != nil {
			numThreads = -1
		} else {
			numThreads = numThreadsOut
		}
	}

	// Not in use
	// openFilesChannel := make(chan func(proc *process.Process) (openFiles []process.OpenFilesStat, err error))
	// go ProcessOpenFiles(openFilesChannel)
	// openFilesOut, err := (<-openFilesChannel)(proc)
	// if err != nil {
	// 	openFiles = []process.OpenFilesStat{}
	// } else {
	// 	openFiles = openFilesOut
	// }

	// Not in use
	// pageFaultsChannel := make(chan func(proc *process.Process) (pageFaults *process.PageFaultsStat, err error))
	// go ProcessPageFaults(pageFaultsChannel)
	// pageFaultsOut, err := (<-pageFaultsChannel)(proc)
	// if err != nil {
	// 	pageFaults = &process.PageFaultsStat{}
	// } else {
	// 	pageFaults = pageFaultsOut
	// }

	if miniOptions.ShowPGIDs || miniOptions.ShowPGLs {
		pgidChannel := make(chan func(proc *process.Process) (pgid int, err error))
		go ProcessPGID(pgidChannel)
		pgidOut, err := (<-pgidChannel)(proc)
		if err != nil {
			pgid = -1
		} else {
			pgid = pgidOut
		}
	}

	// Not in use
	// resourceLimitChannel := make(chan func(proc *process.Process) (resourceLimit []process.RlimitStat, err error))
	// go ProcessResourceLimit(resourceLimitChannel)
	// resourceLimitOut, err := (<-resourceLimitChannel)(proc)
	// if err != nil {
	// 	resourceLimit = []process.RlimitStat{}
	// } else {
	// 	resourceLimit = resourceLimitOut
	// }

	// Not in use
	// resourceLimitUsageChannel := make(chan func(proc *process.Process) (resourceLimitUsage []process.RlimitStat, err error))
	// go ProcessResourceLimitUsage(resourceLimitUsageChannel)
	// resourceLimitUsageOut, err := (<-resourceLimitUsageChannel)(proc)
	// if err != nil {
	// 	resourceLimitUsage = []process.RlimitStat{}
	// } else {
	// 	resourceLimitUsage = resourceLimitUsageOut
	// }

	// This is very expensive so we'll ignore it for now
	// statusChannel := make(chan func(proc *process.Process) (status []string, err error))
	// go ProcessStatus(statusChannel)
	// statusOut, err := (<-statusChannel)(proc)
	// if err != nil {
	// 	status = []string{}
	// } else {
	// 	status = statusOut
	// }

	// Not in use
	// threadsChannel := make(chan func(proc *process.Process) (threads map[int32]*cpu.TimesStat, err error))
	// go ProcessThreads(threadsChannel)
	// threadsOut, err := (<-threadsChannel)(proc)
	// if err != nil {
	// 	threads = map[int32]*cpu.TimesStat{}
	// } else {
	// 	threads = threadsOut
	// }

	if miniOptions.ShowOwner || miniOptions.ShowUserTransitions || miniOptions.OrderBy == "user" {
		usernameChannel := make(chan func(proc *process.Process) (username string, err error))
		go ProcessUsername(usernameChannel)
		usernameOut, err := (<-usernameChannel)(proc)
		if err != nil {
			username = "?"
		} else {
			username = usernameOut
		}
	}

	if miniOptions.ShowUIDTransitions || miniOptions.ShowUserTransitions {
		uidsChannel := make(chan func(proc *process.Process) (uids []uint32, err error))
		go ProcessUIDs(uidsChannel)
		uidsOut, err := (<-uidsChannel)(proc)
		if err != nil {
			uids = []uint32{}
		} else {
			uids = uidsOut
		}
	}

	if len(args) > 0 {
		if args[0] == command {
			if len(args) == 1 {
				args = []string{}
			} else if len(args) > 1 {
				args = args[1:]
			}
		}
	}

	return Process{
		Age:                util.GetUnixTimestamp() - createTime,
		Args:               args,
		Background:         background,
		Child:              -1,
		Children:           children,
		Command:            command,
		Connections:        connections,
		CPUAffinity:        cpuAffinity,
		CPUPercent:         util.RoundFloat(cpuPercent, 2),
		CPUTimes:           cpuTimes,
		CreateTime:         createTime,
		Environment:        environment,
		Foreground:         foreground,
		GIDs:               gids,
		Groups:             groups,
		IOCounters:         ioCounters,
		MemoryInfo:         memoryInfo,
		MemoryInfoEx:       memoryInfoEx,
		MemoryPercent:      memoryPercent,
		NumContextSwitches: numContextSwitches,
		NumFDs:             numFDs,
		NumThreads:         numThreads,
		OpenFiles:          openFiles,
		PageFaults:         pageFaults,
		Parent:             -1,
		PGID:               int32(pgid),
		PID:                pid,
		PPID:               ppid,
		ResourceLimit:      resourceLimit,
		ResourceLimitUsage: resourceLimitUsage,
		Sister:             -1,
		Status:             status,
		Threads:            threads,
		UIDs:               uids,
		Username:           username,
	}
}

// GetProcesses retrieves all system processes and populates the provided processes slice.
//
// This function uses the gopsutil library to get a list of all processes running on the system,
// sorts them by PID, and then generates detailed Process structs for each one using the
// generateProcess function.
//
// Parameters:
//   - processes: A pointer to a slice that will be populated with Process structs
//   - flagOrderBy: A string indicating the order by which to sort the processes
//   - miniOptions: A pointer to a MiniOptions struct containing options for the process tree
func GetProcesses(processes *[]Process, miniOptions DisplayOptions) {
	var (
		err      error
		sorted   []*process.Process
		unsorted []*process.Process
	)
	unsorted, err = process.Processes()
	if err != nil {
		log.Fatalf("Failed to get processes: %v", err)
	}

	sorted = SortByPid(unsorted)

	for _, p := range sorted {
		*processes = append(*processes, GenerateProcess(p, miniOptions))
	}
}
