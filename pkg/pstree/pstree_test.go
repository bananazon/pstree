package pstree

import (
	"testing"

	"github.com/shirou/gopsutil/v4/process"
	"github.com/stretchr/testify/assert"
)

func TestSortByPid(t *testing.T) {
	// Create test processes with different PIDs
	proc1 := &process.Process{Pid: 100}
	proc2 := &process.Process{Pid: 50}
	proc3 := &process.Process{Pid: 200}

	// Create a slice with processes in random order
	procs := []*process.Process{proc1, proc2, proc3}

	// Sort the processes
	sortedProcs := SortByPid(procs)

	// Verify that the processes are sorted by PID in ascending order
	assert.Equal(t, int32(50), sortedProcs[0].Pid)
	assert.Equal(t, int32(100), sortedProcs[1].Pid)
	assert.Equal(t, int32(200), sortedProcs[2].Pid)
}

func TestGetProcessByPid(t *testing.T) {
	// Create test processes
	proc1 := Process{PID: 100, Command: "proc1"}
	proc2 := Process{PID: 200, Command: "proc2"}
	proc3 := Process{PID: 300, Command: "proc3"}

	// Create a slice with the processes
	processes := []Process{proc1, proc2, proc3}

	// Test finding an existing process
	foundProc, err := GetProcessByPid(&processes, 200)
	assert.NoError(t, err)
	assert.Equal(t, "proc2", foundProc.Command)

	// Test finding a non-existent process
	_, err = GetProcessByPid(&processes, 999)
	assert.Error(t, err)
}

func TestSortProcsByAge(t *testing.T) {
	// Create test processes with different ages
	proc1 := Process{PID: 100, Age: 300}
	proc2 := Process{PID: 200, Age: 100}
	proc3 := Process{PID: 300, Age: 200}

	// Create a slice with the processes
	processes := []Process{proc1, proc2, proc3}

	// Sort the processes by age
	SortProcsByAge(&processes)

	// Verify that the processes are sorted by age in ascending order
	assert.Equal(t, int64(100), processes[0].Age)
	assert.Equal(t, int64(200), processes[1].Age)
	assert.Equal(t, int64(300), processes[2].Age)
}

func TestSortProcsByCpu(t *testing.T) {
	// Create test processes with different CPU percentages
	proc1 := Process{PID: 100, CPUPercent: 5.0}
	proc2 := Process{PID: 200, CPUPercent: 1.0}
	proc3 := Process{PID: 300, CPUPercent: 10.0}

	// Create a slice with the processes
	processes := []Process{proc1, proc2, proc3}

	// Sort the processes by CPU percentage
	SortProcsByCpu(&processes)

	// Verify that the processes are sorted by CPU percentage in ascending order
	assert.Equal(t, float64(1.0), processes[0].CPUPercent)
	assert.Equal(t, float64(5.0), processes[1].CPUPercent)
	assert.Equal(t, float64(10.0), processes[2].CPUPercent)
}

func TestSortProcsByMemory(t *testing.T) {
	// Create test processes with different memory usage
	proc1 := Process{PID: 100, MemoryInfo: &process.MemoryInfoStat{RSS: 5000}}
	proc2 := Process{PID: 200, MemoryInfo: &process.MemoryInfoStat{RSS: 1000}}
	proc3 := Process{PID: 300, MemoryInfo: &process.MemoryInfoStat{RSS: 10000}}

	// Create a slice with the processes
	processes := []Process{proc1, proc2, proc3}

	// Sort the processes by memory usage
	SortProcsByMemory(&processes)

	// Verify that the processes are sorted by memory usage in ascending order
	assert.Equal(t, uint64(1000), processes[0].MemoryInfo.RSS)
	assert.Equal(t, uint64(5000), processes[1].MemoryInfo.RSS)
	assert.Equal(t, uint64(10000), processes[2].MemoryInfo.RSS)
}

func TestSortProcsByUsername(t *testing.T) {
	// Create test processes with different usernames
	proc1 := Process{PID: 100, Username: "charlie"}
	proc2 := Process{PID: 200, Username: "alice"}
	proc3 := Process{PID: 300, Username: "bob"}

	// Create a slice with the processes
	processes := []Process{proc1, proc2, proc3}

	// Sort the processes by username
	SortProcsByUsername(&processes)

	// Verify that the processes are sorted by username in ascending alphabetical order
	assert.Equal(t, "alice", processes[0].Username)
	assert.Equal(t, "bob", processes[1].Username)
	assert.Equal(t, "charlie", processes[2].Username)
}

func TestSortProcsByPid(t *testing.T) {
	// Create test processes with different PIDs
	proc1 := Process{PID: 300}
	proc2 := Process{PID: 100}
	proc3 := Process{PID: 200}

	// Create a slice with the processes
	processes := []Process{proc1, proc2, proc3}

	// Sort the processes by PID
	SortProcsByPid(&processes)

	// Verify that the processes are sorted by PID in ascending order
	assert.Equal(t, int32(100), processes[0].PID)
	assert.Equal(t, int32(200), processes[1].PID)
	assert.Equal(t, int32(300), processes[2].PID)
}

func TestSortProcsByNumThreads(t *testing.T) {
	// Create test processes with different thread counts
	proc1 := Process{PID: 100, NumThreads: 5}
	proc2 := Process{PID: 200, NumThreads: 2}
	proc3 := Process{PID: 300, NumThreads: 10}

	// Create a slice with the processes
	processes := []Process{proc1, proc2, proc3}

	// Sort the processes by thread count
	SortProcsByNumThreads(&processes)

	// Verify that the processes are sorted by thread count in ascending order
	assert.Equal(t, int32(2), processes[0].NumThreads)
	assert.Equal(t, int32(5), processes[1].NumThreads)
	assert.Equal(t, int32(10), processes[2].NumThreads)
}

func TestGenerateProcess(t *testing.T) {
	// This is a more complex test that requires mocking the process.Process type
	// For simplicity, we'll just verify that the function doesn't panic

	// Create a minimal process.Process with just a PID
	// In a real test, you would use a mocking library to mock the methods
	proc := &process.Process{Pid: 1}

	// Call generateProcess and verify it doesn't panic
	result := GenerateProcess(proc, DisplayOptions{})

	// Basic verification that the result has the expected PID
	assert.Equal(t, int32(1), result.PID)
}
