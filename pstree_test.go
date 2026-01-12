package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
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

// TestVersionFlag tests the version flag
func TestVersionFlag(t *testing.T) {
	testCases := []struct {
		name       string
		args       []string
		shouldFail bool
	}{
		{"Version", []string{"--version"}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tc.args...)
			output, err := cmd.CombinedOutput()

			if tc.shouldFail {
				assert.Error(t, err, string(output))
			} else {
				assert.NoError(t, err, string(output))
			}
		})
	}
}

func TestCommandLineArgs(t *testing.T) {
	testCases := []struct {
		name       string
		args       []string
		shouldFail bool
	}{
		{"Basic", []string{"pstree"}, false},
		{"Help", []string{"pstree", "--help"}, false},
		{"Version", []string{"pstree", "--version"}, false},
		{"ShowPIDs", []string{"pstree", "--show-pids"}, false},
		{"ShowPPIDs", []string{"pstree", "--show-ppids"}, false},
		{"ShowPGIDs", []string{"pstree", "--show-pgids"}, false},
		{"ShowAllPidStuff", []string{"pstree", "--show-pids", "--show-ppids", "--show-pgids"}, false},
		{"ShowOwner", []string{"pstree", "--show-owner"}, false},
		{"Age", []string{"pstree", "--age"}, false},
		{"UTF8", []string{"pstree", "--utf-8"}, false},
		{"ValidLevel", []string{"pstree", "--level", "2"}, false},
		{"InvalidLevel", []string{"pstree", "--level", "0"}, true},
		{"ConflictingGraphics", []string{"pstree", "--utf-8", "--ibm-850"}, true},
		{"ConflictingTransitions", []string{"pstree", "--uid-transitions", "--user-transitions"}, true},
		{"ConflictingColors", []string{"pstree", "--color", "cpu", "--colorize"}, true},
		{"ValidColorAttr", []string{"pstree", "--color", "age"}, false},
		{"InvalidColorAttr", []string{"pstree", "--color", "invalid"}, true},
		{"ValidOrderBy", []string{"pstree", "--order-by", "cpu"}, false},
		{"InvalidOrderBy", []string{"pstree", "--order-by", "invalid"}, true},
		{"SetUTF8andVT100", []string{"pstree", "--utf-8", "--vt-100"}, true},
		{"ShowLotsOfStuff", []string{"pstree", "--show-owner", "--show-pids", "--show-ppids", "--show-pgids",
			"--age", "--cpu", "--memory", "--threads", "--user-transitions"}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tc.args...)
			output, err := cmd.CombinedOutput()

			if tc.shouldFail {
				assert.Error(t, err, string(output))
			} else {
				assert.NoError(t, err, string(output))
			}
		})
	}
}
