package registry_generate

import "github.com/leefowlercu/nomad-mcp-pack/tests/integration/common"

// TestServerFixture represents expected data for a test server
type TestServerFixture struct {
	Name             string
	Version          string
	PackageType      string
	Transport        string
	Identifier       string
	ExpectedPackName string
}

// GetTestServerFixtures returns known test server fixtures
// These correspond to servers available in the live MCP Registry
func GetTestServerFixtures() []TestServerFixture {
	return []TestServerFixture{
		{
			Name:             "io.github.kirbah/mcp-youtube",
			Version:          "0.2.6",
			PackageType:      "npm",
			Transport:        "stdio",
			Identifier:       "@kirbah/mcp-youtube",
			ExpectedPackName: common.ComputePackName("io.github.kirbah/mcp-youtube", "0.2.6", "npm", "stdio"),
		},
		{
			Name:             "io.github.huoshuiai42/huoshui-file-search",
			Version:          "1.0.0",
			PackageType:      "pypi",
			Transport:        "stdio",
			Identifier:       "huoshui-file-search",
			ExpectedPackName: common.ComputePackName("io.github.huoshuiai42/huoshui-file-search", "1.0.0", "pypi", "stdio"),
		},
	}
}

// GetNPMTestFixture returns a known NPM test server fixture
func GetNPMTestFixture() TestServerFixture {
	fixtures := GetTestServerFixtures()
	for _, fixture := range fixtures {
		if fixture.PackageType == "npm" {
			return fixture
		}
	}
	panic("No NPM test fixture found")
}

// GetPyPITestFixture returns a known PyPI test server fixture
func GetPyPITestFixture() TestServerFixture {
	fixtures := GetTestServerFixtures()
	for _, fixture := range fixtures {
		if fixture.PackageType == "pypi" {
			return fixture
		}
	}
	panic("No PyPI test fixture found")
}

// GetFixtureByName returns a test fixture by server name
func GetFixtureByName(name string) (TestServerFixture, bool) {
	fixtures := GetTestServerFixtures()
	for _, fixture := range fixtures {
		if fixture.Name == name {
			return fixture, true
		}
	}
	return TestServerFixture{}, false
}

// GetFixtureByPackageType returns the first fixture matching the package type
func GetFixtureByPackageType(packageType string) (TestServerFixture, bool) {
	fixtures := GetTestServerFixtures()
	for _, fixture := range fixtures {
		if fixture.PackageType == packageType {
			return fixture, true
		}
	}
	return TestServerFixture{}, false
}

// GetAllPackageTypes returns all package types available in test fixtures
func GetAllPackageTypes() []string {
	fixtures := GetTestServerFixtures()
	typeMap := make(map[string]bool)

	for _, fixture := range fixtures {
		typeMap[fixture.PackageType] = true
	}

	var types []string
	for packageType := range typeMap {
		types = append(types, packageType)
	}

	return types
}

// GetAllTransportTypes returns all transport types available in test fixtures
func GetAllTransportTypes() []string {
	fixtures := GetTestServerFixtures()
	typeMap := make(map[string]bool)

	for _, fixture := range fixtures {
		typeMap[fixture.Transport] = true
	}

	var types []string
	for transportType := range typeMap {
		types = append(types, transportType)
	}

	return types
}

// TestScenario represents a test scenario with expected behavior
type TestScenario struct {
	Name           string
	ServerSpec     string
	Options        []interface{} // CLI options
	ShouldSucceed  bool
	ExpectedError  string
	ExpectedFiles  []string
	Fixture        *TestServerFixture
}

