package cmd

import (
	"github.com/leefowlercu/nomad-mcp-pack/cmd/generate"
	"github.com/leefowlercu/nomad-mcp-pack/cmd/inspect"
	"github.com/leefowlercu/nomad-mcp-pack/cmd/list"
	"github.com/leefowlercu/nomad-mcp-pack/cmd/server"
	"github.com/leefowlercu/nomad-mcp-pack/cmd/watch"
	"github.com/spf13/cobra"
)

var (
	version         string
	mcpRegistryAddr string
)

var rootCmd = &cobra.Command{
	Use:   "nomad-mcp-pack",
	Short: "A generator for HashiCorp Nomad MCP Server Packs",
	Long: "\nNomad MCP Pack generates HashiCorp Nomad Packs for MCP Servers described " +
		"by the configured MCP Registry (default: registry.modelcontextprotocol.io)",
	Version: version,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&mcpRegistryAddr, "mcp-registry-addr", "registry.modelcontextprotocol.io", "MCP Server Registry address")

	rootCmd.AddCommand(generate.GenerateCmd)
	rootCmd.AddCommand(inspect.InspectCmd)
	rootCmd.AddCommand(list.ListCmd)
	rootCmd.AddCommand(server.ServerCmd)
	rootCmd.AddCommand(watch.WatchCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
