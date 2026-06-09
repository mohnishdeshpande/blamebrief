package main

import (
	"blamebrief/cmd"
	"fmt"
	"os"
)

func main() {
	// Active entrypoint for BlameBrief
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
