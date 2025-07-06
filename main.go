package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/kuderr/deepwiki/cmd"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stdout, "Panic: %v\n", r)
			fmt.Fprintf(os.Stdout, "%s\n", debug.Stack())
			os.Exit(1)
		}
	}()

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stdout, "Error: %v\n", err)
		os.Exit(1)
	}
}
