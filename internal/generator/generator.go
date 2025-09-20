package generator

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/leefowlercu/nomad-mcp-pack/internal/output"
	v0 "github.com/modelcontextprotocol/registry/pkg/api/v0"
	"github.com/modelcontextprotocol/registry/pkg/model"
)

type Options struct {
	OutputDir      string
	OutputType     string
	DryRun         bool
	ForceOverwrite bool
}

type Generator struct {
	server   *v0.ServerJSON
	pkg      *model.Package
	packName string
	options  Options
}

func Run(ctx context.Context, srv *v0.ServerJSON, pkg *model.Package, opts Options) error {
	if srv == nil {
		return errors.New("invalid server; server cannot be nil")
	}

	if pkg == nil {
		return errors.New("invalid package; package cannot be nil")
	}

	packName := computePackName(srv.Name, srv.Version, pkg.RegistryType, pkg.Transport.Type)

	generator := &Generator{
		server:   srv,
		pkg:      pkg,
		packName: packName,
		options:  opts,
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

func (g *Generator) dryRunGenerate(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if g.options.OutputType == "archive" {
		archivePath := filepath.Join(g.options.OutputDir, g.packName+".zip")
		output.Info("Would create pack archive: %s", archivePath)
	} else {
		packDir := filepath.Join(g.options.OutputDir, g.packName)
		output.Info("Would create pack directory: %s", packDir)
	}

	return nil
}

func (g *Generator) generatePackdir(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	generateDir := filepath.Join(g.options.OutputDir, g.packName)

	if _, err := os.Stat(generateDir); err == nil && !g.options.ForceOverwrite {
		return fmt.Errorf("pack directory %s already exists: %w", generateDir, ErrPackDirectoryExists)
	}

	if err := os.MkdirAll(generateDir, 0755); err != nil {
		return fmt.Errorf("failed to create pack directory: %w", err)
	}

	templatesDir := filepath.Join(generateDir, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		return fmt.Errorf("failed to create templates directory: %w", err)
	}

	if err := g.generateFiles(ctx, generateDir); err != nil {
		return err
	}

	output.Success("Pack directory created: %s", generateDir)

	return nil
}

func (g *Generator) generateArchive(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	tempDir, err := os.MkdirTemp("", "nomad-mcp-pack-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	generateDir := filepath.Join(tempDir, g.packName)
	if err := os.MkdirAll(generateDir, 0755); err != nil {
		return fmt.Errorf("failed to create pack directory: %w", err)
	}

	templatesDir := filepath.Join(generateDir, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		return fmt.Errorf("failed to create templates directory: %w", err)
	}

	if err := g.generateFiles(ctx, generateDir); err != nil {
		return err
	}

	if err := g.createArchive(ctx, generateDir); err != nil {
		return err
	}

	archivePath := filepath.Join(g.options.OutputDir, g.packName+".zip")

	output.Success("Pack archive created: %s", archivePath)

	return nil
}

func (g *Generator) generateFiles(ctx context.Context, generateDir string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if err := g.generateMetadata(ctx, generateDir); err != nil {
		return fmt.Errorf("failed to generate metadata.hcl; %w", err)
	}

	if err := g.generateVariables(ctx, generateDir); err != nil {
		return fmt.Errorf("failed to generate variables.hcl; %w", err)
	}

	if err := g.generateOutputs(ctx, generateDir); err != nil {
		return fmt.Errorf("failed to generate outputs.tpl; %w", err)
	}

	if err := g.generateReadme(ctx, generateDir); err != nil {
		return fmt.Errorf("failed to generate README.md; %w", err)
	}

	if err := g.generateJobTemplate(ctx, generateDir); err != nil {
		return fmt.Errorf("failed to generate job template; %w", err)
	}

	if err := g.generateHelpers(ctx, generateDir); err != nil {
		return fmt.Errorf("failed to generate helpers template; %w", err)
	}

	return nil
}

func (g *Generator) generateMetadata(ctx context.Context, generateDir string) error {
	content, err := renderMetadataTemplate(g.server, g.pkg)
	if err != nil {
		return err
	}

	return g.writeFile(ctx, generateDir, "metadata.hcl", content)
}

func (g *Generator) generateVariables(ctx context.Context, generateDir string) error {
	content, err := renderVariablesTemplate(g.server, g.pkg)
	if err != nil {
		return err
	}

	return g.writeFile(ctx, generateDir, "variables.hcl", content)
}

func (g *Generator) generateOutputs(ctx context.Context, generateDir string) error {
	content, err := renderOutputsTemplate(g.server)
	if err != nil {
		return err
	}

	return g.writeFile(ctx, generateDir, "outputs.tpl", content)
}

func (g *Generator) generateReadme(ctx context.Context, generateDir string) error {
	content, err := renderReadmeTemplate(g.server, g.pkg)
	if err != nil {
		return err
	}

	return g.writeFile(ctx, generateDir, "README.md", content)
}

func (g *Generator) generateJobTemplate(ctx context.Context, generateDir string) error {
	content, err := renderJobTemplate(g.server, g.pkg)
	if err != nil {
		return err
	}

	templatePath := filepath.Join("templates", "mcp-server.nomad.tpl")
	return g.writeFile(ctx, generateDir, templatePath, content)
}

func (g *Generator) generateHelpers(ctx context.Context, generateDir string) error {
	content, err := renderHelpersTemplate()
	if err != nil {
		return err
	}

	templatePath := filepath.Join("templates", "_helpers.tpl")
	return g.writeFile(ctx, generateDir, templatePath, content)
}

func (g *Generator) writeFile(ctx context.Context, generateDir, relativePath, content string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	fullPath := filepath.Join(generateDir, relativePath)

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s; %w", dir, err)
	}

	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file %s; %w", fullPath, err)
	}

	return nil
}

func computePackName(serverName, version, packageType, transportType string) string {
	sanitized := sanitizeServerName(serverName)
	sanitizedVersion := strings.ReplaceAll(version, ".", "-")
	return fmt.Sprintf("%s-%s-%s-%s", sanitized, sanitizedVersion, packageType, transportType)
}

func sanitizeServerName(name string) string {
	sanitized := strings.ReplaceAll(name, "/", "-")
	sanitized = strings.ReplaceAll(sanitized, ".", "-")

	var result strings.Builder
	for _, r := range sanitized {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			result.WriteRune(r)
		}
	}

	return result.String()
}
