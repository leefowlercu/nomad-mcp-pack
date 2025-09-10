# Integration Tests

This directory contains integration tests for the `nomad-mcp-pack` Registry Client. These tests run against a real MCP Registry instance to verify that our client works correctly with the actual API.

## Prerequisites

### 1. Docker and Docker Compose

Make sure you have Docker and Docker Compose installed:

```bash
# Check Docker
docker --version

# Check Docker Compose
docker-compose --version
```

### 2. Local MCP Registry

The integration tests require a local MCP Registry running on `http://localhost:8080`. You can start this using the official MCP Registry docker-compose setup.

#### Starting the Local Registry

The registry is included as a Git submodule in this directory. Use the Makefile targets from the project root:

```bash
# From project root
make registry-up
```

This will automatically initialize the submodule if needed and start the registry. Wait for the registry to be ready:

```bash
# Check if it's running
curl http://localhost:8080/v0/health
```

You should see a response like:
```json
{
  "status": "ok",
  "github_client_id": "..."
}
```

#### Stopping the Local Registry

```bash
# From project root
make registry-down
```

## Running Integration Tests

### Using Make (Recommended)

From the project root directory:

```bash
# Run integration tests
make test-integration

# Run integration tests with verbose output
make test-integration-verbose

# Run only unit tests (skip integration)
make test-unit

# Start local registry
make registry-up

# Stop local registry
make registry-down
```

### Using Go Commands Directly

```bash
# Run integration tests from project root
go test -v ./tests/integration

# Run with timeout
go test -timeout 60s -v ./tests/integration

# Run specific test
go test -v ./tests/integration -run TestIntegrationListServers

# Run in short mode (skips integration tests)
go test -short -v ./tests/integration
```

## Environment Variables

The integration tests support several environment variables for configuration:

| Variable | Default | Description |
|----------|---------|-------------|
| `INTEGRATION_TEST_REGISTRY_URL` | `http://localhost:8080` | URL of the MCP Registry to test against |
| `SKIP_INTEGRATION_TESTS` | `false` | Set to `true` to forcibly skip integration tests |
| `INTEGRATION_TEST_TIMEOUT` | `30s` | Timeout for individual test operations |

### Examples

```bash
# Test against a different registry URL
INTEGRATION_TEST_REGISTRY_URL=http://registry.example.com:8080 go test -v ./tests/integration

# Skip integration tests
SKIP_INTEGRATION_TESTS=true go test -v ./tests/integration

# Use longer timeout
INTEGRATION_TEST_TIMEOUT=60s go test -v ./tests/integration
```

## Test Categories

### Core API Tests
- **TestIntegrationListServers**: Tests server listing with various options
- **TestIntegrationSearchServers**: Tests server search functionality
- **TestIntegrationGetLatestServers**: Tests filtering for latest versions
- **TestIntegrationGetUpdatedServers**: Tests time-based filtering
- **TestIntegrationGetLatestActiveServer**: Tests semantic version resolution
- **TestIntegrationGetServerByNameAndVersion**: Tests specific server retrieval

### Edge Case Tests
- **TestIntegrationPagination**: Tests cursor-based pagination
- **TestIntegrationErrorHandling**: Tests error scenarios and timeout handling
- **TestIntegrationConcurrency**: Tests thread-safety of the client

## Test Behavior

### Automatic Skipping

Integration tests will automatically skip in these scenarios:
1. Running with `go test -short`
2. Environment variable `SKIP_INTEGRATION_TESTS=true`
3. Local registry not available at the configured URL

### Test Data

The tests are designed to work with whatever data is available in the registry. They:
- Don't create or modify data (read-only tests)
- Adapt to available servers in the registry
- Skip tests that require specific data if it's not available
- Use realistic search terms and parameters

### Cleanup

Since these are read-only tests, no cleanup is required. The tests don't modify the registry state.

## Registry Submodule

The MCP Registry is included as a Git submodule at `tests/integration/registry/` pinned to v1.0.0 to ensure consistent integration testing. The submodule is automatically initialized when using the Makefile targets.

### Manual Submodule Operations

**Initialize submodule manually:**
```bash
# From project root
git submodule update --init tests/integration/registry
```

**Update registry to latest version:**
```bash
# From project root
make registry-update
git commit -m "chore: update registry submodule to latest version"
```

**Check submodule status:**
```bash
git submodule status tests/integration/registry
```

**Working with the registry directly:**
```bash
# Navigate to submodule
cd tests/integration/registry

# Check current version
git describe --tags

# View available tags
git tag | grep -E "^v[0-9]" | sort -V
```

## Troubleshooting

### Registry Not Starting

If the registry fails to start:

1. Check if ports 8080 and 5432 are already in use:
   ```bash
   lsof -i :8080
   lsof -i :5432
   ```

2. Check Docker logs:
   ```bash
   docker-compose logs registry
   docker-compose logs postgres
   ```

3. Ensure you have enough disk space and memory for PostgreSQL

### Tests Timing Out

If tests are timing out:

1. Increase the timeout:
   ```bash
   INTEGRATION_TEST_TIMEOUT=60s go test -v ./tests/integration
   ```

2. Check registry performance:
   ```bash
   curl -w "Time: %{time_total}s\n" http://localhost:8080/v0/health
   ```

3. Check if the registry is under load or needs restart

### No Test Data Available

If tests are skipping due to no available data:

1. Check if the registry has any servers:
   ```bash
   curl http://localhost:8080/v0/servers
   ```

2. The registry might be empty in a fresh setup - this is normal
3. Tests will skip gracefully when no suitable test data is available

### Connection Refused

If you get "connection refused" errors:

1. Verify the registry is running:
   ```bash
   docker ps | grep registry
   ```

2. Check the correct URL:
   ```bash
   curl http://localhost:8080/v0/health
   ```

3. Verify no firewall is blocking the connection

## Adding New Integration Tests

When adding new integration tests:

1. Follow the naming convention: `TestIntegration<FeatureName>`
2. Always call `skipIfNoRegistry(t)` at the start
3. Use the helper functions from `helpers_test.go`
4. Make tests resilient to different registry states
5. Include proper error handling and logging
6. Test both success and failure paths when possible

### Example Test Template

```go
func TestIntegrationNewFeature(t *testing.T) {
    skipIfNoRegistry(t)
    
    client := createTestClient(t)
    checkRegistryHealth(t, client)
    
    ctx, cancel := context.WithTimeout(context.Background(), getTestTimeout())
    defer cancel()
    
    // Your test logic here
    
    resp, err := client.SomeMethod(ctx, params)
    assertNoError(t, err, "SomeMethod failed")
    assertNotNil(t, resp, "Response should not be nil")
    
    // Additional assertions...
}
```

## Performance Testing

The integration tests include basic performance and concurrency tests. For more extensive performance testing:

```bash
# Run benchmarks
go test -bench=. -v ./tests/integration

# Run with CPU profiling
go test -bench=. -cpuprofile=cpu.prof ./tests/integration

# Run with memory profiling  
go test -bench=. -memprofile=mem.prof ./tests/integration
```

## CI/CD Integration

For continuous integration:

1. Tests automatically skip when registry is not available
2. Use `go test -short` to skip integration tests in fast CI runs
3. Set `SKIP_INTEGRATION_TESTS=true` to disable integration tests
4. Consider running integration tests only on specific branches or schedules