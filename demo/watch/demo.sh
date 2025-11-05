#!/bin/bash

# nomad-mcp-pack Watch Command Demo
#
# This script demonstrates the functionality of the nomad-mcp-pack watch command
# by showcasing various features including continuous polling, filtering, state
# persistence, and graceful shutdown.
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
    echo -e "${CYAN}ℹ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
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
DEMO_STATE_FILE="./watch-demo.json"
DEMO_POLL_INTERVAL=30
DEMO_WATCH_DURATION=45  # Seconds to run watch before stopping
DEMO_SERVER_FILTER="io.github.pshivapr/selenium-mcp"

# Set log level to error for cleaner demo output
export NOMAD_MCP_PACK_LOG_LEVEL=error

# Track background processes for cleanup
WATCH_PIDS=()

# Cleanup function
cleanup() {
    echo ""
    print_step "Cleaning up demo artifacts"

    # Kill any running watch processes
    for pid in "${WATCH_PIDS[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            print_info "Stopping watch process (PID: $pid)"
            kill -INT "$pid" 2>/dev/null || true
            wait "$pid" 2>/dev/null || true
        fi
    done

    cd "$ORIGINAL_DIR"
    rm -rf "$DEMO_OUTPUT_DIR"
    rm -f "$DEMO_STATE_FILE"
    print_success "Demo cleanup complete"
}

# Trap cleanup on exit
trap cleanup EXIT

# Function to start watch in background and wait for duration
run_watch_for_duration() {
    local duration=$1
    shift
    local cmd_args=("$@")

    # Start watch in background
    nomad-mcp-pack watch "${cmd_args[@]}" &
    local watch_pid=$!
    WATCH_PIDS+=("$watch_pid")

    print_info "Watch started (PID: $watch_pid), running for ${duration} seconds..."

    # Wait for the specified duration
    sleep "$duration"

    # Send graceful shutdown signal
    print_info "Sending shutdown signal..."
    kill -INT "$watch_pid" 2>/dev/null || true

    # Wait for process to exit
    wait "$watch_pid" 2>/dev/null || true

    print_success "Watch stopped gracefully"
}

