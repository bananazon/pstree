package main

import (
	"os"

	"github.com/bananazon/pstree/cmd"
)

// main is the entry point for the pstree application.
// It executes the root command and handles any errors that occur.
// If an error is encountered, the program will exit with a non-zero status code.
func main() {
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
