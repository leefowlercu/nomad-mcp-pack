# Watch Command Demo

The watch demo script showcases all major features of the watch command including:

- ✅ **Continuous Polling**: Automatic registry monitoring with configurable intervals
- ✅ **State Persistence**: State file tracking to avoid regenerating unchanged packs
- ✅ **Server Name Filtering**: Watch only specific servers by name
- ✅ **Package Type Filtering**: Generate packs only for specific package types
- ✅ **Transport Type Filtering**: Filter by transport types (stdio, http, sse)
- ✅ **Dry Run Mode**: Preview what would be generated without creating files
- ✅ **Multiple Filters Combined**: Precise control with combined filters
- ✅ **Concurrent Generation**: Control parallel pack generation with max-concurrent
- ✅ **Force Overwrite**: Regenerate packs even when they already exist
- ✅ **Graceful Shutdown**: Proper handling of SIGINT/SIGTERM signals
- ✅ **Help System**: Comprehensive help and configuration hierarchy
- ✅ **Interactive Flow**: Pause points for command explanation during presentations

## Usage

### Interactive Mode (Recommended for presentations)

```bash
./demo.sh
```

The script will pause at each step, waiting for user input. This is ideal for live demonstrations where you want to explain each feature. The script pauses after displaying each command but before executing it, giving you time to explain what the command will do.

### Automatic Mode (For quick testing)

```bash
./demo.sh auto
```

Runs the complete demo automatically without user prompts. Useful for CI/CD testing or quick functionality verification.

### Help

```bash
./demo.sh help
```

Shows usage information and available options.

## Prerequisites

1. **nomad-mcp-pack installed**: Run `make install` from the project root
2. **Internet connection**: Required to access the MCP Registry
3. **Optional**: `jq` command for better JSON state file visualization

## Watch Demo Configuration

The demo uses these settings optimized for demonstration:

- **Poll Interval**: 30 seconds (default is 300 seconds)
- **Watch Duration**: 45 seconds per demo scenario (enough for one poll cycle)
- **Output Directory**: `./packs`
- **State File**: `./watch-demo.json`
- **Test Server**: `io.github.pshivapr/selenium-mcp` (NPM package type)

These shorter intervals allow the demo to complete quickly while still showing the full watch functionality.

## Demo Flow

The script runs through 8 comprehensive scenarios:

1. **Basic Watch Mode** - Continuous polling, pack generation, and graceful shutdown
2. **State File Persistence** - Shows state tracking and prevents unnecessary regeneration
3. **Package Type Filtering** - Filter by specific package types (npm only)
4. **Dry Run Mode** - Preview generation without creating files
5. **Multiple Filters Combined** - Combine server name, package type, and transport filters
6. **Max Concurrent Control** - Limit parallel pack generations
7. **Force Overwrite** - Regenerate existing packs with --force-overwrite
8. **Help and Configuration** - Documentation and configuration hierarchy

### Enhanced Presentation Features

- **Background Execution**: Watch runs in background with automatic graceful shutdown
- **Clean Output**: All demos run with `NOMAD_MCP_PACK_LOG_LEVEL=error` to suppress verbose logging
- **Command Preview**: Each command is displayed in bold purple before execution
- **Pause for Explanation**: Script waits for input after showing the command
- **Organized Output**: All artifacts are generated to `./packs` directory
- **State Inspection**: Demo shows how to examine the state file with `jq`

## Sample Output

```bash
$ ./demo.sh
==========================================
nomad-mcp-pack Watch Command Demo
==========================================

This demo showcases the nomad-mcp-pack watch command functionality.
Watch continuously polls the MCP Registry and automatically generates packs for
new or updated servers. It includes state tracking, filtering, and graceful shutdown.

ℹ  Demo output directory: ./packs
ℹ  Demo state file: ./watch-demo.json
ℹ  Poll interval: 30 seconds

Press Enter to continue...
```

## Cleanup

The demo script automatically cleans up all generated artifacts when it exits, including:

- Generated pack directories
- State files
- Background watch processes
- Temporary demo files

This ensures a clean environment for repeated demonstrations.

## Understanding Watch Behavior

### Long-Running Nature

Unlike the generate command which runs once and exits, watch is designed to run continuously. In the demo:

1. Watch starts in the background
2. Runs for a specified duration (45 seconds by default)
3. Receives a graceful shutdown signal (SIGINT)
4. Completes current operations and exits cleanly

This mirrors real-world usage where watch would run as a long-lived service.

### State File Tracking

The state file (`watch-demo.json`) tracks:
- Server namespace and name
- Version
- Package and transport types
- Last update timestamp
- Generation timestamp

This prevents unnecessary regeneration when:
- Server version hasn't changed
- Pack was already generated for this configuration
- No forced overwrite is requested

### Filtering Strategy

Filters work in combination:
- **Server Names**: Only watch specific servers
- **Package Types**: Only generate for specific package managers (npm, pypi, oci, nuget)
- **Transport Types**: Only generate for specific transport methods (stdio, http, sse)

All filters must match for a pack to be generated.

## Production Usage Recommendations

When using watch in production:

1. **Poll Interval**: Use longer intervals (300+ seconds) to avoid overloading the registry
2. **State File**: Place in persistent storage for resume capability
3. **Filters**: Use specific filters to reduce unnecessary processing
4. **Max Concurrent**: Adjust based on available system resources
5. **Output Directory**: Use dedicated directory with appropriate permissions
6. **Monitoring**: Run as a systemd service or in a container orchestrator

## Customization

To use different test servers, modify the variables at the top of the script:

```bash
DEMO_POLL_INTERVAL=30
DEMO_WATCH_DURATION=45
DEMO_SERVER_FILTER="io.github.your-org/your-server"
```

Ensure the chosen server:
- Exists in the MCP Registry
- Has package definitions (npm, pypi, oci, or nuget)
- Is actively maintained (not deprecated or deleted)

## Use Cases

- **Product Demonstrations**: Interactive mode with step-by-step explanations
- **Training Sessions**: Educational walkthrough of watch capabilities
- **CI/CD Testing**: Automatic mode for functionality verification
- **Development**: Quick testing during feature development
- **Documentation**: Living example of watch command usage
- **Operations Training**: Understanding production deployment patterns
