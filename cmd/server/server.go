package cmdserver

import (
	"log/slog"

	"github.com/leefowlercu/nomad-mcp-pack/internal/config"
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
}

func runServer(cmd *cobra.Command, args []string) error {
	cfg, err := config.GetConfig()
	if err != nil {
		return err
	}

	slog.Info("starting server",
		"addr", cfg.Server.Addr,
		"read_timeout", cfg.Server.ReadTimeout,
		"write_timeout", cfg.Server.WriteTimeout,
		"output_dir", cfg.OutputDir,
		"mcp_registry_url", cfg.MCPRegistryURL,
	)

	slog.Debug("configuration sources",
		"config_file", viper.ConfigFileUsed(),
		"env", cfg.Env,
	)

	return nil
}
