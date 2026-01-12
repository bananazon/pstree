package pstree

import (
	"testing"

	"github.com/shirou/gopsutil/v4/process"
	"github.com/stretchr/testify/assert"
)

func TestProcessMetricsFunctions(t *testing.T) {
	// Test ProcessArgs
	t.Run("ProcessArgs", func(t *testing.T) {
		ch := make(chan func(proc *process.Process) (args []string, err error))
		go ProcessArgs(ch)
		fn := <-ch

		// We can't easily test the actual function behavior without a real process
		// But we can verify that a function was sent through the channel
		assert.NotNil(t, fn)
	})

	// Test ProcessCommandName
	t.Run("ProcessCommandName", func(t *testing.T) {
		ch := make(chan func(proc *process.Process) (string, error))
		go ProcessCommandName(ch)
		fn := <-ch

		assert.NotNil(t, fn)
	})

	// Test ProcessCpuPercent
	t.Run("ProcessCpuPercent", func(t *testing.T) {
		ch := make(chan func(proc *process.Process) (float64, error))
		go ProcessCpuPercent(ch)
		fn := <-ch

		assert.NotNil(t, fn)
	})

	// Test ProcessCreateTime
	t.Run("ProcessCreateTime", func(t *testing.T) {
		ch := make(chan func(proc *process.Process) (int64, error))
		go ProcessCreateTime(ch)
		fn := <-ch

		assert.NotNil(t, fn)
	})

	// Test ProcessMemoryInfo
	t.Run("ProcessMemoryInfo", func(t *testing.T) {
		ch := make(chan func(proc *process.Process) (memoryInfo *process.MemoryInfoStat, err error))
		go ProcessMemoryInfo(ch)
		fn := <-ch

		assert.NotNil(t, fn)
	})

	// Test ProcessNumThreads
	t.Run("ProcessNumThreads", func(t *testing.T) {
		ch := make(chan func(proc *process.Process) (numThreads int32, err error))
		go ProcessNumThreads(ch)
		fn := <-ch

		assert.NotNil(t, fn)
	})

	// Test ProcessUsername
	t.Run("ProcessUsername", func(t *testing.T) {
		ch := make(chan func(proc *process.Process) (username string, err error))
		go ProcessUsername(ch)
		fn := <-ch

		assert.NotNil(t, fn)
	})

	// Test ProcessUIDs
	t.Run("ProcessUIDs", func(t *testing.T) {
		ch := make(chan func(proc *process.Process) (uids []uint32, err error))
		go ProcessUIDs(ch)
		fn := <-ch

		assert.NotNil(t, fn)
	})

	// Test ProcessPPID
	t.Run("ProcessPPID", func(t *testing.T) {
		ch := make(chan func(proc *process.Process) (ppid int32, err error))
		go ProcessPPID(ch)
		fn := <-ch

		assert.NotNil(t, fn)
	})
}
