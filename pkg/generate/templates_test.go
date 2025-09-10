package generate

import (
	"strings"
	"testing"

	"github.com/modelcontextprotocol/registry/pkg/model"
)

// Test template rendering functions

func TestRenderMetadataTemplate(t *testing.T) {
	server := mockServerJSON()
	serverSpec := mockServerSpec()
	pkg := &server.Packages[0] // npm package

	content, err := renderMetadataTemplate(server, serverSpec, pkg)
	if err != nil {
		t.Fatalf("Failed to render metadata template: %v", err)
	}

	// Verify content contains expected values
	expectedContents := []string{
		"test-server-name-npm", // pack name
		"Test MCP Server for unit testing", // description
		"1.0.0", // version
		"https://github.com/test/repo", // app URL
	}

	for _, expected := range expectedContents {
		if !contains(content, expected) {
			t.Errorf("Metadata template missing expected content: %q\nActual content:\n%s", expected, content)
		}
	}

	// Verify HCL structure
	if !contains(content, "app {") || !contains(content, "pack {") {
		t.Errorf("Metadata template missing HCL block structure")
	}
}

func TestRenderMetadataTemplate_NoRepository(t *testing.T) {
	server := mockServerJSON()
	server.Repository = model.Repository{} // empty repository
	serverSpec := mockServerSpec()
	pkg := &server.Packages[0]

	content, err := renderMetadataTemplate(server, serverSpec, pkg)
	if err != nil {
		t.Fatalf("Failed to render metadata template: %v", err)
	}

	// Should fallback to registry URL when no repository
	if !contains(content, "https://registry.modelcontextprotocol.io") {
		t.Errorf("Expected fallback registry URL, got content:\n%s", content)
	}
}

func TestRenderVariablesTemplate(t *testing.T) {
	server := mockServerJSON()
	pkg := &server.Packages[0] // npm package with environment vars

	content, err := renderVariablesTemplate(server, pkg)
	if err != nil {
		t.Fatalf("Failed to render variables template: %v", err)
	}

	// Verify standard variables are present
	standardVars := []string{
		"variable \"datacenters\"",
		"variable \"region\"",
		"variable \"count\"",
		"variable \"cpu\"",
		"variable \"memory\"",
	}

	for _, expected := range standardVars {
		if !contains(content, expected) {
			t.Errorf("Variables template missing standard variable: %q", expected)
		}
	}

	// Verify environment variables are included
	if !contains(content, "variable \"api_key\"") {
		t.Errorf("Variables template missing environment variable")
	}
	if !contains(content, "sensitive   = true") {
		t.Errorf("Variables template missing sensitive flag for secret variable")
	}

	// Verify runtime arguments are included
	if !contains(content, "variable \"port\"") {
		t.Errorf("Variables template missing runtime argument variable")
	}
}

func TestRenderVariablesTemplate_MinimalData(t *testing.T) {
	server := mockMinimalServerJSON()
	pkg := &server.Packages[0] // minimal package with no env vars or args

	content, err := renderVariablesTemplate(server, pkg)
	if err != nil {
		t.Fatalf("Failed to render variables template: %v", err)
	}

	// Should still have standard variables
	if !contains(content, "variable \"datacenters\"") {
		t.Errorf("Variables template missing standard variables for minimal server")
	}

	// Should not have environment or argument sections
	if contains(content, "variable \"api_key\"") {
		t.Errorf("Variables template should not have environment variables for minimal server")
	}
}

func TestRenderJobTemplate_NPM(t *testing.T) {
	server := mockServerJSON()
	pkg := &server.Packages[0] // npm package

	content, err := renderJobTemplate(server, pkg)
	if err != nil {
		t.Fatalf("Failed to render NPM job template: %v", err)
	}

	// Verify NPM-specific content
	expectedContent := []string{
		"driver = \"exec\"",
		"command = \"node\"",
		"npm://", // npm registry artifact
		"NODE_PATH", // npm-specific environment
	}

	for _, expected := range expectedContent {
		if !contains(content, expected) {
			t.Errorf("NPM job template missing expected content: %q", expected)
		}
	}

	// Verify task name is sanitized
	if !contains(content, "task \"test-server\"") {
		t.Errorf("NPM job template missing proper task name")
	}
}

