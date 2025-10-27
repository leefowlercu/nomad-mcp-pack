# nomad-mcp-pack Demos

This directory contains interactive demonstration scripts for the `nomad-mcp-pack` commands.

## Available Demos

### [Generate Command Demo](./generate/)

Demonstrates one-time pack generation for specific MCP servers.

**Key Features:**
- Version resolution with `@latest` syntax
- Specific version targeting
- Package and transport type selection
- Output formats (directory/archive)
- Dry-run mode
- Force overwrite handling
- Error handling and validation

**Quick Start:**
```bash
cd generate
./demo.sh         # Interactive mode
./demo.sh auto    # Automatic mode
```

[View Generate Demo Documentation →](./generate/README.md)

---

### [Watch Command Demo](./watch/)

Demonstrates continuous monitoring and automatic pack generation.

**Key Features:**
- Continuous polling with configurable intervals
- State file persistence (prevents unnecessary regeneration)
- Server name filtering
- Package and transport type filtering
- Concurrent generation control
- Graceful shutdown handling
- Background process management

**Quick Start:**
```bash
cd watch
./demo.sh         # Interactive mode
./demo.sh auto    # Automatic mode
```

[View Watch Demo Documentation →](./watch/README.md)

---

## Prerequisites

All demos require:
1. **nomad-mcp-pack installed**: Run `make install` from the project root
2. **Internet connection**: Required to access the MCP Registry
3. **Optional tools**:
   - `tree` - Enhanced directory visualization (generate demo)
   - `jq` - JSON state file visualization (watch demo)

## Running Demos

Each demo supports two modes:

### Interactive Mode (Recommended for presentations)

The script pauses at each step, allowing you to explain what will happen before execution.

```bash
cd <demo-directory>
./demo.sh
```

**Best for:**
- Live demonstrations
- Training sessions
- Detailed walkthroughs

### Automatic Mode (For testing)

Runs the complete demo automatically without user interaction.

```bash
cd <demo-directory>
./demo.sh auto
```

**Best for:**
- CI/CD testing
- Quick functionality verification
- Development testing

### Help

```bash
cd <demo-directory>
./demo.sh help
```

## Demo Structure

Each demo directory contains:
- `demo.sh` - Executable demonstration script
- `README.md` - Comprehensive documentation
- Generated artifacts are stored in `./packs/` (cleaned up automatically)

## Cleanup

All demos automatically clean up generated artifacts on exit:
- Generate demo: Removes `./packs/` directory and any ZIP archives
- Watch demo: Removes `./packs/`, `./watch-demo.json`, and background processes

If cleanup fails for any reason, artifacts are gitignored and can be manually removed.

## Customization

Each demo can be customized by modifying variables at the top of the `demo.sh` script:

**Generate Demo:**
```bash
DEMO_SERVER_NPM="io.github.your-org/your-npm-server"
DEMO_SERVER_OCI="io.github.your-org/your-oci-server"
DEMO_SERVER_NPM_VERSION="1.0.0"
```

**Watch Demo:**
```bash
DEMO_POLL_INTERVAL=30
DEMO_WATCH_DURATION=45
DEMO_SERVER_FILTER="io.github.your-org/your-server"
```

## Use Cases

### Product Demonstrations
Use interactive mode to showcase features with explanations between steps. Perfect for sales demos, conference talks, or customer training.

### CI/CD Testing
Use automatic mode in continuous integration pipelines to verify functionality across commits and releases.

### Training Sessions
Interactive mode provides natural pause points for explaining concepts, commands, and best practices.

### Development Testing
Quick verification during feature development using automatic mode. Test changes without manual intervention.

### Documentation
Living examples of tool usage patterns. Demo scripts serve as executable documentation that stays current with the codebase.

## Common Issues

### Demo Cleanup Fails
If cleanup fails (e.g., due to manual interruption), artifacts are gitignored. Manually clean up:
```bash
cd generate  # or watch
rm -rf packs/
rm -f *.zip *.json
```

### Port or Resource Conflicts
If watch demo shows "address already in use" or similar errors, ensure no other watch instances are running:
```bash
pkill -f "nomad-mcp-pack watch"
```

### Network Issues
Demos require internet access to the MCP Registry. If you see connection errors:
- Verify internet connectivity
- Check firewall settings
- Confirm registry URL: https://registry.modelcontextprotocol.io

## Contributing

When creating or modifying demos:
1. Place scripts in appropriate subdirectory
2. Use relative paths for all generated artifacts
3. Include automatic cleanup in trap handlers
4. Support both interactive and automatic modes
5. Update README.md documentation
6. Test in both modes before committing

## Additional Resources

- [Project README](../README.md) - Main project documentation
- [Generate Command Help](../README.md#generate-command) - Command reference
- [Watch Command Help](../README.md#watch-command) - Command reference
- [MCP Registry](https://registry.modelcontextprotocol.io/) - Official registry
- [Model Context Protocol](https://modelcontextprotocol.io/) - MCP documentation
