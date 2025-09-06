package generate

import (
	"github.com/spf13/cobra"
)

var (
	outputDir   string
	dryRun      bool
	force       bool
	packageType string
)

var GenerateCmd = &cobra.Command{
	Use:   "generate <mcp-server@version>",
	Short: "Generate a Nomad Pack for an MCP Server",
	Long: "\nGenerate a Nomad Pack for an MCP Server at a specific version.\n\n" +
		"The command fetches the MCP Server definition matching the supplied Server name and version from the configured " +
		"MCP Registry (default: registry.modelcontextprotocol.io) and creates a complete Nomad Pack with Job " +
		"Templates, Variables, and Metadata.\n\n" +
		"The syntax for the MCP Server argument is <mcp-server@version> where `mcp-server` is the name of the MCP Server and  `version` " +
		"can either be a semver-formatted string or the keyword 'latest'. If a non-deprecated, non-deleted version can be found for the given MCP Server " +
		"name the Nomad Pack will be generated. If there is no matching MCP Server Version found, or if there is a match but the matching " +
		"MCP Server Version has been declared deprecated or deleted the command will error and indicate the reason.",
	Example: `  # Generate a pack for the GitHub MCP server
  nomad-mcp-pack generate io.github.modelcontextprotocol/filesystem@1.0.2
  
  # Generate to a specific directory
  nomad-mcp-pack generate mcp-github -o ./packs/github
  
  # Dry run to see what would be generated
  nomad-mcp-pack generate mcp-github --dry-run
  
  # Force overwrite existing pack
  nomad-mcp-pack generate mcp-github --force`,
	Args: cobra.ExactArgs(1),
	RunE: runGenerate,
}

func init() {
	GenerateCmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory for the generated pack (defaults to ./<server-id>)")
	GenerateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be generated without writing files")
	GenerateCmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing pack if it exists")
	GenerateCmd.Flags().StringVar(&packageType, "package-type", "", "Preferred package type (docker, npm, pypi, github)")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	// TODO: Parse argument stanzas

	// TODO: Call internal/generate package implementation
	// return generate.Run(cmd.Context(), serverID, generate.Options{
	//     OutputDir:   outputDir,
	//     DryRun:      dryRun,
	//     Force:       force,
	//     PackageType: packageType,
	// })

	// TODO: Handle Error situations
	// Error: MCP Server Version argument malformed
	// Error: MCP Server Version with specified name not found
	// Error: MCP Server Version specified was found but marked deprecated
	// Error: MCP Server Version specified was found but marked deleted

	return nil
}
