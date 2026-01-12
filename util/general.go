package util

import (
	"bytes"
	"fmt"
	"runtime"
	"slices"

	"math"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/mem"
	terminal "github.com/wayneashleyberry/terminal-dimensions"
)

type Duration struct {
	Days    int64
	Hours   int64
	Minutes int64
	Seconds int64
}

// ExecutePipeline executes a command pipeline and returns the results.
//
// This function takes a command string that may contain pipe operators (|) and executes
// each command in sequence, piping the output of each command to the input of the next.
// It handles setting up the necessary pipes between commands and captures stdout and stderr.
//
// Parameters:
//   - commandStr: The command string to execute, which may contain pipe operators
//
// Returns:
//   - int: Exit code of the last command in the pipeline
//   - string: Combined stdout output from the last command
//   - string: Combined stderr output from all commands
//   - error: Any error encountered during execution
func ExecutePipeline(commandStr string) (int, string, string, error) {
	commands := strings.Split(commandStr, "|")
	var cmds []*exec.Cmd

	// Trim spaces and create command slices
	for _, cmdStr := range commands {
		parts := strings.Fields(strings.TrimSpace(cmdStr))
		if len(parts) == 0 {
			continue
		}
		cmds = append(cmds, exec.Command(parts[0], parts[1:]...))
	}

	if len(cmds) == 0 {
		return -1, "", "No commands provided", fmt.Errorf("empty command pipeline")
	}

	// Set up pipes for the pipeline
	var stdoutBuf, stderrBuf bytes.Buffer
	var previousCmd *exec.Cmd

	for _, cmd := range cmds {
		cmd.Stderr = &stderrBuf // Capture stderr for each command

		if previousCmd != nil {
			// Create pipe between previous and current command
			pipeIn, err := previousCmd.StdoutPipe()
			if err != nil {
				return -1, "", "", fmt.Errorf("failed to create stdout pipe: %v", err)
			}
			cmd.Stdin = pipeIn
		}

		previousCmd = cmd // Move to the next command
	}

	// Capture output of the last command
	cmds[len(cmds)-1].Stdout = &stdoutBuf

	// Start and wait for all commands
	for _, cmd := range cmds {
		if err := cmd.Start(); err != nil {
			return -1, "", stderrBuf.String(), err
		}
	}

	// Ensure all commands complete execution
	for _, cmd := range cmds {
		if err := cmd.Wait(); err != nil {
			return -1, "", stderrBuf.String(), err
		}
	}

	// Get the exit code of the last command
	exitCode := 0
	if exitErr, ok := cmds[len(cmds)-1].ProcessState.Sys().(interface{ ExitCode() int }); ok {
		exitCode = exitErr.ExitCode()
	}

	return exitCode, strings.TrimRight(stdoutBuf.String(), "\n"), strings.TrimRight(stderrBuf.String(), "\n"), nil
}

// GetTotalMemory retrieves information about the system's virtual memory.
//
// This function uses the gopsutil library to get detailed statistics about
// the system's memory usage, including total memory, available memory, and usage percentages.
//
// Returns:
//   - *mem.VirtualMemoryStat: Structure containing memory statistics
//   - error: Any error encountered while retrieving memory information
func GetTotalMemory() (*mem.VirtualMemoryStat, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return &mem.VirtualMemoryStat{}, err
	}
	return v, nil
}

// StrToInt32 converts a string to an int32 value.
//
// This function parses a string representation of an integer and returns it as an int32.
// If parsing fails, it silently returns 0.
//
// Parameters:
//   - input: String to convert to int32
//
// Returns:
//   - int32: The converted value, or 0 if conversion fails
func StrToInt32(input string) int32 {
	num, _ := strconv.ParseInt(input, 10, 32)
	return int32(num)
}

// Int32toStr converts an int32 value to a string.
//
// Parameters:
//   - input: int32 value to convert
//
// Returns:
//   - string: String representation of the input value
func Int32toStr(input int32) string {
	output := strconv.Itoa(int(input))
	return output
}

// SortSlice sorts a slice of int32 values in ascending order.
//
// Parameters:
//   - unsorted: Slice of int32 values to sort
//
// Returns:
//   - []int32: Sorted slice in ascending order
func SortSlice(unsorted []int32) []int32 {
	slices.Sort(unsorted)
	return unsorted
}

// Contains checks if a string slice contains a specific value.
//
// Parameters:
//   - elems: Slice of strings to search in
//   - v: String value to search for
//
// Returns:
//   - bool: true if the value is found, false otherwise
func Contains(elems []string, v string) bool {
	return slices.Contains(elems, v)
}

// GetScreenWidth determines the width of the terminal in characters.
//
// This function attempts to get the width of the terminal using the terminal-dimensions
// package. If it fails, it returns a default width of 132 characters.
//
// Returns:
//   - int: Width of the terminal in characters
func GetScreenWidth() int {
	var (
		err   error
		width uint
	)
	width, err = terminal.Width()
	if err != nil {
		return 132
	}

	return int(width)
}

