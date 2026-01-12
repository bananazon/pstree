package pstree

import (
	"log/slog"
	"testing"
)

// BenchmarkBuildTree benchmarks the BuildTree function with different numbers of processes
func BenchmarkBuildTree(b *testing.B) {
	// Create a logger that discards output
	logger := slog.New(slog.NewTextHandler(nil, nil))

	// Define benchmark cases with different numbers of processes
	benchCases := []struct {
		name      string
		numProcs  int
		maxDepth  int // Maximum depth of the process tree
		branching int // Average number of children per process
	}{
		{"Small_10", 10, 3, 2},
		{"Medium_100", 100, 4, 3},
		{"Large_1000", 1000, 5, 4},
		{"Xlarge_10000", 10000, 6, 5},
	}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			// Generate test processes
			processes := generateTestProcesses(bc.numProcs, bc.maxDepth, bc.branching)

			// Reset the timer before the benchmark loop
			b.ResetTimer()

			// Run the benchmark
			for i := 0; i < b.N; i++ {
				// Create a process tree
				processTree := &ProcessTree{
					Logger:         logger,
					Nodes:          processes,
					PidToIndexMap:  make(map[int32]int),
					IndexToPidMap:  make(map[int]int32),
					DisplayOptions: DisplayOptions{},
				}

				// Initialize PID to index mapping
				for i, proc := range processes {
					processTree.PidToIndexMap[proc.PID] = i
					processTree.IndexToPidMap[i] = proc.PID
				}

				// Build the tree - this is what we're benchmarking
				processTree.BuildTree()
			}
		})
	}
}

// generateTestProcesses creates a slice of test processes with a realistic hierarchy
func generateTestProcesses(numProcs, maxDepth, branching int) []Process {
	processes := make([]Process, 0, numProcs)

	// Always start with init process (PID 1)
	processes = append(processes, Process{
		PID:     1,
		PPID:    0,
		Command: "init",
	})

	// Generate the rest of the processes
	pid := int32(2)
	currentDepth := 1
	parentIndex := 0

	for len(processes) < numProcs {
		// Calculate how many children this parent should have
		numChildren := min(branching, numProcs-len(processes))
		if numChildren <= 0 {
			break
		}

		// Create children for this parent
		parentPID := processes[parentIndex].PID
		for i := 0; i < numChildren; i++ {
			if len(processes) >= numProcs {
				break
			}

			processes = append(processes, Process{
				PID:     pid,
				PPID:    parentPID,
				Command: "proc" + string(rune('A'+i%26)),
			})
			pid++
		}

		// Move to the next parent
		parentIndex++

		// If we've processed all parents at this depth, move to the next depth
		if parentIndex >= len(processes) || processes[parentIndex].PPID != processes[parentIndex-1].PPID {
			currentDepth++
			if currentDepth > maxDepth {
				break
			}
		}
	}

	return processes
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
