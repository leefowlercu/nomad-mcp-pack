package generator

import (
	"bytes"
	"embed"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	v0 "github.com/modelcontextprotocol/registry/pkg/api/v0"
	"github.com/modelcontextprotocol/registry/pkg/model"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

var (
	metadataTemplate  *template.Template
	variablesTemplate *template.Template
	outputsTemplate   *template.Template
	readmeTemplate    *template.Template
	jobTemplates      map[string]*template.Template
)

func init() {
	var err error

	funcMap := template.FuncMap{
		"lower":   strings.ToLower,
		"upper":   strings.ToUpper,
		"replace": strings.ReplaceAll,
		"base":    filepath.Base,
	}

	metadataTemplate, err = template.New("metadata.hcl.tmpl").Funcs(funcMap).ParseFS(templateFS, "templates/metadata.hcl.tmpl")
	if err != nil {
		panic(fmt.Sprintf("failed to parse metadata template; %v", err))
	}

	variablesTemplate, err = template.New("variables.hcl.tmpl").Funcs(funcMap).ParseFS(templateFS, "templates/variables.hcl.tmpl")
	if err != nil {
		panic(fmt.Sprintf("failed to parse variables template; %v", err))
	}

	outputsTemplate, err = template.New("outputs.tpl.tmpl").Funcs(funcMap).ParseFS(templateFS, "templates/outputs.tpl.tmpl")
	if err != nil {
		panic(fmt.Sprintf("failed to parse outputs template; %v", err))
	}

	readmeTemplate, err = template.New("readme.md.tmpl").Funcs(funcMap).ParseFS(templateFS, "templates/readme.md.tmpl")
	if err != nil {
		panic(fmt.Sprintf("failed to parse readme template; %v", err))
	}

	jobTemplates = make(map[string]*template.Template)
	packageTypes := []string{"oci", "npm", "pypi", "nuget"}

	for _, pkgType := range packageTypes {
		templatePath := fmt.Sprintf("templates/job-%s.nomad.tmpl", pkgType)
		templateName := fmt.Sprintf("job-%s.nomad.tmpl", pkgType)
		tmpl, err := template.New(templateName).Funcs(funcMap).ParseFS(templateFS, templatePath)
		if err != nil {
			panic(fmt.Sprintf("failed to parse %s job template; %v", pkgType, err))
		}
		jobTemplates[pkgType] = tmpl
	}
}

type MetadataData struct {
	PackName        string
	PackDescription string
	PackVersion     string
	AppURL          string
	ServerName      string
}

type VariablesData struct {
	ServerName     string
	PackageType    string
	HasEnvironment bool
	Environment    []model.KeyValueInput
	HasArguments   bool
	Arguments      []model.Argument
	PackageID      string
	PackageVersion string
}

type JobData struct {
	ServerName     string
	TaskName       string
	PackageType    string
	PackageID      string
	PackageVersion string
	RegistryURL    string
	RunTimeHint    string
	Environment    []model.KeyValueInput
	RuntimeArgs    []model.Argument
	PackageArgs    []model.Argument
	Transport      model.Transport
	HasTransport   bool
}

type ReadmeData struct {
	ServerName    string
	Description   string
	PackageType   string
	PackageID     string
	Version       string
	RepositoryURL string
	HasRepository bool
}

func renderMetadataTemplate(server *v0.ServerJSON, pkg *model.Package) (string, error) {
	packName := computePackName(server.Name, server.Version, pkg.RegistryType, pkg.Transport.Type)

	appURL := ""
	if server.Repository.URL != "" {
		appURL = server.Repository.URL
	}

	data := MetadataData{
		PackName:        packName,
		PackDescription: server.Description,
		PackVersion:     server.Version,
		AppURL:          appURL,
		ServerName:      server.Name,
	}

	var buf bytes.Buffer
	if err := metadataTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute metadata template; %w", err)
	}

	return buf.String(), nil
}

