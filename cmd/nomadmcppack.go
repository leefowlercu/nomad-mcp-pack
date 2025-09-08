package cmd

import (
	cmdgenerate "github.com/leefowlercu/nomad-mcp-pack/cmd/generate"
	cmdserver "github.com/leefowlercu/nomad-mcp-pack/cmd/server"
	cmdwatch "github.com/leefowlercu/nomad-mcp-pack/cmd/watch"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version string
)

var nomadMcpPackCmd = &cobra.Command{
	Use:   "nomad-mcp-pack",
	Short: "A generator for HashiCorp Nomad MCP Server Packs",
	Long: "\nNomad MCP Pack generates HashiCorp Nomad Packs for MCP Servers described " +
		"by the configured MCP Registry (default: https://registry.modelcontextprotocol.io)",
	Version: version,
}

func init() {
	nomadMcpPackCmd.PersistentFlags().String("mcp-registry-url", "", "MCP Registry URL")

	viper.BindPFlag("mcp_registry_url", nomadMcpPackCmd.PersistentFlags().Lookup("mcp-registry-url"))

	nomadMcpPackCmd.AddCommand(cmdgenerate.GenerateCmd)
	nomadMcpPackCmd.AddCommand(cmdserver.ServerCmd)
	nomadMcpPackCmd.AddCommand(cmdwatch.WatchCmd)
}

func Execute() error {
	return nomadMcpPackCmd.Execute()
}
