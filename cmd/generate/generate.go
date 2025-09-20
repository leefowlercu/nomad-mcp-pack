package cmdgenerate

import (
	"fmt"
	"log/slog"
	"net/url"

	"github.com/leefowlercu/go-mcp-registry/mcp"
	"github.com/leefowlercu/nomad-mcp-pack/internal/config"
	"github.com/leefowlercu/nomad-mcp-pack/internal/generator"
	"github.com/leefowlercu/nomad-mcp-pack/internal/output"
	"github.com/leefowlercu/nomad-mcp-pack/internal/server"
	"github.com/leefowlercu/nomad-mcp-pack/internal/validate"
	v0 "github.com/modelcontextprotocol/registry/pkg/api/v0"
	"github.com/modelcontextprotocol/registry/pkg/model"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
  nomad-mcp-pack generate io.github.datastax/astra-db-mcp@latest --force-overwrite
  
  # Specify package type (default 'oci' (Docker))
  nomad-mcp-pack generate io.github.datastax/astra-db-mcp@latest --package-type npm
  
  # Specify transport type (default 'http')
  nomad-mcp-pack generate io.github.datastax/astra-db-mcp@latest --transport-type sse`,
	Args:    cobra.ExactArgs(1),
	PreRunE: runValidate,
	RunE:    runGenerate,
}

func init() {
	GenerateCmd.Flags().String("package-type", config.DefaultConfig.GeneratePackageType, "Package type {npm|pypi|oci|nuget}")
	GenerateCmd.Flags().String("transport-type", config.DefaultConfig.GenerateTransportType, "Transport type {stdio|http|sse}")

	viper.BindPFlag("generate.package_type", GenerateCmd.Flags().Lookup("package-type"))
	viper.BindPFlag("generate.transport_type", GenerateCmd.Flags().Lookup("transport-type"))

	GenerateCmd.Flags().SortFlags = false
}

func runValidate(cmd *cobra.Command, args []string) error {
	slog.Info("starting generate command input validation")

	serverSearchSpecArg := args[0]

	cfg, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration; %w", err)
	}

	slog.Debug("validating generate command inputs with configuration",
		slog.Group("common_config",
			"registry_url", cfg.RegistryURL,
			"log_level", cfg.LogLevel,
			"env", cfg.Env,
			"output_dir", cfg.OutputDir,
			"output_type", cfg.OutputType,
			"allow_deprecated", cfg.AllowDeprecated,
			"dry_run", cfg.DryRun,
			"force_overwrite", cfg.ForceOverwrite,
		),
		slog.Group("generate_config",
			"package_type", cfg.Generate.PackageType,
			"transport_type", cfg.Generate.TransportType,
		),
	)

	packageType := cfg.Generate.PackageType
	transportType := cfg.Generate.TransportType

	if _, err = server.ParseSearchSpec(serverSearchSpecArg); err != nil {
		return fmt.Errorf("could not parse server argument; %w", err)
	}

	if err := validate.PackageType(packageType); err != nil {
		return fmt.Errorf("could not validate package type; %w", err)
	}

	if err := validate.TransportType(transportType); err != nil {
		return fmt.Errorf("could not validate transport type; %w", err)
	}

	slog.Info("generate command input validation completed successfully")

	// Any errors after this point are runtime errors, not usage-related errors
	cmd.SilenceUsage = true

	return nil
}

func runGenerate(cmd *cobra.Command, args []string) error {
	slog.Info("starting generate command run")

	ctx := cmd.Context()
	serverSearchSpecArg := args[0]

	cfg, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration; %w", err)
	}

	slog.Debug("running generate command with configuration",
		slog.Group("common_config",
			"registry_url", cfg.RegistryURL,
			"log_level", cfg.LogLevel,
			"env", cfg.Env,
			"output_dir", cfg.OutputDir,
			"output_type", cfg.OutputType,
			"allow_deprecated", cfg.AllowDeprecated,
			"dry_run", cfg.DryRun,
			"force_overwrite", cfg.ForceOverwrite,
		),
		slog.Group("generate_config",
			"package_type", cfg.Generate.PackageType,
			"transport_type", cfg.Generate.TransportType,
		),
	)

	packageType := cfg.Generate.PackageType
	transportType := cfg.Generate.TransportType

	registryURL := cfg.RegistryURL
	outputDir := cfg.OutputDir
	outputType := cfg.OutputType
	allowDeprecated := cfg.AllowDeprecated
	dryRun := cfg.DryRun
	forceOverwrite := cfg.ForceOverwrite

	client := mcp.NewClient(nil)
	registryURLParsed, err := url.Parse(registryURL)
	if err != nil {
		return fmt.Errorf("could not parse registry URL; %w", err)
	}
	client.BaseURL = registryURLParsed

	serverSearchSpec, err := server.ParseSearchSpec(serverSearchSpecArg)
	if err != nil {
		return fmt.Errorf("could not parse server argument; %w", err)
	}

	output.Info("Generating pack for %s...", serverSearchSpec)

	serverSpec, err := server.Find(ctx, serverSearchSpec, client)
	if err != nil {
		return fmt.Errorf("could not retrieve server %q from registry; %w", serverSearchSpec, err)
	}

	if serverSpec.IsDeleted() {
		return fmt.Errorf("server %q, version %s is deleted and cannot be used", serverSpec.Name(), serverSpec.Version())
	}
	if serverSpec.IsDeprecated() && !allowDeprecated {
		return fmt.Errorf("server %q, version %s is deprecated (use --allow-deprecated to generate anyway)", serverSpec.Name(), serverSpec.Version())
	}
	if serverSpec.IsDeprecated() && allowDeprecated {
		output.Warning("Server %s@%s is deprecated", serverSpec.Name(), serverSpec.Version())
		slog.Warn("generating pack for deprecated server", "server", serverSpec.Name(), "version", serverSpec.Version())
	}

	var srv *v0.ServerJSON
	var pkg *model.Package

	srv = serverSpec.JSON
	pkg, err = server.FindPackageWithTransport(srv, packageType, transportType)
	if err != nil {
		return fmt.Errorf("unable to generate pack for server %q, version %s: %w", serverSpec.Name(), serverSpec.Version(), err)
	}

	opts := generator.Options{
		OutputDir:      outputDir,
		OutputType:     string(outputType),
		DryRun:         dryRun,
		ForceOverwrite: forceOverwrite,
	}

	err = generator.Run(ctx, srv, pkg, opts)
	if err != nil {
		output.Failure("Pack generation failed: %v", err)
		return fmt.Errorf("failed to generate pack; %w", err)
	}

	slog.Info("generate command run completed successfully")

	return nil
}
