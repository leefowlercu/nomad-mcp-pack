package generate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/leefowlercu/nomad-mcp-pack/internal/genutils"
	v0 "github.com/modelcontextprotocol/registry/pkg/api/v0"
	"github.com/modelcontextprotocol/registry/pkg/model"
)

type Options struct {
	OutputDir  string
	OutputType string
	DryRun     bool
	Force      bool
}

type Generator struct {
	server     *v0.ServerJSON
	serverSpec *genutils.ServerSpec
	pkg        *model.Package
	outputDir  string
	packDir    string
	options    Options
}

func Run(ctx context.Context, server *v0.ServerJSON, serverSpec *genutils.ServerSpec, packageType string, opts Options) error {
	if server == nil {
		return fmt.Errorf("server cannot be nil")
	}
	if serverSpec == nil {
		return fmt.Errorf("serverSpec cannot be nil")
	}

	var pkg *model.Package
	for _, p := range server.Packages {
		if p.RegistryType == packageType {
			pkg = &p
			break
		}
	}
	if pkg == nil {
		return fmt.Errorf("no package found for type %s", packageType)
	}

	// Create pack directory name: <server-name>-<version>-<package-type>
	packDirName := sanitizePackName(serverSpec.ServerName) + "-" + serverSpec.Version + "-" + packageType
	packDir := filepath.Join(opts.OutputDir, packDirName)

	generator := &Generator{
		server:     server,
		serverSpec: serverSpec,
		pkg:        pkg,
		outputDir:  opts.OutputDir,
		packDir:    packDir,
		options:    opts,
	}

	return generator.Generate(ctx)
}

func (g *Generator) Generate(ctx context.Context) error {
	if g.options.DryRun {
		return g.dryRunGenerate(ctx)
	}

	if g.options.OutputType == "archive" {
		return g.generateArchive(ctx)
	} else {
		return g.generatePackdir(ctx)
	}
}

func (g *Generator) generatePackdir(ctx context.Context) error {
	// Check if pack directory already exists
	if _, err := os.Stat(g.packDir); err == nil && !g.options.Force {
		return fmt.Errorf("pack directory %s already exists (use --force to overwrite)", g.packDir)
	}

	// Create pack directory
	if err := os.MkdirAll(g.packDir, 0755); err != nil {
		return fmt.Errorf("failed to create pack directory: %w", err)
	}

	// Create templates directory
	templatesDir := filepath.Join(g.packDir, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		return fmt.Errorf("failed to create templates directory: %w", err)
	}

	return g.generateFiles()
}

func (g *Generator) generateArchive(ctx context.Context) error {
	// Create temporary directory for pack files
	tempDir, err := os.MkdirTemp("", "nomad-mcp-pack-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Set pack directory to temp directory for file generation
	originalPackDir := g.packDir
	packName := sanitizePackName(g.serverSpec.ServerName) + "-" + g.serverSpec.Version + "-" + g.pkg.RegistryType
	g.packDir = filepath.Join(tempDir, packName)

	// Create pack and templates directory in temp location
	if err := os.MkdirAll(g.packDir, 0755); err != nil {
		return fmt.Errorf("failed to create temporary pack directory: %w", err)
	}

	templatesDir := filepath.Join(g.packDir, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		return fmt.Errorf("failed to create temporary templates directory: %w", err)
	}

	// Generate files to temp directory
	if err := g.generateFiles(); err != nil {
		return err
	}

	// Restore original pack dir for archive creation
	g.packDir = originalPackDir

	// Create archive from temp directory
	return g.createArchive(filepath.Join(tempDir, packName))
}

func (g *Generator) generateFiles() error {
	// Generate each file
	if err := g.generateMetadata(); err != nil {
		return fmt.Errorf("failed to generate metadata.hcl: %w", err)
	}

	if err := g.generateVariables(); err != nil {
		return fmt.Errorf("failed to generate variables.hcl: %w", err)
	}

	if err := g.generateOutputs(); err != nil {
		return fmt.Errorf("failed to generate outputs.tpl: %w", err)
	}

	if err := g.generateReadme(); err != nil {
		return fmt.Errorf("failed to generate README.md: %w", err)
	}

	if err := g.generateJobTemplate(); err != nil {
		return fmt.Errorf("failed to generate job template: %w", err)
	}

	return nil
}

func (g *Generator) dryRunGenerate(ctx context.Context) error {
	if g.options.OutputType == "archive" {
		packName := sanitizePackName(g.serverSpec.ServerName) + "-" + g.serverSpec.Version + "-" + g.pkg.RegistryType
		archivePath := filepath.Join(g.options.OutputDir, packName+".zip")
		fmt.Printf("Would create pack archive: %s\n", archivePath)
	} else {
		fmt.Printf("Would create pack directory: %s\n", g.packDir)
	}
	
	fmt.Printf("Would generate files:\n")
	fmt.Printf("  - metadata.hcl\n")
	fmt.Printf("  - variables.hcl\n")
	fmt.Printf("  - outputs.tpl\n")
	fmt.Printf("  - README.md\n")
	fmt.Printf("  - templates/mcp-server.nomad.tpl\n")

	return nil
}

func (g *Generator) generateMetadata() error {
	content, err := renderMetadataTemplate(g.server, g.serverSpec, g.pkg)
	if err != nil {
		return err
	}

	return g.writeFile("metadata.hcl", content)
}

func (g *Generator) generateVariables() error {
	content, err := renderVariablesTemplate(g.server, g.pkg)
	if err != nil {
		return err
	}

	return g.writeFile("variables.hcl", content)
}

func (g *Generator) generateOutputs() error {
	content, err := renderOutputsTemplate(g.server)
	if err != nil {
		return err
	}

	return g.writeFile("outputs.tpl", content)
}

func (g *Generator) generateReadme() error {
	content, err := renderReadmeTemplate(g.server, g.serverSpec, g.pkg)
	if err != nil {
		return err
	}

	return g.writeFile("README.md", content)
}

func (g *Generator) generateJobTemplate() error {
	content, err := renderJobTemplate(g.server, g.pkg)
	if err != nil {
		return err
	}

	templatePath := filepath.Join("templates", "mcp-server.nomad.tpl")
	return g.writeFile(templatePath, content)
}

func (g *Generator) writeFile(relativePath, content string) error {
	fullPath := filepath.Join(g.packDir, relativePath)

	// Create directory if needed
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write file
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", fullPath, err)
	}

	return nil
}

// sanitizePackName converts server names to filesystem-safe names
// Replaces '/' with '-' and removes other problematic characters
func sanitizePackName(name string) string {
	// Replace slashes with dashes
	sanitized := strings.ReplaceAll(name, "/", "-")

	// Replace dots with dashes (for reverse domain notation)
	sanitized = strings.ReplaceAll(sanitized, ".", "-")

	// Remove any other potentially problematic characters
	var result strings.Builder
	for _, r := range sanitized {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			result.WriteRune(r)
		}
	}

	return result.String()
}