func TestRenderJobTemplate_OCI(t *testing.T) {
	server := mockServerJSON()
	pkg := &server.Packages[1] // oci package

	content, err := renderJobTemplate(server, pkg)
	if err != nil {
		t.Fatalf("Failed to render OCI job template: %v", err)
	}

	// Verify OCI-specific content
	expectedContent := []string{
		"driver = \"docker\"",
		"image = \"test/image:latest\"",
		"ports = [\"mcp\"]",
	}

	for _, expected := range expectedContent {
		if !contains(content, expected) {
			t.Errorf("OCI job template missing expected content: %q", expected)
		}
	}
}

func TestRenderJobTemplate_InvalidPackageType(t *testing.T) {
	server := mockServerJSON()
	// Create a package with an unsupported type
	pkg := &model.Package{
		RegistryType: "unsupported",
		Identifier:   "test-package",
		Version:      "1.0.0",
	}

	_, err := renderJobTemplate(server, pkg)
	if err == nil {
		t.Error("Expected error for unsupported package type, got none")
	}
	if !contains(err.Error(), "no job template found") {
		t.Errorf("Expected error about missing template, got: %v", err)
	}
}

func TestRenderOutputsTemplate(t *testing.T) {
	server := mockServerJSON()

	content, err := renderOutputsTemplate(server)
	if err != nil {
		t.Fatalf("Failed to render outputs template: %v", err)
	}

	// Verify outputs template contains server name
	if !contains(content, server.Name) {
		t.Errorf("Outputs template missing server name")
	}

	// Verify template has proper structure
	if !contains(content, "Deployed") || !contains(content, "successfully") {
		t.Errorf("Outputs template missing expected structure")
	}
}

func TestRenderReadmeTemplate(t *testing.T) {
	server := mockServerJSON()
	serverSpec := mockServerSpec()
	pkg := &server.Packages[0]

	content, err := renderReadmeTemplate(server, serverSpec, pkg)
	if err != nil {
		t.Fatalf("Failed to render README template: %v", err)
	}

	// Verify README contains key information
	expectedContent := []string{
		serverSpec.ServerName, // server name in title
		server.Description, // server description
		pkg.Identifier, // package identifier
		serverSpec.Version, // version
		server.Repository.URL, // repository URL
		"## Variables", // variables section
		"## Usage", // usage section
		"nomad-pack run", // usage example
	}

	for _, expected := range expectedContent {
		if !contains(content, expected) {
			t.Errorf("README template missing expected content: %q", expected)
		}
	}
}

func TestRenderReadmeTemplate_NoRepository(t *testing.T) {
	server := mockServerJSON()
	server.Repository = model.Repository{} // no repository
	serverSpec := mockServerSpec()
	pkg := &server.Packages[0]

	content, err := renderReadmeTemplate(server, serverSpec, pkg)
	if err != nil {
		t.Fatalf("Failed to render README template: %v", err)
	}

	// Should not contain repository section
	if contains(content, "Repository") {
		t.Errorf("README template should not have repository section when no repository provided")
	}
}

// Test template data structures

func TestMetadataData_Population(t *testing.T) {
	server := mockServerJSON()
	serverSpec := mockServerSpec()
	pkg := &server.Packages[0]

	// This tests the data structure used in template rendering
	packName := sanitizePackName(serverSpec.ServerName) + "-" + pkg.RegistryType
	
	data := MetadataData{
		PackName:        packName,
		PackDescription: server.Description,
		PackVersion:     serverSpec.Version,
		AppURL:          server.Repository.URL,
		ServerName:      serverSpec.ServerName,
	}

	// Verify data is properly populated
	if data.PackName != "test-server-name-npm" {
		t.Errorf("Expected pack name 'test-server-name-npm', got %q", data.PackName)
	}
	if data.PackDescription != server.Description {
		t.Errorf("Pack description mismatch")
	}
	if data.PackVersion != serverSpec.Version {
		t.Errorf("Pack version mismatch")
	}
}

