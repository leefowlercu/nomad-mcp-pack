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

### [Integration Demo](./integration/)

Demonstrates complete end-to-end workflow from pack generation through live deployment to a Nomad cluster.

**Key Features:**
- Live Nomad cluster deployment
- Docker-based MCP server deployment (QueryWeaver)
- Static port allocation for load balancer integration
- Automated authentication token generation
- Traefik routing configuration
- Claude Code MCP server integration
- Health checking and verification
- Complete cleanup workflow

**Quick Start:**
```bash
cd integration
./demo.sh         # Interactive mode
./demo.sh auto    # Automatic mode
```

**Important:** This demo requires a live Nomad cluster with proper access credentials. See the [Integration Demo Documentation](./integration/README.md) for detailed prerequisites.

[View Integration Demo Documentation →](./integration/README.md)

---

## Choosing the Right Demo

- **Generate Demo**: CLI-focused demonstration showcasing pack generation features. No infrastructure required. Perfect for understanding the `generate` command.

- **Watch Demo**: CLI-focused demonstration showing continuous monitoring and state management. No infrastructure required. Perfect for understanding the `watch` command.

- **Integration Demo**: Production-focused demonstration requiring a live Nomad cluster. Shows the complete workflow from MCP Registry to deployed server accessible in Claude Code. Perfect for integration testing and production validation.

| Demo | Purpose | Prerequisites | Duration | Cluster Required |
|------|---------|---------------|----------|------------------|
| **Generate** | Pack generation CLI features | nomad-mcp-pack only | 2-3 min | No |
| **Watch** | Continuous monitoring & state tracking | nomad-mcp-pack only | 3-5 min | No |
| **Integration** | End-to-end deployment workflow | Full Nomad cluster + tooling | 5-10 min | **Yes** |

## Prerequisites

### Generate and Watch Demos (CLI Testing)

These demos only require the CLI tool itself:

1. **nomad-mcp-pack installed**: Run `make install` from the project root
2. **Internet connection**: Required to access the MCP Registry
3. **Optional tools**:
   - `tree` - Enhanced directory visualization (generate demo)
   - `jq` - JSON state file visualization (watch demo)

### Integration Demo (Live Deployment)

The integration demo requires a full Nomad cluster setup:

**Required Environment Variables:**
```bash
export NOMAD_ADDR="http://your-nomad-cluster:4646"
export NOMAD_TOKEN="your-nomad-acl-token"
```

**Required Tools:**
1. **nomad-mcp-pack** - Build from this repository: `make install`
2. **nomad-pack** - HashiCorp Nomad Pack CLI
   ```bash
   # macOS
   brew install hashicorp/tap/nomad-pack

   # Linux/Other - See: https://github.com/hashicorp/nomad-pack#installation
   ```
3. **nomad** - Nomad CLI
   ```bash
   # macOS
   brew install nomad

   # Linux/Other - See: https://www.nomadproject.io/downloads
   ```
4. **jq** - JSON processor for automated token generation
   ```bash
   # macOS
   brew install jq

   # Linux (Debian/Ubuntu)
   apt-get install jq
   ```

**Nomad Cluster Requirements:**
- Nomad cluster accessible via `NOMAD_ADDR`
- ACL token with permissions to submit jobs, read allocations, and view logs
- Docker driver enabled on Nomad clients
- Internet access for Docker image pulls
- Optionally: Traefik load balancer for external access (demonstrated in integration demo)

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
- **Generate demo**: Removes `./packs/` directory and any ZIP archives
- **Watch demo**: Removes `./packs/`, `./watch-demo.json`, and background processes
- **Integration demo**: Stops Nomad job, removes `./packs/`, and cleans up authentication artifacts (`/tmp/queryweaver_cookies.txt`)

If cleanup fails for any reason, artifacts are gitignored and can be manually removed.

**Manual cleanup for integration demo:**
```bash
cd integration
nomad job stop <job-name>
rm -rf packs/
rm -f /tmp/queryweaver_cookies.txt
```

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

### Integration Testing (Integration Demo)
Verify the complete end-to-end workflow from pack generation through Nomad deployment, service registration, and MCP client configuration. Validates that all components work together correctly.

### Production Deployment Validation (Integration Demo)
Test real-world deployment scenarios including load balancer integration, authentication flows, static port allocation, and health checking before deploying to production environments.

### Training and Onboarding (Integration Demo)
Hands-on experience with the entire nomad-mcp-pack → Nomad → Claude Code workflow. Ideal for training new team members on HashiCorp Nomad and MCP server deployment patterns.

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

### Missing Environment Variables (Integration Demo)
If you see "NOMAD_ADDR environment variable is not set":
```bash
export NOMAD_ADDR="http://your-nomad-cluster:4646"
export NOMAD_TOKEN="your-token"
```

Verify these are set correctly:
```bash
echo $NOMAD_ADDR
echo $NOMAD_TOKEN
```

### Cluster Connectivity Issues (Integration Demo)
If the integration demo cannot connect to your Nomad cluster:
```bash
# Verify Nomad cluster is accessible
nomad server members
nomad node status

# Check ACL token permissions
nomad acl token self
```

### Token Generation Failures (Integration Demo)
If automated token generation fails with "Gateway Timeout" or JSON parsing errors:
- The QueryWeaver backend (FalkorDB) may not be fully initialized (takes 30-60 seconds)
- Wait for the health check to complete
- See the [Integration Demo Documentation](./integration/README.md#troubleshooting-connection-issues) for detailed troubleshooting

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
