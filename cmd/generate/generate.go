package cmdgenerate

import (
	"fmt"
	"log/slog"

	"github.com/leefowlercu/nomad-mcp-pack/internal/genutils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	dryRun      bool
	force       bool
	packageType string
)

var GenerateCmd = &cobra.Command{
	Use:   "generate <mcp-server@version>",
	Short: "Generate a Nomad Pack for an MCP Server",
	Long: "\nGenerate a Nomad Pack for an MCP Server at a specific version.\n\n" +
		"This command fetches the MCP Server definition matching the supplied Server name and version from the configured " +
		"MCP Registry (default: https://registry.modelcontextprotocol.io) and creates a complete Nomad Pack with Job " +
		"Templates, Variables, and Metadata.\n\n" +
		"The syntax for the MCP Server argument is <mcp-server@version> where `mcp-server` is the name of the MCP Server and  `version` " +
		"can either be a semver-formatted string or the keyword 'latest'. Using the keyword 'latest' will attempt to find the latest non-deprecated, " +
		"non-deleted version. If a non-deprecated, non-deleted version can be found for the given MCP Server Version " +
		"the Nomad Pack will be generated. If there is no matching MCP Server Version found, or if there is a match but the matching " +
		"MCP Server Version has been declared deprecated or deleted the command will error and indicate the reason.",
	Example: `  # Generate a pack for a specific version of an MCP server
  nomad-mcp-pack generate io.github.datastax/astra-db-mcp@0.0.1-seed
  
  # Generate a pack for the latest version
  nomad-mcp-pack generate io.github.datastax/astra-db-mcp@latest
  
  # Generate to a specific directory
  nomad-mcp-pack generate io.github.datastax/astra-db-mcp@0.0.1-seed -o ./packs/astra-db
  
  # Dry run to see what would be generated
  nomad-mcp-pack generate io.github.datastax/astra-db-mcp@latest --dry-run
  
  # Force overwrite existing pack
  nomad-mcp-pack generate io.github.datastax/astra-db-mcp@latest --force
  
  # Specify preferred package type when multiple are available
  nomad-mcp-pack generate io.github.datastax/astra-db-mcp@latest --package-type oci`,
	Args: cobra.ExactArgs(1),
	RunE: runGenerate,
}

func init() {
	GenerateCmd.Flags().StringP("output-dir", "o", "", "Output directory for the generated pack")
	GenerateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be generated without writing files")
	GenerateCmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing pack if it exists")
	GenerateCmd.Flags().StringVar(&packageType, "package-type", "oci", "Preferred package type {npm|pypi|oci|nuget}")

	viper.BindPFlag("output_dir", GenerateCmd.Flags().Lookup("output-dir"))
}

func runGenerate(cmd *cobra.Command, args []string) error {
	serverSpec, err := genutils.ParseServerSpec(args[0])
	if err != nil {
		return fmt.Errorf("invalid server specification: %w", err)
	}

	outputDir := viper.GetString("output_dir")
	registryURL := viper.GetString("mcp_registry_url")

	slog.Info("generating nomad pack",
		"server", serverSpec.ServerName,
		"version", serverSpec.Version,
		"output_dir", outputDir,
		"registry_url", registryURL,
		"dry_run", dryRun,
		"force", force,
		"package_type", packageType,
	)

	if serverSpec.IsLatest() {
		slog.Debug("resolving latest version", "server", serverSpec.ServerName)
		// TODO: Query registry for latest non-deprecated, non-deleted version
	}

	// TODO: Call pkg/generate package implementation
	// return generate.Run(cmd.Context(), serverSpec, generate.Options{
	//     OutputDir:   outputDir,
	//     DryRun:      dryRun,
	//     Force:       force,
	//     PackageType: packageType,
	//     RegistryURL: registryURL,
	// })

	// TODO: Handle Error situations
	// Error: MCP Server Version with specified name not found
	// Error: MCP Server Version specified was found but marked deprecated
	// Error: MCP Server Version specified was found but marked deleted

	fmt.Printf("Would generate pack for %s\n", serverSpec)
	return nil
}