func TestVariablesData_WithEnvironmentAndArguments(t *testing.T) {
	server := mockServerJSON()
	pkg := &server.Packages[0] // npm package with env vars and args

	data := VariablesData{
		ServerName:      server.Name,
		PackageType:     pkg.RegistryType,
		HasEnvironment:  len(pkg.EnvironmentVariables) > 0,
		Environment:     pkg.EnvironmentVariables,
		HasArguments:    len(pkg.RuntimeArguments) > 0 || len(pkg.PackageArguments) > 0,
		PackageID:       pkg.Identifier,
		PackageVersion:  pkg.Version,
	}

	// Verify flags are set correctly
	if !data.HasEnvironment {
		t.Error("Expected HasEnvironment to be true")
	}
	if !data.HasArguments {
		t.Error("Expected HasArguments to be true")
	}

	// Verify data includes environment variables
	if len(data.Environment) == 0 {
		t.Error("Expected environment variables to be included")
	}
}

func TestJobData_Population(t *testing.T) {
	server := mockServerJSON()
	pkg := &server.Packages[0]

	data := JobData{
		ServerName:     server.Name,
		TaskName:       sanitizePackName(server.Name),
		PackageType:    pkg.RegistryType,
		PackageID:      pkg.Identifier,
		PackageVersion: pkg.Version,
		RegistryURL:    "https://registry.npmjs.org",
		Environment:    pkg.EnvironmentVariables,
		RuntimeArgs:    pkg.RuntimeArguments,
		PackageArgs:    pkg.PackageArguments,
		HasTransport:   pkg.Transport.Type != "",
	}

	// Verify task name is sanitized
	if data.TaskName != "test-server" {
		t.Errorf("Expected sanitized task name 'test-server', got %q", data.TaskName)
	}

	// Verify registry URL is set
	if !strings.Contains(data.RegistryURL, "npm") {
		t.Errorf("Expected NPM registry URL for npm package")
	}
}

// Integration test for template rendering consistency

func TestTemplateRendering_Integration(t *testing.T) {
	server := mockServerJSON()
	serverSpec := mockServerSpec()
	pkg := &server.Packages[0]

	// Render all templates and verify they don't error
	templates := []struct {
		name string
		fn   func() (string, error)
	}{
		{"metadata", func() (string, error) { return renderMetadataTemplate(server, serverSpec, pkg) }},
		{"variables", func() (string, error) { return renderVariablesTemplate(server, pkg) }},
		{"outputs", func() (string, error) { return renderOutputsTemplate(server) }},
		{"readme", func() (string, error) { return renderReadmeTemplate(server, serverSpec, pkg) }},
		{"job", func() (string, error) { return renderJobTemplate(server, pkg) }},
	}

	for _, tmpl := range templates {
		t.Run(tmpl.name, func(t *testing.T) {
			content, err := tmpl.fn()
			if err != nil {
				t.Errorf("Template %s failed to render: %v", tmpl.name, err)
			}
			if len(content) == 0 {
				t.Errorf("Template %s produced empty content", tmpl.name)
			}
		})
	}
}

func TestTemplateRendering_AllPackageTypes(t *testing.T) {
	packageTypes := []string{"npm", "oci"}
	
	for _, pkgType := range packageTypes {
		t.Run(pkgType, func(t *testing.T) {
			server := mockServerJSON()
			var pkg *model.Package
			
			// Find the package of the desired type
			for _, p := range server.Packages {
				if p.RegistryType == pkgType {
					pkg = &p
					break
				}
			}
			
			if pkg == nil {
				t.Fatalf("No package of type %s found in mock data", pkgType)
			}

			content, err := renderJobTemplate(server, pkg)
			if err != nil {
				t.Errorf("Failed to render job template for %s: %v", pkgType, err)
			}
			if len(content) == 0 {
				t.Errorf("Empty content for %s job template", pkgType)
			}
		})
	}
}