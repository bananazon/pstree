package cmd

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var binaryPath string

func configureLogger() {
	log.SetPrefix("INFO: ")
	log.SetFlags(0)
}

func findGoModuleRoot(startDir string) (string, error) {
	configureLogger()
	dir := startDir
	log.Printf("Finding go module root from: %s", dir)

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			log.Printf("Found go module root at: %s", dir)
			return dir, nil
		}

		// Reached root without finding go.mod
		parent := filepath.Dir(dir)
		if parent == dir {
			log.Printf("go.mod not found")
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

func TestMain(m *testing.M) {
	configureLogger()

	// Use the same binary path as the Makefile
	absPath, err := filepath.Abs(".")
	if err != nil {
		panic("failed to find absolute path: " + err.Error())
	}

	goModuleRoot, err := findGoModuleRoot(absPath)
	binaryPath = filepath.Join(goModuleRoot, "pstree.testbin")
	if err != nil {
		panic("failed to find go module root: " + err.Error())
	}

	log.Printf("Binary path: %s", binaryPath)

	// The binary should already be built by the Makefile
	// We'll just verify it exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		// If it doesn't exist, build it
		log.Printf("Binary not found, building it now")

		// Ensure directory exists
		dir := filepath.Dir(binaryPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			panic("failed to create directory: " + err.Error())
		}

		cmd := exec.Command("go", "build", "-o", binaryPath, goModuleRoot)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			panic("failed to build binary: " + err.Error())
		}
	}

	code := m.Run()

	// We'll let the Makefile handle cleanup
	// _ = os.Remove(binaryPath)

	os.Exit(code)
}

// TestRealPstreeOutput runs the actual pstree command and verifies the output format
func TestRealPstreeOutput(t *testing.T) {
	// Skip this test if we're running in a CI environment without process access
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test cases with different flag combinations
	testCases := []struct {
		name     string
		args     []string
		patterns []string // Regex patterns that should match in the output
	}{
		{
			name: "basic_tree",
			args: []string{},
			patterns: []string{
				// The init process should be in the output
				`init`,
			},
		},
		{
			name: "show_pids",
			args: []string{"--show-pids"},
			patterns: []string{
				// Process IDs should be shown in parentheses
				`\(\d+\)`,
			},
		},
		{
			name: "show_ppids",
			args: []string{"--show-ppids"},
			patterns: []string{
				// Parent Process IDs should be shown in parentheses
				`\(\d+\)`,
			},
		},
		{
			name: "show_owner",
			args: []string{"--show-owner"},
			patterns: []string{
				// Username should be shown before the command
				`root /`,
			},
		},
		{
			name: "max_depth_2",
			args: []string{"--level=2"},
			patterns: []string{
				// There should be some output but not too deep
				`.+`,
			},
		},
		{
			name: "order_by_pid",
			args: []string{"--order-by=pid"},
			patterns: []string{
				// There should be some output
				`.+`,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Run the pstree command with the specified arguments
			args := append([]string{}, tc.args...)
			cmd := exec.Command(binaryPath, args...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			err := cmd.Run()

			// If the command failed, print the error and stderr
			if err != nil {
				t.Logf("Command failed: %v", err)
				t.Logf("Stderr: %s", stderr.String())
			}

			// We don't assert on the error because some flag combinations might be invalid
			// and we want to check the error message in that case

			output := stdout.String()
			t.Logf("Command output (first 500 chars): %s", output[:min(500, len(output))])

			// Verify that the output matches the expected patterns
			for _, pattern := range tc.patterns {
				matched, err := regexp.MatchString(pattern, output)
				require.NoError(t, err, "Regex pattern should compile")
				assert.True(t, matched, "Output should match pattern %q", pattern)
			}
		})
	}
}

// Helper function to get the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestShowPPIDsRealOutput specifically tests the --show-ppids flag with real data
func TestShowPPIDsRealOutput(t *testing.T) {
	// Skip this test if we're running in a CI environment without process access
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Run the pstree command with --show-ppids flag
	cmd := exec.Command(binaryPath, "--show-ppids")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	require.NoError(t, err, "Command should execute successfully")

	output := stdout.String()
	t.Logf("Command output (first 500 chars): %s", output[:min(500, len(output))])

	// Verify that the output contains parent process IDs
	ppidPattern := `\(\d+\)`
	matched, err := regexp.MatchString(ppidPattern, output)
	require.NoError(t, err, "Regex pattern should compile")
	assert.True(t, matched, "Output should show parent process IDs in parentheses")
}

// TestOrderByRealOutput specifically tests the --order-by flag with real data
func TestOrderByRealOutput(t *testing.T) {
	// Skip this test if we're running in a CI environment without process access
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test cases for different ordering options
	orderOptions := []string{"pid", "cpu", "mem", "age", "threads", "user"}

	for _, option := range orderOptions {
		t.Run("order_by_"+option, func(t *testing.T) {
			// Run the pstree command with --order-by flag
			cmd := exec.Command(binaryPath, "--order-by="+option)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			err := cmd.Run()

			// If the command failed, print the error and stderr
			if err != nil {
				t.Logf("Command failed: %v", err)
				t.Logf("Stderr: %s", stderr.String())
				t.Fail()
			}

			output := stdout.String()
			t.Logf("Command output with --order-by=%s (first 500 chars): %s", option, output[:min(500, len(output))])

			// Verify that the output contains some process tree structure
			treePattern := `\|-|\\-`
			matched, err := regexp.MatchString(treePattern, output)
			require.NoError(t, err, "Regex pattern should compile")
			assert.True(t, matched, "Output should show a process tree structure")
		})
	}
}

// TestFlagCombinationsRealOutput tests various combinations of flags with real data
func TestFlagCombinationsRealOutput(t *testing.T) {
	// Skip this test if we're running in a CI environment without process access
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test cases with different flag combinations
	testCases := []struct {
		name     string
		args     []string
		patterns []string // Regex patterns that should match in the output
	}{
		{
			name: "show_pids_and_owner",
			args: []string{"--show-pids", "--show-owner"},
			patterns: []string{
				// Process IDs and owners should be shown
				`\(\d+\)`,
				`\([a-zA-Z0-9_]+\)`,
			},
		},
		{
			name: "show_ppids_and_pids",
			args: []string{"--show-ppids", "--show-pids"},
			patterns: []string{
				// Process IDs and parent process IDs should be shown in format (ppid,pid)
				`\(\d+,\d+\)`,
			},
		},
		{
			name: "utf8_graphics",
			args: []string{"--utf-8"},
			patterns: []string{
				// UTF-8 tree characters should be used
				`[├└]─`,
			},
		},
		{
			name: "compact_not",
			args: []string{"--compact-not"},
			patterns: []string{
				// There should be some output
				`.+`,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Run the pstree command with the specified arguments
			cmd := exec.Command(binaryPath, tc.args...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			err := cmd.Run()

			// If the command failed, print the error and stderr
			if err != nil {
				t.Logf("Command failed: %v", err)
				t.Logf("Stderr: %s", stderr.String())
				t.Fail()
			}

			output := stdout.String()
			t.Logf("Command output (first 500 chars): %s", output[:min(500, len(output))])

			// Verify that the output matches the expected patterns
			for _, pattern := range tc.patterns {
				matched, err := regexp.MatchString(pattern, output)
				require.NoError(t, err, "Regex pattern should compile")
				assert.True(t, matched, "Output should match pattern %q", pattern)
			}
		})
	}
}