# Main demo function
main() {
    print_header "nomad-mcp-pack Watch Command Demo"

    echo -e "This demo showcases the ${BOLD}nomad-mcp-pack watch${NC} command functionality."
    echo -e "Watch continuously polls the MCP Registry and automatically generates packs for"
    echo -e "new or updated servers. It includes state tracking, filtering, and graceful shutdown."
    echo ""
    print_info "Demo output directory: $DEMO_OUTPUT_DIR"
    print_info "Demo state file: $DEMO_STATE_FILE"
    print_info "Poll interval: $DEMO_POLL_INTERVAL seconds"

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
    print_step "Setting up demo environment"
    mkdir -p "$DEMO_OUTPUT_DIR"
    print_success "Demo output directory created: $DEMO_OUTPUT_DIR"

    wait_for_user

    # Demo 1: Basic watch mode
    print_header "Demo 1: Basic Watch Mode"
    echo "This demonstrates the watch command polling the registry and generating packs."
    echo "Watch will run for ~45 seconds (enough for one complete poll cycle)."
    echo "We'll use a short poll interval (30 seconds) for demo purposes."
    echo ""

    print_command "nomad-mcp-pack watch --poll-interval $DEMO_POLL_INTERVAL --output-dir $DEMO_OUTPUT_DIR --state-file $DEMO_STATE_FILE --filter-server-names \"$DEMO_SERVER_FILTER\""
    wait_for_user

    run_watch_for_duration $DEMO_WATCH_DURATION \
        --poll-interval "$DEMO_POLL_INTERVAL" \
        --output-dir "$DEMO_OUTPUT_DIR" \
        --state-file "$DEMO_STATE_FILE" \
        --filter-server-names "$DEMO_SERVER_FILTER"

    print_success "Watch mode completed successfully!\n"
    print_info "Let's see what was generated:"
    print_command "ls -la $DEMO_OUTPUT_DIR"
    ls -la "$DEMO_OUTPUT_DIR" 2>/dev/null || echo "No packs generated yet"

    wait_for_user

    # Demo 2: State file persistence
    print_header "Demo 2: State File Persistence"
    echo "The watch command maintains a state file to track generated packs."
    echo "This prevents unnecessary regeneration of unchanged servers."
    echo ""

    print_info "Let's examine the state file:"
    print_command "cat $DEMO_STATE_FILE | jq ."
    if [ -f "$DEMO_STATE_FILE" ]; then
        cat "$DEMO_STATE_FILE" | jq . 2>/dev/null || cat "$DEMO_STATE_FILE"
    else
        print_warning "State file not found"
    fi

    wait_for_user

    print_info "Now let's run watch again - it should not regenerate existing packs:"
    print_command "nomad-mcp-pack watch --poll-interval $DEMO_POLL_INTERVAL --output-dir $DEMO_OUTPUT_DIR --state-file $DEMO_STATE_FILE --filter-server-names \"$DEMO_SERVER_FILTER\""
    wait_for_user

    run_watch_for_duration $DEMO_WATCH_DURATION \
        --poll-interval "$DEMO_POLL_INTERVAL" \
        --output-dir "$DEMO_OUTPUT_DIR" \
        --state-file "$DEMO_STATE_FILE" \
        --filter-server-names "$DEMO_SERVER_FILTER"

    print_success "State file prevented regeneration of existing packs!"

    wait_for_user

    # Demo 3: Package type filtering
    print_header "Demo 3: Package Type Filtering"
    echo "Filter pack generation by specific package types."
    echo "This is useful when you only want to support certain deployment methods."
    echo ""

    rm -f "$DEMO_STATE_FILE"  # Reset state for clean demo
    print_command "nomad-mcp-pack watch --poll-interval $DEMO_POLL_INTERVAL --output-dir $DEMO_OUTPUT_DIR --state-file $DEMO_STATE_FILE --filter-package-types npm --filter-transport-types stdio"
    wait_for_user

    run_watch_for_duration $DEMO_WATCH_DURATION \
        --poll-interval "$DEMO_POLL_INTERVAL" \
        --output-dir "$DEMO_OUTPUT_DIR" \
        --state-file "$DEMO_STATE_FILE" \
        --filter-package-types npm \
        --filter-transport-types stdio

    print_success "Watch completed with package type filtering!"

    wait_for_user

    # Demo 4: Dry run mode
    print_header "Demo 4: Dry Run Mode"
    echo "Dry run mode shows what would be generated without creating files."
    echo "This is useful for testing filters and understanding what watch will do."
    echo ""

    rm -f "$DEMO_STATE_FILE"  # Reset state for clean demo
    rm -rf "$DEMO_OUTPUT_DIR"/* # Clear output

    print_command "nomad-mcp-pack watch --poll-interval $DEMO_POLL_INTERVAL --output-dir $DEMO_OUTPUT_DIR --state-file $DEMO_STATE_FILE --filter-server-names \"$DEMO_SERVER_FILTER\" --dry-run"
    wait_for_user

    run_watch_for_duration $DEMO_WATCH_DURATION \
        --poll-interval "$DEMO_POLL_INTERVAL" \
        --output-dir "$DEMO_OUTPUT_DIR" \
        --state-file "$DEMO_STATE_FILE" \
        --filter-server-names "$DEMO_SERVER_FILTER" \
        --dry-run

    print_success "Dry run completed - no files were created\n"
    print_info "Verifying no packs were created:"
    print_command "ls -la $DEMO_OUTPUT_DIR"
    ls -la "$DEMO_OUTPUT_DIR" 2>/dev/null | grep -v "^total" || echo "(empty directory)"

    wait_for_user

    # Demo 5: Multiple filters combined
    print_header "Demo 5: Multiple Filters Combined"
    echo "Combine multiple filters for precise control over pack generation."
    echo "We'll filter by both server name and package type."
    echo ""

    rm -f "$DEMO_STATE_FILE"  # Reset state for clean demo
    rm -rf "$DEMO_OUTPUT_DIR"/* # Clear output

    print_command "nomad-mcp-pack watch --poll-interval $DEMO_POLL_INTERVAL --output-dir $DEMO_OUTPUT_DIR --state-file $DEMO_STATE_FILE --filter-server-names \"$DEMO_SERVER_FILTER\" --filter-package-types npm --filter-transport-types stdio"
    wait_for_user

    run_watch_for_duration $DEMO_WATCH_DURATION \
        --poll-interval "$DEMO_POLL_INTERVAL" \
        --output-dir "$DEMO_OUTPUT_DIR" \
        --state-file "$DEMO_STATE_FILE" \
        --filter-server-names "$DEMO_SERVER_FILTER" \
        --filter-package-types npm \
        --filter-transport-types stdio

    print_success "Watch completed with multiple filters!"

    wait_for_user

    # Demo 6: Max concurrent control
    print_header "Demo 6: Max Concurrent Pack Generation"
    echo "Control how many packs are generated concurrently."
    echo "This helps manage system resources during heavy generation."
    echo ""

    print_info "Default is 5 concurrent generations, let's set it to 2:"
    print_command "nomad-mcp-pack watch --poll-interval $DEMO_POLL_INTERVAL --output-dir $DEMO_OUTPUT_DIR --state-file $DEMO_STATE_FILE --max-concurrent 2"
    wait_for_user

    run_watch_for_duration 35 \
        --poll-interval "$DEMO_POLL_INTERVAL" \
        --output-dir "$DEMO_OUTPUT_DIR" \
        --state-file "$DEMO_STATE_FILE" \
        --max-concurrent 2

    print_success "Watch completed with concurrency limit!"

    wait_for_user

    # Demo 7: Help and configuration
    print_header "Demo 7: Help and Configuration"
    echo "The watch command provides comprehensive help and supports configuration."
    echo ""

    print_command "nomad-mcp-pack watch --help | head -30"
    wait_for_user
    nomad-mcp-pack watch --help | head -30

    print_info "Configuration can be provided via:"
    echo "  • Command-line flags (highest priority)"
    echo "  • Environment variables (NOMAD_MCP_PACK_*)"
    echo "  • Configuration file (config.yaml)"
    echo "  • Built-in defaults (lowest priority)"
    echo ""
    print_info "Key watch-specific settings:"
    echo "  • --poll-interval: How often to check for updates (default: 300s)"
    echo "  • --state-file: Where to track generated packs (default: ./watch.json)"
    echo "  • --max-concurrent: Parallel pack generations (default: 5)"
    echo "  • --filter-server-names: Specific servers to watch"
    echo "  • --filter-package-types: Package types to generate"
    echo "  • --filter-transport-types: Transport types to generate"

    wait_for_user

    # Summary
    print_header "Demo Summary"
    print_success "Successfully demonstrated nomad-mcp-pack watch functionality!"
    echo ""
    echo "Key features showcased:"
    echo "  ✓ Continuous polling with automatic pack generation"
    echo "  ✓ State file persistence (avoids unnecessary regeneration)"
    echo "  ✓ Server name filtering"
    echo "  ✓ Package type filtering"
    echo "  ✓ Dry-run mode for testing"
    echo "  ✓ Multiple filters combined"
    echo "  ✓ Concurrent generation control"
    echo "  ✓ Graceful shutdown handling"
    echo "  ✓ Comprehensive help system"
    echo ""
    print_info "Watch is ideal for:"
    echo "  • Automatically keeping packs up-to-date"
    echo "  • CI/CD pipelines that monitor registry changes"
    echo "  • Maintaining a complete pack registry mirror"
    echo "  • Selective pack generation with filters"
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
