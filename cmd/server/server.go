package cmdserver

import (
	"fmt"
	"log/slog"

	"github.com/leefowlercu/nomad-mcp-pack/internal/config"
	"github.com/leefowlercu/nomad-mcp-pack/internal/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Start a Nomad MCP Pack server",
	Long: "Start a Nomad MCP Pack server that provides remote access to pack generation.\n\n" +
		"The server exposes a REST API endpoint for generating Nomad MCP Server Packs.",
	Example: `  # Start server with default settings
  nomad-mcp-pack server
  
  # Start on custom port
  nomad-mcp-pack server --addr ":9090"
	
  # Start with custom timeouts
  nomad-mcp-pack server --read-timeout 30 --write-timeout 30`,
	RunE: runServer,
}

func init() {
	ServerCmd.Flags().String("addr", config.DefaultConfig.ServerAddr, "Server address")
	ServerCmd.Flags().Int("read-timeout", config.DefaultConfig.ServerReadTimeout, "Read timeout in seconds")
	ServerCmd.Flags().Int("write-timeout", config.DefaultConfig.ServerWriteTimeout, "Write timeout in seconds")

	viper.BindPFlag("server.addr", ServerCmd.Flags().Lookup("addr"))
	viper.BindPFlag("server.read_timeout", ServerCmd.Flags().Lookup("read-timeout"))
	viper.BindPFlag("server.write_timeout", ServerCmd.Flags().Lookup("write-timeout"))
}

func runServer(cmd *cobra.Command, args []string) error {
	slog.Info("starting server command run")

	cfg, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	output.Info("Starting server on %s...", cfg.Server.Addr)

	// TODO: Implement actual server functionality
	output.Warning("Server command is not yet implemented")

	slog.Info("server command run completed successfully")

	return nil
}
