package main

import (
	"log/slog"
	"os"

	"github.com/leefowlercu/nomad-mcp-pack/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		slog.Error("fatal error", "err", err)
		os.Exit(1)
	}
}
