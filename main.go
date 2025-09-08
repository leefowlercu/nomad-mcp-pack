package main

import (
	"log/slog"
	"os"

	"github.com/leefowlercu/nomad-mcp-pack/cmd"
	"github.com/leefowlercu/nomad-mcp-pack/internal/config"
	"github.com/leefowlercu/nomad-mcp-pack/internal/logutils"
)

func main() {
	config.InitConfig()

	cfg, err := config.GetConfig()
	if err != nil {
		os.Stderr.WriteString("Failed to load configuration: " + err.Error() + "\n")
		os.Exit(1)
	}

	logger := logutils.SetupLogger(cfg)
	slog.SetDefault(logger)

	slog.Info("starting nomad-mcp-pack",
		"mcp_registry_url", cfg.MCPRegistryURL,
		"log_level", cfg.LogLevel,
		"env", cfg.Env,
	)

	if err := cmd.Execute(); err != nil {
		slog.Error("fatal error", "err", err)
		os.Exit(1)
	}
}
