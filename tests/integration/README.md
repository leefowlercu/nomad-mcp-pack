# Integration Tests

This directory contains integration tests for `nomad-mcp-pack` that test the complete workflows from the MCP Registry through pack generation and deployment.

## Test Organization

Integration tests are organized by the integrations they validate:

### Registry ↔ nomad-mcp-pack Integration
- **`registry_generate/`** - Tests the `generate` command against the MCP Registry
- **`registry_watch/`** - Tests the `watch` command with registry polling
- **`registry_server/`** - Tests the HTTP `server` command *(planned)*

### nomad-mcp-pack ↔ Nomad Integration *(planned)*
- **`nomad_deploy/`** - Tests deploying generated packs to Nomad
- **`nomad_lifecycle/`** - Tests pack lifecycle management in Nomad

## Quick Start

### Prerequisites

1. **Docker** - Required for local MCP Registry
2. **Go 1.21+** - For running tests
3. **Git submodules** - MCP Registry submodule must be initialized

### Initialize Submodules

```bash
# Initialize the MCP Registry submodule (v1.0.0)
git submodule update --init --recursive
```

### Run All Integration Tests

```bash
# Start local registry and run all tests
make test-integration

# Or manually:
cd tests/integration/registry
docker-compose up -d
cd ../..
go test ./tests/integration/... -v
```

### Run Specific Test Suites

```bash
# Only registry_generate tests
go test ./tests/integration/registry_generate/... -v

# Only registry_watch tests
go test ./tests/integration/registry_watch/... -v

# Only registry tests
go test ./tests/integration/registry_*/... -v

# Only nomad tests (planned)
go test ./tests/integration/nomad_*/... -v
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `INTEGRATION_REGISTRY_URL` | `http://localhost:8080` | Registry URL for tests |
| `USE_LIVE_REGISTRY` | `false` | Use live registry instead of local |
| `SKIP_REGISTRY_TESTS` | `false` | Skip all registry integration tests |
| `KEEP_TEST_OUTPUT` | `false` | Keep generated test files for debugging |
| `KEEP_WATCH_LOGS` | `false` | Keep watch command logs for debugging |
| `WATCH_TEST_TIMEOUT` | `2m` | Timeout for watch tests (minimum 30s) |

### Test Registry Options

#### Option 1: Local Docker Registry (Default)

```bash
# Start local registry
cd tests/integration/registry
docker-compose up -d

# Run tests (uses localhost:8080)
go test ./tests/integration/registry_generate/... -v

# Stop registry
docker-compose down
```

#### Option 2: Live Registry

```bash
# Use live registry at registry.modelcontextprotocol.io
USE_LIVE_REGISTRY=true go test ./tests/integration/registry_generate/... -v
```

#### Option 3: Custom Registry URL

```bash
# Use custom registry URL
INTEGRATION_REGISTRY_URL=https://my-registry.example.com go test ./tests/integration/registry_generate/... -v
```

## Test Structure

### Common Utilities (`common/`)

Shared utilities for all integration tests:

- **`registry.go`** - Registry lifecycle management
- **`cli.go`** - CLI command execution helpers
- **`assertions.go`** - Test assertion utilities

### Test Patterns

Each test suite follows this pattern:

```go
func TestMain(m *testing.M) {
    // Global setup (build binary, start registry)
    code := m.Run()
    // Global cleanup
    os.Exit(code)
}

func TestFeature(t *testing.T) {
    // Setup test-specific resources
    t.Cleanup(func() { /* cleanup */ })

    // Run CLI commands using common.CLIRunner
    // Assert results using common.Assert* functions
}
```

## Registry Generate Tests (`registry_generate/`)

Tests the `generate` command against the MCP Registry.

### Test Coverage

- ✅ **Basic NPM package generation** - Standard NPM package workflow
- ✅ **Basic PyPI package generation** - Standard PyPI package workflow
- ✅ **Latest version handling** - Using `@latest` version specifier
- ✅ **Dry run mode** - Testing `--dry-run` flag
- ✅ **Archive output** - Testing `--output-type archive`
- ✅ **Invalid servers** - Error handling for nonexistent servers
- ✅ **Invalid formats** - Error handling for malformed server specs
- ✅ **Force overwrite** - Testing `--force-overwrite` flag

### Known Test Servers

The local registry includes these test servers (seed data):

| Server | Version | Package Type | Transport | Identifier |
|--------|---------|--------------|-----------|------------|
| `io.github.21st-dev/magic-mcp` | `0.0.1-seed` | `npm` | `stdio` | `@21st-dev/magic` |
| `io.github.adfin-engineering/mcp-server-adfin` | `0.0.1-seed` | `pypi` | `stdio` | `adfinmcp` |

### Running Generate Tests

```bash
# All generate tests
go test ./tests/integration/registry_generate/... -v

# Specific test
go test ./tests/integration/registry_generate/... -run TestGenerateFromRegistry_BasicNPM -v

# With debugging output
KEEP_TEST_OUTPUT=true go test ./tests/integration/registry_generate/... -v
```