func renderVariablesTemplate(server *v0.ServerJSON, pkg *model.Package) (string, error) {
	data := VariablesData{
		ServerName:     server.Name,
		PackageType:    pkg.RegistryType,
		HasEnvironment: len(pkg.EnvironmentVariables) > 0,
		Environment:    pkg.EnvironmentVariables,
		HasArguments:   len(pkg.RuntimeArguments) > 0 || len(pkg.PackageArguments) > 0,
		PackageID:      pkg.Identifier,
		PackageVersion: pkg.Version,
	}

	data.Arguments = append(pkg.RuntimeArguments, pkg.PackageArguments...)

	var buf bytes.Buffer
	if err := variablesTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute variables template; %w", err)
	}

	return buf.String(), nil
}

func renderJobTemplate(server *v0.ServerJSON, pkg *model.Package) (string, error) {
	tmpl, exists := jobTemplates[pkg.RegistryType]
	if !exists {
		return "", fmt.Errorf("no job template found for package type; %s", pkg.RegistryType)
	}

	registryURL := pkg.RegistryBaseURL
	if registryURL == "" {
		switch pkg.RegistryType {
		case "npm":
			registryURL = "https://registry.npmjs.org"
		case "pypi":
			registryURL = "https://pypi.org"
		case "oci":
			registryURL = "https://index.docker.io"
		case "nuget":
			registryURL = "https://api.nuget.org"
		}
	}

	data := JobData{
		ServerName:     server.Name,
		TaskName:       sanitizeServerName(server.Name),
		PackageType:    pkg.RegistryType,
		PackageID:      pkg.Identifier,
		PackageVersion: pkg.Version,
		RegistryURL:    registryURL,
		RunTimeHint:    pkg.RunTimeHint,
		Environment:    pkg.EnvironmentVariables,
		RuntimeArgs:    pkg.RuntimeArguments,
		PackageArgs:    pkg.PackageArguments,
		Transport:      pkg.Transport,
		HasTransport:   pkg.Transport.Type != "",
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute job template for %s; %w", pkg.RegistryType, err)
	}

	return buf.String(), nil
}

func renderOutputsTemplate(server *v0.ServerJSON) (string, error) {
	data := map[string]interface{}{
		"ServerName": server.Name,
	}

	var buf bytes.Buffer
	if err := outputsTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute outputs template; %w", err)
	}

	return buf.String(), nil
}

func renderReadmeTemplate(server *v0.ServerJSON, pkg *model.Package) (string, error) {
	data := ReadmeData{
		ServerName:    server.Name,
		Description:   server.Description,
		PackageType:   pkg.RegistryType,
		PackageID:     pkg.Identifier,
		Version:       server.Version,
		RepositoryURL: server.Repository.URL,
		HasRepository: server.Repository.URL != "",
	}

	var buf bytes.Buffer
	if err := readmeTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute readme template; %w", err)
	}

	return buf.String(), nil
}

// Helper functions for templates
func formatEnvironmentVars(envVars []model.KeyValueInput) string {
	if len(envVars) == 0 {
		return ""
	}

	var parts []string
	for _, env := range envVars {
		if env.IsSecret {
			parts = append(parts, fmt.Sprintf("%s = var.%s", env.Name, strings.ToLower(env.Name)))
		} else if env.Default != "" {
			parts = append(parts, fmt.Sprintf("%s = var.%s", env.Name, strings.ToLower(env.Name)))
		} else {
			parts = append(parts, fmt.Sprintf("%s = var.%s", env.Name, strings.ToLower(env.Name)))
		}
	}

	return strings.Join(parts, "\n        ")
}

func formatArguments(args []model.Argument) string {
	if len(args) == 0 {
		return ""
	}

	var parts []string
	for _, arg := range args {
		if arg.Type == model.ArgumentTypePositional {
			parts = append(parts, fmt.Sprintf("var.%s", strings.ToLower(arg.Name)))
		} else {
			if arg.Name != "" {
				parts = append(parts, fmt.Sprintf("--%s", arg.Name))
				parts = append(parts, fmt.Sprintf("var.%s", strings.ToLower(arg.Name)))
			}
		}
	}

	return strings.Join(parts, "\", \"")
}