// GetBasicTestScenarios returns basic test scenarios for common operations
func GetBasicTestScenarios() []TestScenario {
	npmFixture := GetNPMTestFixture()
	pypiFixture := GetPyPITestFixture()

	return []TestScenario{
		{
			Name:          "NPM package generation",
			ServerSpec:    npmFixture.Name + "@" + npmFixture.Version,
			Options:       []interface{}{},
			ShouldSucceed: true,
			ExpectedFiles: []string{"metadata.hcl", "variables.hcl", "outputs.tpl", "README.md", "templates/mcp-server.nomad.tpl"},
			Fixture:       &npmFixture,
		},
		{
			Name:          "PyPI package generation",
			ServerSpec:    pypiFixture.Name + "@" + pypiFixture.Version,
			Options:       []interface{}{},
			ShouldSucceed: true,
			ExpectedFiles: []string{"metadata.hcl", "variables.hcl", "outputs.tpl", "README.md", "templates/mcp-server.nomad.tpl"},
			Fixture:       &pypiFixture,
		},
		{
			Name:          "Latest version NPM",
			ServerSpec:    npmFixture.Name + "@latest",
			Options:       []interface{}{},
			ShouldSucceed: true,
			ExpectedFiles: []string{"metadata.hcl", "variables.hcl", "outputs.tpl", "README.md", "templates/mcp-server.nomad.tpl"},
			Fixture:       &npmFixture,
		},
		{
			Name:          "Archive output type",
			ServerSpec:    npmFixture.Name + "@" + npmFixture.Version,
			Options:       []interface{}{"--output-type", "archive"},
			ShouldSucceed: true,
			ExpectedFiles: []string{npmFixture.ExpectedPackName + ".zip"},
			Fixture:       &npmFixture,
		},
		{
			Name:          "Dry run mode",
			ServerSpec:    npmFixture.Name + "@" + npmFixture.Version,
			Options:       []interface{}{"--dry-run"},
			ShouldSucceed: true,
			ExpectedFiles: []string{}, // No files should be created
			Fixture:       &npmFixture,
		},
	}
}

// GetErrorTestScenarios returns test scenarios that should fail
func GetErrorTestScenarios() []TestScenario {
	return []TestScenario{
		{
			Name:          "Nonexistent server",
			ServerSpec:    "nonexistent-server@1.0.0",
			Options:       []interface{}{},
			ShouldSucceed: false,
			ExpectedError: "not found",
		},
		{
			Name:          "Invalid server format",
			ServerSpec:    "invalid-format",
			Options:       []interface{}{},
			ShouldSucceed: false,
			ExpectedError: "invalid format",
		},
		{
			Name:          "Empty server spec",
			ServerSpec:    "",
			Options:       []interface{}{},
			ShouldSucceed: false,
			ExpectedError: "server spec",
		},
	}
}

// GetAdvancedTestScenarios returns more complex test scenarios
func GetAdvancedTestScenarios() []TestScenario {
	npmFixture := GetNPMTestFixture()

	return []TestScenario{
		{
			Name:          "Force overwrite existing",
			ServerSpec:    npmFixture.Name + "@" + npmFixture.Version,
			Options:       []interface{}{"--force-overwrite"},
			ShouldSucceed: true,
			ExpectedFiles: []string{"metadata.hcl", "variables.hcl", "outputs.tpl", "README.md", "templates/mcp-server.nomad.tpl"},
			Fixture:       &npmFixture,
		},
		{
			Name:          "Custom package type",
			ServerSpec:    npmFixture.Name + "@" + npmFixture.Version,
			Options:       []interface{}{"--package-type", "npm"},
			ShouldSucceed: true,
			ExpectedFiles: []string{"metadata.hcl", "variables.hcl", "outputs.tpl", "README.md", "templates/mcp-server.nomad.tpl"},
			Fixture:       &npmFixture,
		},
		{
			Name:          "Custom transport type",
			ServerSpec:    npmFixture.Name + "@" + npmFixture.Version,
			Options:       []interface{}{"--transport-type", "stdio"},
			ShouldSucceed: true,
			ExpectedFiles: []string{"metadata.hcl", "variables.hcl", "outputs.tpl", "README.md", "templates/mcp-server.nomad.tpl"},
			Fixture:       &npmFixture,
		},
	}
}