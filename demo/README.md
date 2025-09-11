# nomad-mcp-pack Generate Command Demo

This directory contains a comprehensive demonstration script for the `nomad-mcp-pack generate` command.

## Demo Script: `demo-generate.sh`

The demo script showcases all major features of the generate command including:

- ✅ **Version Resolution**: Using `@latest` syntax to find the newest active version
- ✅ **Specific Versions**: Generating packs for exact version specifications  
- ✅ **Dry Run Mode**: Preview generation without creating files
- ✅ **Output Types**: Both directory (`packdir`) and ZIP archive (`archive`) formats
- ✅ **Package Types**: NPM and OCI (Docker) with driver configuration comparison
- ✅ **Force Overwrite**: Handling existing pack conflicts with error demonstration
- ✅ **Help System**: Comprehensive help and configuration hierarchy
- ✅ **Error Handling**: Graceful handling of invalid inputs with clear messages
- ✅ **Clean Output**: Suppressed logging for presentation clarity
- ✅ **Interactive Flow**: Pause points for command explanation during presentations

## Usage

### Interactive Mode (Recommended for presentations)

```bash
./demo-generate.sh
```

The script will pause at each step, waiting for user input. This is ideal for live demonstrations where you want to explain each feature. **Key feature**: The script pauses after displaying each command but before executing it, giving you time to explain what the command will do.

### Automatic Mode (For quick testing)

```bash
./demo-generate.sh auto
```

Runs the complete demo automatically without user prompts. Useful for CI/CD testing or quick functionality verification.

### Help

```bash
./demo-generate.sh help
```

Shows usage information and available options.

## Prerequisites

1. **nomad-mcp-pack installed**: Run `make install` from the project root
2. **Internet connection**: Required to access the MCP Registry
3. **Optional**: `tree` command for better directory visualization

## Demo Servers

The demo uses two different test servers to showcase different package types:

### NPM Package Type: `io.github.pshivapr/selenium-mcp`
- ✅ Has multiple versions (0.3.8, 0.3.9) for version resolution testing
- ✅ Contains NPM package definitions for realistic pack generation
- ✅ Is actively maintained and available in the registry
- ✅ Used for @latest demonstrations and NPM-specific templates

### OCI Package Type: `io.github.domdomegg/airtable-mcp-server`
- ✅ Contains both NPM and OCI package definitions
- ✅ Allows demonstration of Docker/container-based deployments
- ✅ Shows how the same server can support multiple package types
- ✅ Used for OCI template generation and package type comparison

Using different servers ensures the demo accurately represents each package type's functionality rather than attempting to use unsupported package types.

## Demo Flow

The script runs through 8 comprehensive scenarios:

1. **Generate Pack from @latest Version** - Shows version resolution and pack examination
2. **Generate Pack from Specific Version** - Demonstrates exact version targeting
3. **Generate Pack in Dry Run Mode** - Preview functionality without file creation
4. **Generate Pack Archive** - ZIP file generation instead of directories
5. **Different Package Types** - NPM vs OCI driver configuration comparison
6. **Force Overwrite** - Handling existing pack conflicts with error demonstration
7. **Help and Configuration** - Documentation and configuration hierarchy
8. **Error Handling** - Graceful failure scenarios with invalid server names

### Enhanced Presentation Features

- **Clean Output**: All demos run with `NOMAD_MCP_PACK_LOG_LEVEL=error` to suppress verbose JSON logging
- **Command Preview**: Each command is displayed in bold purple before execution
- **Pause for Explanation**: Script waits for input after showing the command, allowing time for explanation
- **Organized Output**: All artifacts are generated to `./packs` directory for clean organization
- **Driver Comparison**: Demo 5 specifically compares NPM (`exec` driver) vs OCI (`docker` driver) configurations

## Sample Output

```bash
$ ./demo-generate.sh
=================================
nomad-mcp-pack Generate Command Demo
=================================

This demo showcases the nomad-mcp-pack generate command functionality.
We'll use io.github.pshivapr/selenium-mcp to showcase Nomad Pack Generation from an
MCP Server hosted on NPM and io.github.domdomegg/airtable-mcp-server for to showcase Nomad Pack Generation
from an MCP Server hosted on Dockerhub

ℹ  Demo output directory: ./packs
ℹ  NPM Test Server: io.github.pshivapr/selenium-mcp
ℹ  OCI Test Server: io.github.domdomegg/airtable-mcp-server

Press Enter to continue...
```

## Cleanup

The demo script automatically cleans up all generated artifacts when it exits, including:

- Generated pack directories
- ZIP archives
- Temporary demo files

This ensures a clean environment for repeated demonstrations.

## Customization

To use different test servers, modify the variables at the top of the script:

```bash
DEMO_SERVER_NPM="io.github.your-org/your-npm-server"
DEMO_SERVER_OCI="io.github.your-org/your-oci-server"
DEMO_SERVER_NPM_VERSION="1.0.0"
```

Ensure the chosen servers have:
- Multiple versions for testing `@latest` resolution
- Package definitions in supported registries (npm, pypi, oci, nuget)
- Active status in the MCP Registry
- Appropriate package types (NPM server should have npm packages, OCI server should have oci packages)

## Use Cases

- **Product Demonstrations**: Interactive mode with step-by-step explanations
- **Training Sessions**: Educational walkthrough of tool capabilities  
- **CI/CD Testing**: Automatic mode for functionality verification
- **Development**: Quick testing during feature development
- **Documentation**: Living example of tool usage patterns