#!/bin/bash

# nomad-mcp-pack Generate Command Demo
# 
# This script demonstrates the functionality of the nomad-mcp-pack generate command
# by showcasing various features including version resolution, output types, and 
# different package types.
#
# Prerequisites:
# - nomad-mcp-pack must be installed (run 'make install' from project root)
# - Internet connection to access the MCP Registry

set -e  # Exit on any error

# Color output functions
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
BOLD_PURPLE='\033[1;35m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

print_header() {
    echo -e "${BOLD}${BLUE}==========================================${NC}"
    echo -e "${BOLD}${BLUE}$1${NC}"
    echo -e "${BOLD}${BLUE}==========================================${NC}"
}

print_step() {
    echo -e "${BOLD}${GREEN}Step: $1${NC}"
}

print_info() {
    echo -e "${CYAN}ℹ  $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠  $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_command() {
    echo -e "${BOLD_PURPLE}$ $1${NC}"
}

wait_for_user() {
    if [[ "${DEMO_AUTO:-}" != "true" ]]; then
        echo -e "\n${YELLOW}Press Enter to continue...${NC}"
        read -r
    else
        sleep 2
    fi
}

# Demo configuration
ORIGINAL_DIR=$(pwd)
DEMO_OUTPUT_DIR="./packs"
DEMO_SERVER_NPM="io.github.pshivapr/selenium-mcp"
DEMO_SERVER_OCI="io.github.domdomegg/airtable-mcp-server"
DEMO_VERSION_LATEST="latest"
DEMO_SERVER_NPM_VERSION="0.3.8"

# Set log level to error for cleaner demo output
export NOMAD_MCP_PACK_LOG_LEVEL=error

# Cleanup function
cleanup() {
    print_step "Cleaning up demo artifacts"
    cd "$ORIGINAL_DIR"
    rm -rf "$DEMO_OUTPUT_DIR"
    print_success "Demo cleanup complete"
}

# Trap cleanup on exit
trap cleanup EXIT

