package main

import (
	"os"

	"github.com/leefowlercu/nomad-mcp-pack/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