## Debugging

### Keep Test Output

```bash
# Keep generated files for inspection
KEEP_TEST_OUTPUT=true go test ./tests/integration/registry_generate/... -v
```

Test files will be preserved in temporary directories and paths logged.

### Registry Health Check

```bash
# Check if local registry is healthy
curl http://localhost:8080/v0/health

# List available servers
curl http://localhost:8080/v0/servers?limit=10
```

### Binary Building

Tests automatically build the `nomad-mcp-pack` binary once and reuse it. To force rebuild:

```bash
# Clean and rebuild
go clean -testcache
go test ./tests/integration/registry_generate/... -v
```

### Registry Logs

```bash
# View registry logs
cd tests/integration/registry
docker-compose logs -f registry
```

## CI/CD Integration

### Skip Integration Tests

```bash
# Skip all integration tests (for fast CI)
go test -short ./...

# Or set environment variable
SKIP_REGISTRY_TESTS=true go test ./tests/integration/...
```

### Test with Live Registry

```bash
# Test against production registry
USE_LIVE_REGISTRY=true go test ./tests/integration/registry_generate/... -v
```

## Troubleshooting

### Common Issues

**Registry not starting:**
```bash
# Check Docker is running
docker version

# Check port not in use
lsof -i :8080

# Reset registry
cd tests/integration/registry
docker-compose down -v
docker-compose up -d
```

**Test binary build fails:**
```bash
# Check Go environment
go version
go mod tidy

# Build manually
go build -o /tmp/nomad-mcp-pack ./cmd/nomadmcppack.go
```

**Tests timing out:**
```bash
# Increase timeout
go test -timeout 5m ./tests/integration/registry_generate/... -v
```

### Getting Help

1. Check registry health: `curl http://localhost:8080/v0/health`
2. Review registry logs: `docker-compose logs registry`
3. Enable debug output: `KEEP_TEST_OUTPUT=true`
4. Check test timeout settings in `helpers_test.go`

## Registry Watch Tests (`registry_watch/`)

Tests the `watch` command against the MCP Registry with continuous polling and pack generation.

### Test Coverage

- ✅ **Basic polling** - Standard watch polling workflow
- ✅ **State management** - State file creation and persistence
- ✅ **Name filtering** - Server name pattern filtering
- ✅ **Package type filtering** - Package type filtering (npm, pypi, etc.)
- ✅ **Transport type filtering** - Transport type filtering (stdio, sse, etc.)
- ✅ **Error recovery** - Handling corrupted state files
- ✅ **Dry run mode** - Testing `--dry-run` flag
- ✅ **Signal handling** - Graceful shutdown on SIGTERM/SIGINT
- ✅ **Pre-existing packs** - Overwrite behavior testing
- ✅ **Concurrent generation** - Testing `--max-concurrent` option

### Special Considerations

**Minimum Poll Interval**: All watch tests respect the 30-second minimum poll interval configured in the application. Tests use helper functions to run the watch command for specific durations or until certain conditions are met.

**State File Management**: Tests create unique state files to avoid conflicts and test state persistence, recovery, and corruption scenarios.

**Signal Handling**: Tests verify graceful shutdown behavior when receiving SIGTERM signals.

### Running Watch Tests

```bash
# All watch tests
go test ./tests/integration/registry_watch/... -v

# Specific test
go test ./tests/integration/registry_watch/... -run TestWatch_BasicPolling -v

# With debugging output and logs
KEEP_WATCH_LOGS=true KEEP_TEST_OUTPUT=true go test ./tests/integration/registry_watch/... -v

# Custom timeout (minimum 30s)
WATCH_TEST_TIMEOUT=5m go test ./tests/integration/registry_watch/... -v
```

### Test Helper Functions

The watch tests include specialized helper functions:

- **`runWatchForDuration()`** - Runs watch command for specific time periods
- **`runWatchUntilGeneration()`** - Runs until expected number of packs generated
- **`readStateFile()`** - Parses and validates state file contents
- **`countGeneratedPacks()`** - Counts generated pack directories/archives
- **`waitForStateFile()`** - Waits for state file creation with timeout
- **`createCorruptedStateFile()`** - Creates invalid state files for error testing

## Future Test Suites

### Planned Registry Integration Tests

- **Server Tests** (`registry_server/`) - Test HTTP server API endpoints

### Planned Nomad Integration Tests

- **Deploy Tests** (`nomad_deploy/`) - Test pack deployment to Nomad
- **Lifecycle Tests** (`nomad_lifecycle/`) - Test pack start/stop/update cycles

### Contributing

When adding new integration tests:

1. Follow the established patterns in `registry_generate/`
2. Use common utilities in `common/` package
3. Add fixtures for known test data
4. Include both success and error scenarios
5. Update this documentation