# Main demo function
main() {
    print_header "nomad-mcp-pack Generate Command Demo"
    
    echo -e "This demo showcases the ${BOLD}nomad-mcp-pack generate${NC} command functionality."
    echo -e "We'll use ${BOLD}${DEMO_SERVER_NPM}${NC} to showcase Nomad Pack Generation from an"
    echo -e "MCP Server hosted on NPM and ${BOLD}${DEMO_SERVER_OCI}${NC} for to showcase Nomad Pack Generation"
    echo -e "from an MCP Server hosted on Dockerhub"
    echo ""
    print_info "Demo output directory: $DEMO_OUTPUT_DIR"
    print_info "NPM Test Server: $DEMO_SERVER_NPM"
    print_info "OCI Test Server: $DEMO_SERVER_OCI"
    
    wait_for_user
    
    # Check if binary is available
    print_step "Checking nomad-mcp-pack installation"
    if ! command -v nomad-mcp-pack &> /dev/null; then
        print_error "nomad-mcp-pack is not installed or not in PATH"
        print_info "Please run 'make install' from the project root directory"
        exit 1
    fi
    
    print_command "nomad-mcp-pack --version"
    nomad-mcp-pack --version
    print_success "nomad-mcp-pack is available"
    
    wait_for_user
    
    # Create demo directory
    print_step "Setting up demo output directory"
    mkdir -p "$DEMO_OUTPUT_DIR"
    print_success "Demo output directory created: $DEMO_OUTPUT_DIR"
    
    wait_for_user
    
    # Demo 1: Generate with @latest
    print_header "Demo 1: Generate Pack from @latest Version"
    echo "This demonstrates version resolution using the @latest keyword syntax."
    echo "The command will find the latest version of the server. If the latest version is not marked"
    echo "'deprecated' or 'deleted' then a Nomad Pack will be generated."
    echo ""
    
    print_command "nomad-mcp-pack generate ${DEMO_SERVER_NPM}@${DEMO_VERSION_LATEST} --package-type npm --output-dir $DEMO_OUTPUT_DIR"
    wait_for_user
    nomad-mcp-pack generate "${DEMO_SERVER_NPM}@${DEMO_VERSION_LATEST}" --package-type npm --output-dir $DEMO_OUTPUT_DIR
    
    print_success "Pack generated successfully!\n"
    print_info "Let's examine what was created:"
    
    PACK_DIR=$(find "$DEMO_OUTPUT_DIR" -maxdepth 1 -name "*selenium-mcp*" -type d | head -1)
    if [[ -n "$PACK_DIR" ]]; then
        print_command "tree $PACK_DIR"
        tree "$PACK_DIR" 2>/dev/null || find "$PACK_DIR" -type f | sort
        
        echo ""
        print_info "Let's look at the generated metadata:"
        print_command "cat $PACK_DIR/metadata.hcl"
        cat "$PACK_DIR/metadata.hcl"
        echo ""
    else
        print_warning "Could not find generated pack directory. Let's see what was created:"
        print_command "ls -la $DEMO_OUTPUT_DIR"
        ls -la "$DEMO_OUTPUT_DIR"
    fi
    
    wait_for_user
    
    # Demo 2: Generate specific version
    print_header "Demo 2: Generate Pack from specific Version"
    echo "Now let's generate a pack for a specific version to show version handling."
    echo ""
    
    print_command "nomad-mcp-pack generate ${DEMO_SERVER_NPM}@${DEMO_SERVER_NPM_VERSION} --package-type npm --output-dir $DEMO_OUTPUT_DIR"
    wait_for_user
    nomad-mcp-pack generate "${DEMO_SERVER_NPM}@${DEMO_SERVER_NPM_VERSION}" --package-type npm --output-dir $DEMO_OUTPUT_DIR
    
    print_success "Specific version pack generated!"
    
    wait_for_user
    
    # Demo 3: Dry run mode
    print_header "Demo 3: Generate Pack in Dry Run Mode"
    echo "The --dry-run flag shows what would be generated without creating files."
    echo "This is useful for previewing or validation."
    echo ""
    
    print_command "nomad-mcp-pack generate ${DEMO_SERVER_NPM}@latest --package-type npm --output-dir $DEMO_OUTPUT_DIR --dry-run"
    wait_for_user
    nomad-mcp-pack generate "${DEMO_SERVER_NPM}@latest" --package-type npm --output-dir $DEMO_OUTPUT_DIR --dry-run
    
    print_success "Dry run completed - no files were created"
    
    wait_for_user
    
    # Demo 4: Archive output type
    print_header "Demo 4: Generate Pack Archive"
    echo "Generate a pack as a ZIP archive instead of a directory."
    echo ""
    
    print_command "nomad-mcp-pack generate ${DEMO_SERVER_NPM}@latest --package-type npm --output-dir $DEMO_OUTPUT_DIR --output-type archive"
    wait_for_user
    nomad-mcp-pack generate "${DEMO_SERVER_NPM}@latest" --package-type npm --output-dir $DEMO_OUTPUT_DIR --output-type archive
    
    print_success "Archive generated!\n"
    print_info "Let's see the archive that was created:"
    print_command "ls -la ${DEMO_OUTPUT_DIR}/*.zip"
    ls -la "${DEMO_OUTPUT_DIR}"/*.zip 2>/dev/null || print_info "No zip files found"
    
    wait_for_user
    
    # Demo 5: Different package types
    print_header "Demo 5: Different Package Types"
    echo "Generate packs for different package types using a server that hosted on Dockerhub (OCI)."
    echo "We'll use ${DEMO_SERVER_OCI} which has both npm and oci packages."
    echo ""
    
    print_command "nomad-mcp-pack generate ${DEMO_SERVER_OCI}@latest --package-type oci --output-dir $DEMO_OUTPUT_DIR"
    wait_for_user
    nomad-mcp-pack generate "${DEMO_SERVER_OCI}@latest" --package-type oci --output-dir $DEMO_OUTPUT_DIR
    
    print_success "OCI/Docker version generated!"
    
    print_info "Let's compare the job templates between npm and oci versions:"
    NPM_JOB=$(find "$DEMO_OUTPUT_DIR" -path "*selenium-mcp*" -name "*.nomad.tpl" | head -1)
    OCI_JOB=$(find "$DEMO_OUTPUT_DIR" -path "*airtable-mcp-server*" -name "*.nomad.tpl" | head -1)
    
    if [[ -n "$NPM_JOB" && -n "$OCI_JOB" ]]; then
        echo -e "\n${BOLD}NPM Job Template (driver configuration):${NC}"
        grep -A 10 "driver.*=" "$NPM_JOB" | head -15
        echo -e "\n${BOLD}OCI Job Template (driver configuration):${NC}"
        grep -A 10 "driver.*=" "$OCI_JOB" | head -15
    fi
    
    wait_for_user
    
    # Demo 6: Force overwrite
    print_header "Demo 6: Force Overwrite"
    echo "Demonstrate the --force flag to overwrite existing packs."
    echo ""
    
    print_info "First, let's try to generate to the same location (should fail):"
    print_command "nomad-mcp-pack generate ${DEMO_SERVER_NPM}@latest --package-type npm --output-dir $DEMO_OUTPUT_DIR"
    wait_for_user
    set +e  # Don't exit on error for this demo
    nomad-mcp-pack generate "${DEMO_SERVER_NPM}@latest" --package-type npm --output-dir $DEMO_OUTPUT_DIR
    RESULT=$?
    set -e
    
    if [[ $RESULT -ne 0 ]]; then
        print_success "Command failed as expected (pack already exists)\n"
        
        print_info "Now let's use --force to overwrite:"
        print_command "nomad-mcp-pack generate ${DEMO_SERVER_NPM}@latest --package-type npm --output-dir $DEMO_OUTPUT_DIR --force"
        wait_for_user
        nomad-mcp-pack generate "${DEMO_SERVER_NPM}@latest" --package-type npm --output-dir $DEMO_OUTPUT_DIR --force
        print_success "Pack overwritten successfully with --force!"
    fi
    
    wait_for_user
    
    # Demo 7: Show help and configuration
    print_header "Demo 7: Help and Configuration"
    echo "The tool provides comprehensive help and supports configuration."
    echo ""
    
    print_command "nomad-mcp-pack generate --help | head -20"
    wait_for_user
    nomad-mcp-pack generate --help | head -20
    
    print_info "Configuration can be provided via:"
    echo "  • Command-line flags (highest priority)"
    echo "  • Environment variables (NOMAD_MCP_PACK_*)"
    echo "  • Configuration file (config.yaml)"
    echo "  • Built-in defaults (lowest priority)"
    
    wait_for_user
    
    # Demo 8: Error handling
    print_header "Demo 8: Error Handling"
    echo "Let's see how the tool handles invalid inputs."
    echo ""
    
    print_info "Trying to generate a pack for a non-existent server:"
    print_command "nomad-mcp-pack generate non.existent/server@latest"
    wait_for_user
    set +e
    nomad-mcp-pack generate non.existent/server@latest 2>&1 | head -5
    set -e
    print_success "Tool provides clear error messages for invalid inputs"
    
    wait_for_user
    
    # Summary
    print_header "Demo Summary"
    print_success "Successfully demonstrated nomad-mcp-pack generate functionality!"
    echo ""
    echo "Key features showcased:"
    echo "  ✓ Version resolution with @latest syntax"
    echo "  ✓ Specific version generation"
    echo "  ✓ Dry-run mode for preview"
    echo "  ✓ Multiple output types (directory/archive)"
    echo "  ✓ Different package types (npm, oci)"
    echo "  ✓ Force overwrite capability"
    echo "  ✓ Comprehensive help system"
    echo "  ✓ Error handling and validation"
    echo ""
    print_info "Generated artifacts will be cleaned up automatically"
    print_info "Demo completed successfully!"
}

# Handle command line arguments
case "${1:-}" in
    "auto")
        export DEMO_AUTO=true
        print_info "Running in automatic mode (no user interaction)"
        ;;
    "help"|"-h"|"--help")
        echo "Usage: $0 [auto|help]"
        echo ""
        echo "  auto    Run demo in automatic mode (no user prompts)"
        echo "  help    Show this help message"
        echo ""
        echo "Default behavior is interactive mode with user prompts."
        exit 0
        ;;
esac

# Run the main demo
main
