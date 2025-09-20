package cmd

import (
	"fmt"
	"log/slog"
	"os"

	cmdgenerate "github.com/leefowlercu/nomad-mcp-pack/cmd/generate"
	cmdserver "github.com/leefowlercu/nomad-mcp-pack/cmd/server"
	cmdwatch "github.com/leefowlercu/nomad-mcp-pack/cmd/watch"
	"github.com/leefowlercu/nomad-mcp-pack/internal/config"
	"github.com/leefowlercu/nomad-mcp-pack/internal/utils"
	"github.com/leefowlercu/nomad-mcp-pack/internal/validate"
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
	Version:           version,
	PersistentPreRunE: runInitAndValidate,
}

func init() {
	nomadMcpPackCmd.PersistentFlags().String("registry-url", config.DefaultConfig.RegistryURL, "Registry URL")
	nomadMcpPackCmd.PersistentFlags().String("output-dir", config.DefaultConfig.OutputDir, "Output directory for generated packs")
	nomadMcpPackCmd.PersistentFlags().String("output-type", config.DefaultConfig.OutputType, "Output type {packdir|archive}")
	nomadMcpPackCmd.PersistentFlags().Bool("dry-run", config.DefaultConfig.DryRun, "Show what would be generated without writing files")
	nomadMcpPackCmd.PersistentFlags().Bool("force-overwrite", config.DefaultConfig.ForceOverwrite, "Overwrite existing pack or archive if it exists")
	nomadMcpPackCmd.PersistentFlags().Bool("allow-deprecated", config.DefaultConfig.AllowDeprecated, "Allow generation of packs for deprecated servers")
	nomadMcpPackCmd.PersistentFlags().BoolP("silent", "s", config.DefaultConfig.Silent, "Suppress user-facing output (errors still shown)")

	viper.BindPFlag("registry_url", nomadMcpPackCmd.PersistentFlags().Lookup("registry-url"))
	viper.BindPFlag("output_dir", nomadMcpPackCmd.PersistentFlags().Lookup("output-dir"))
	viper.BindPFlag("output_type", nomadMcpPackCmd.PersistentFlags().Lookup("output-type"))
	viper.BindPFlag("dry_run", nomadMcpPackCmd.PersistentFlags().Lookup("dry-run"))
	viper.BindPFlag("force_overwrite", nomadMcpPackCmd.PersistentFlags().Lookup("force-overwrite"))
	viper.BindPFlag("allow_deprecated", nomadMcpPackCmd.PersistentFlags().Lookup("allow-deprecated"))
	viper.BindPFlag("silent", nomadMcpPackCmd.PersistentFlags().Lookup("silent"))

	nomadMcpPackCmd.AddCommand(cmdgenerate.GenerateCmd)
	nomadMcpPackCmd.AddCommand(cmdserver.ServerCmd)
	nomadMcpPackCmd.AddCommand(cmdwatch.WatchCmd)
}

func Execute() error {
	// Disable Cobra's default error and usage handling
	// We handle this in our own way in order to provide cleaner output
	// and still show usage when appropriate
	nomadMcpPackCmd.SilenceErrors = true
	nomadMcpPackCmd.SilenceUsage = true

	err := nomadMcpPackCmd.Execute()

	if err != nil {
		cmd, _, _ := nomadMcpPackCmd.Find(os.Args[1:])

		// Command will either be a leaf command or nil
		// If nil the command was the root command itself
		if cmd == nil {
			cmd = nomadMcpPackCmd
		}

		// Log to stderr
		slog.Error("command execution failed", "error", err)

		// Show user-friendly error message to stdout
		fmt.Printf("Error: %v\n", err)
		if !cmd.SilenceUsage {
			fmt.Printf("\n")
			cmd.Usage()
		}

		return err
	}

	return nil
}

func runInitAndValidate(cmd *cobra.Command, args []string) error {
	config.InitConfig()

	cfg, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	logger := utils.SetupLogger(cfg.LogLevel, cfg.Env)
	slog.SetDefault(logger)

	slog.Info("starting nomad-mcp-pack", "log_level", cfg.LogLevel, "env", cfg.Env)

	slog.Info("starting root command input validation")

	if err := validate.OutputDir(cfg.OutputDir); err != nil {
		return fmt.Errorf("could not validate output directory; %w", err)
	}

	if err := validate.OutputType(string(cfg.OutputType)); err != nil {
		return fmt.Errorf("could not validate output type; %w", err)
	}

	slog.Info("root command input validation completed successfully")

	return nil
}