// TruncateString truncates a string to the specified maximum length.
//
// If the string is longer than the specified length, it returns a substring
// containing the first 'length' characters. Otherwise, it returns the original string.
//
// Parameters:
//   - s: String to truncate
//   - length: Maximum length of the returned string
//
// Returns:
//   - string: Truncated string
func TruncateString(s string, length int) string {
	if len(s) > length {
		return s[:length]
	}
	return s
}

// HasColorSupport determines if the terminal supports color output and how many colors.
//
// This function uses the 'tput colors' command to determine the number of colors
// supported by the terminal. It considers color support to be available if at least
// 8 colors are supported.
//
// Returns:
//   - bool: true if the terminal supports at least 8 colors, false otherwise
//   - int: Number of colors supported by the terminal, or 0 if color is not supported
func HasColorSupport() (bool, int) {
	switch runtime.GOOS {
	case "windows":
		return true, 256
	case "darwin", "linux":
		returncode, stdout, _, err := ExecutePipeline("/usr/bin/tput colors")
		if err != nil || returncode != 0 {
			return false, 0
		}
		colors, err := strconv.Atoi(stdout)
		if err != nil {
			return false, 0
		}
		if colors < 8 {
			return false, 0
		}
		return true, colors
	default:
		return false, 0
	}
}

// UserExists checks if a user with the specified username exists on the system.
//
// Parameters:
//   - username: Username to check for existence
//
// Returns:
//   - bool: true if the user exists, false otherwise
func UserExists(username string) bool {
	_, err := user.Lookup(username)
	return err == nil
}

// RoundFloat rounds a floating-point number to the specified precision.
//
// Parameters:
//   - val: Floating-point value to round
//   - precision: Number of decimal places to round to
//
// Returns:
//   - float64: Rounded value
func RoundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

// ByteConverter formats a byte count as a human-readable string with appropriate units.
//
// This function converts a raw byte count into a human-readable string with binary prefixes
// (Ki, Mi, Gi, etc.) according to the IEC standard. The result is formatted with two decimal
// places of precision.
//
// Parameters:
//   - num: Number of bytes to format
//
// Returns:
//   - string: Formatted string with appropriate binary unit prefix
func ByteConverter(num uint64) string {
	var (
		absolute float64
		suffix   string = "B"
		unit     string
	)
	absolute = math.Abs(float64(num))

	for _, unit = range []string{"", "Ki", "Mi", "Gi", "Ti", "Pi", "Ei", "Zi"} {
		if absolute < 1024.0 {
			return fmt.Sprintf("%.2f %s%s", RoundFloat(absolute, 2), unit, suffix)
		}
		absolute = absolute / 1024
	}
	return fmt.Sprintf("%.2f Yi%s", RoundFloat(absolute, 2), suffix)
}

// BtoI converts a boolean value to an integer (1 for true, 0 for false).
//
// Parameters:
//   - b: Boolean value to convert
//
// Returns:
//   - int: 1 if the input is true, 0 if false
func BtoI(b bool) int {
	if b {
		return 1
	}
	return 0
}

// StoI converts a string to an integer based on whether it's empty or not.
//
// This function returns 1 if the string is not empty, and 0 if it is empty.
// It's primarily used for counting non-empty strings in flag validation.
//
// Parameters:
//   - s: String to check
//
// Returns:
//   - int: 1 if the string is not empty, 0 if empty
func StoI(s string) int {
	if s != "" {
		return 1
	}
	return 0
}

// GetUnixTimestamp returns the current Unix timestamp in seconds.
//
// This function provides the number of seconds elapsed since January 1, 1970 UTC.
//
// Returns:
//   - int64: Current Unix timestamp in seconds
func GetUnixTimestamp() int64 {
	return time.Now().Unix()
}

// DetermineUsername gets the username of the current user.
//
// This function attempts to retrieve the username of the current user using the os/user
// package. If it fails, it returns an empty string.
//
// Returns:
//   - string: Username of the current user, or empty string if it cannot be determined
func DetermineUsername() string {
	username, err := user.Current()
	if err != nil {
		return ""
	}
	return username.Username
}

// FindDuration converts a duration in seconds to a structured Duration type.
//
// This function breaks down a total number of seconds into days, hours, minutes,
// and remaining seconds for more readable time representation.
//
// Parameters:
//   - seconds: Total duration in seconds
//
// Returns:
//   - Duration: Structured representation with days, hours, minutes, and seconds
func FindDuration(seconds int64) (duration Duration) {
	days := int64(seconds / 86400)
	hours := int64(((seconds - (days * 86400)) / 3600))
	minutes := int64(((seconds - days*86400 - hours*3600) / 60))
	secs := int64((seconds - (days * 86400) - (hours * 3600) - (minutes * 60)))
	return Duration{
		Days:    days,
		Hours:   hours,
		Minutes: minutes,
		Seconds: secs,
	}
}

// DeleteSliceElement removes an element from a slice of strings at the specified index.
//
// Parameters:
//   - slice: Slice to modify
//   - index: Index of the element to remove
//
// Returns:
//   - []string: New slice with the element removed, or the original slice if index is out of range
func DeleteSliceElement(slice []string, index int) []string {
	if len(slice) == 0 || index < 0 || index >= len(slice) {
		return slice
	}
	return append(slice[:index], slice[index+1:]...)
}
