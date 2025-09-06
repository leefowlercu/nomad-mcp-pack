package list

import (
	"github.com/spf13/cobra"
)

var (
	outputFormat      string
	filterCategory    string
	filterPackageType string
	limit             int
	offset            int
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available MCP Servers from the configured MCP Registry",
	Long: "List all available MCP Servers from the configured MCP Registry.\n\n" +
		"This command queries the configured MCP Registry (default: registry.modelcontextprotocol.io) and displays " +
		"available MCP Servers with filtering and pagination options.",
	Example: `  # List all servers
  nomad-mcp-pack list
  
  # List servers with JSON output
  nomad-mcp-pack list --format json
  
  # Filter by category
  nomad-mcp-pack list --category "AI Tools"
  
  # Filter by package type
  nomad-mcp-pack list --package-type docker
  
  # Paginate results
  nomad-mcp-pack list --limit 10 --offset 20`,
	RunE: runList,
}

func init() {
	ListCmd.Flags().StringVarP(&outputFormat, "format", "f", "table", "Output format (table, json, yaml)")
	ListCmd.Flags().StringVar(&filterCategory, "category", "", "Filter servers by category")
	ListCmd.Flags().StringVar(&filterPackageType, "package-type", "", "Filter by package type (docker, npm, pypi, github)")
	ListCmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of servers to display")
	ListCmd.Flags().IntVar(&offset, "offset", 0, "Number of servers to skip")
}

func runList(cmd *cobra.Command, args []string) error {
	// TODO: Call internal/list package implementation
	// return list.Run(cmd.Context(), list.Options{
	//     Format:      outputFormat,
	//     Category:    filterCategory,
	//     PackageType: filterPackageType,
	//     Limit:       limit,
	//     Offset:      offset,
	// })

	return nil
}
