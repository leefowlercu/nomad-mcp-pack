package cmdgenerate

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/leefowlercu/nomad-mcp-pack/internal/config"
	"github.com/leefowlercu/nomad-mcp-pack/internal/genutils"
	"github.com/leefowlercu/nomad-mcp-pack/pkg/generate"
	"github.com/leefowlercu/nomad-mcp-pack/pkg/registry"
	v0 "github.com/modelcontextprotocol/registry/pkg/api/v0"
	"github.com/modelcontextprotocol/registry/pkg/model"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	dryRun          bool
	force           bool
	packageType     string
	allowDeprecated bool
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
  nomad-mcp-pack generate io.github.datastax/astra-db-mcp@0.0.1-seed --output-dir ./astra-packs
  
  # Dry run to see what would be generated
  nomad-mcp-pack generate io.github.datastax/astra-db-mcp@latest --dry-run
  
  # Force overwrite existing pack
  nomad-mcp-pack generate io.github.datastax/astra-db-mcp@latest --force
  
  # Specify package type (default 'oci' (Docker))
  nomad-mcp-pack generate io.github.datastax/astra-db-mcp@latest --package-type npm`,
	Args: cobra.ExactArgs(1),
	RunE: runGenerate,
}

func init() {
	GenerateCmd.Flags().StringVar(&packageType, "package-type", "oci", "Preferred package type {npm|pypi|oci|nuget}")
	GenerateCmd.Flags().String("output-dir", config.DefaultConfig.OutputDir, "Output directory for the generated pack (itself, a directory)")
	GenerateCmd.Flags().String("output-type", config.DefaultConfig.OutputType, "Output type {packdir|archive}")
	GenerateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be generated without writing files")
	GenerateCmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing pack if it exists")
	GenerateCmd.Flags().BoolVar(&allowDeprecated, "allow-deprecated", false, "Allow generation of packs for deprecated servers")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	serverSpecArg := args[0]
	slog.Debug("parsing generate command argument", "argument", serverSpecArg)

	serverSpec, err := genutils.ParseServerSpec(serverSpecArg)
	if err != nil {
		return fmt.Errorf("invalid server specification: %w", err)
	}

	outputDir := viper.GetString("output_dir")
	outputType := viper.GetString("output_type")
	registryURL := viper.GetString("mcp_registry_url")

	slog.Info("generating nomad pack",
		"server", serverSpec.ServerName,
		"version", serverSpec.Version,
		"output_dir", outputDir,
		"output_type", outputType,
		"registry_url", registryURL,
		"dry_run", dryRun,
		"force", force,
		"package_type", packageType,
		"allow_deprecated", allowDeprecated,
	)

	client, err := registry.NewClient(registryURL)
	if err != nil {
		return fmt.Errorf("failed to create registry client: %w", err)
	}

	ctx := cmd.Context()
	var server *v0.ServerJSON

	if serverSpec.IsLatest() {
		slog.Debug("resolving latest version", "server", serverSpec.ServerName)

		serverResp, err := client.GetServerByNameAndVersion(ctx, serverSpec.ServerName, "latest")
		if err != nil {
			return fmt.Errorf("failed to get latest server version: %w", err)
		}
		server = serverResp

		serverSpec.Version = server.Version
		slog.Info("resolved latest version", "server", serverSpec.ServerName, "version", server.Version)
	} else {
		slog.Debug("fetching specific version", "server", serverSpec.ServerName, "version", serverSpec.Version)

		serverResp, err := client.GetServerByNameAndVersion(ctx, serverSpec.ServerName, serverSpec.Version)
		if err != nil {
			return fmt.Errorf("failed to get server version %s: %w", serverSpec.Version, err)
		}
		server = serverResp
	}

	if server.Status == model.StatusDeleted {
		return fmt.Errorf("server %s@%s is deleted and cannot be used", serverSpec.ServerName, server.Version)
	}
	if server.Status == model.StatusDeprecated && !allowDeprecated {
		return fmt.Errorf("server %s@%s is deprecated (use --allow-deprecated to generate anyway)", serverSpec.ServerName, server.Version)
	}
	if server.Status == model.StatusDeprecated && allowDeprecated {
		slog.Warn("generating pack for deprecated server", "server", serverSpec.ServerName, "version", server.Version)
	}

	hasMatchingPackage := false
	for _, pkg := range server.Packages {
		if pkg.RegistryType == packageType {
			hasMatchingPackage = true
			break
		}
	}

	if !hasMatchingPackage {
		availableTypes := make([]string, 0, len(server.Packages))
		for _, pkg := range server.Packages {
			availableTypes = append(availableTypes, pkg.RegistryType)
		}

		if len(availableTypes) == 0 {
			return fmt.Errorf("server %s@%s has no packages defined", serverSpec.ServerName, server.Version)
		}

		return fmt.Errorf("server %s@%s does not have a package of type %q (available: %s)",
			serverSpec.ServerName, server.Version, packageType, strings.Join(availableTypes, ", "))
	}

	// Generate the Nomad Pack
	opts := generate.Options{
		OutputDir:  outputDir,
		OutputType: outputType,
		DryRun:     dryRun,
		Force:      force,
	}

	err = generate.Run(cmd.Context(), server, serverSpec, packageType, opts)
	if err != nil {
		return fmt.Errorf("failed to generate pack: %w", err)
	}

	if dryRun {
		slog.Info("dry run completed successfully")
	} else {
		packName := fmt.Sprintf("%s-%s-%s",
			strings.ReplaceAll(serverSpec.ServerName, "/", "-"),
			serverSpec.Version,
			packageType)
		slog.Info("pack generated successfully",
			"pack_name", packName,
			"output_dir", outputDir)
	}

	return nil
}
