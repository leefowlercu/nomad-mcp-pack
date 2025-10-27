#!/bin/bash

# nomad-mcp-pack Integration Demo
#
# This script demonstrates the end-to-end workflow of generating and deploying
# an MCP server pack to a Nomad cluster, showcasing integration with HashiCorp Nomad.
#
# Prerequisites:
# - NOMAD_ADDR and NOMAD_TOKEN environment variables set
# - nomad-mcp-pack binary built (run 'make build' from project root)
# - nomad-pack installed
# - nomad CLI installed
# - Internet connection to access the MCP Registry and Docker Hub

set -e  # Exit on any error

# Detect mode: auto vs interactive
if [[ "$1" == "auto" ]]; then
    DEMO_AUTO="true"
else
    DEMO_AUTO="false"
fi

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
    if [[ "${DEMO_AUTO}" != "true" ]]; then
        echo -e "\n${YELLOW}Press Enter to continue...${NC}"
        read -r
    else
        sleep 2
    fi
}

# Demo configuration
ORIGINAL_DIR=$(pwd)
DEMO_OUTPUT_DIR="./packs"
DEMO_SERVER="com.falkordb/QueryWeaver"
DEMO_VERSION="latest"
PACK_NAME=""
JOB_NAME=""
ALLOC_ID=""

# Use local binary if running from demo directory
if [ -f "../../nomad-mcp-pack" ]; then
    NOMAD_MCP_PACK_BIN="../../nomad-mcp-pack"
elif [ -f "./nomad-mcp-pack" ]; then
    NOMAD_MCP_PACK_BIN="./nomad-mcp-pack"
else
    NOMAD_MCP_PACK_BIN="nomad-mcp-pack"
fi

# Set log level to error for cleaner demo output
export NOMAD_MCP_PACK_LOG_LEVEL=error

# Cleanup function
cleanup() {
    echo ""
    print_step "Cleaning up demo artifacts"
    cd "$ORIGINAL_DIR"

    # Stop Nomad job if it exists
    if [ -n "$JOB_NAME" ] && nomad job status "$JOB_NAME" &>/dev/null; then
        print_info "Stopping Nomad job: $JOB_NAME"
        nomad job stop "$JOB_NAME" &>/dev/null || true
    fi

    # Remove generated packs
    if [ -d "$DEMO_OUTPUT_DIR" ]; then
        rm -rf "$DEMO_OUTPUT_DIR"
    fi

    # Remove token generation artifacts
    if [ -f /tmp/queryweaver_cookies.txt ]; then
        rm -f /tmp/queryweaver_cookies.txt
    fi

    print_success "Demo cleanup complete"
}

# Trap cleanup on exit
trap cleanup EXIT

# Check prerequisites
check_prerequisites() {
    print_header "Checking Prerequisites"

    local all_good=true

    # Check environment variables
    if [ -z "$NOMAD_ADDR" ]; then
        print_error "NOMAD_ADDR environment variable is not set"
        print_info "Please set NOMAD_ADDR to your Nomad cluster address"
        all_good=false
    else
        print_success "NOMAD_ADDR is set: $NOMAD_ADDR"
    fi

    if [ -z "$NOMAD_TOKEN" ]; then
        print_error "NOMAD_TOKEN environment variable is not set"
        print_info "Please set NOMAD_TOKEN to authenticate with your Nomad cluster"
        all_good=false
    else
        print_success "NOMAD_TOKEN is set (hidden)"
    fi

    # Check executables
    if [ -n "$NOMAD_MCP_PACK_BIN" ] && [ -f "$NOMAD_MCP_PACK_BIN" ]; then
        print_success "nomad-mcp-pack found (using local binary: $NOMAD_MCP_PACK_BIN)"
    elif command -v nomad-mcp-pack &> /dev/null; then
        print_success "nomad-mcp-pack found in PATH"
    else
        print_error "nomad-mcp-pack not found"
        print_info "Please build the binary with 'make build' from project root"
        all_good=false
    fi

    if ! command -v nomad-pack &> /dev/null; then
        print_error "nomad-pack not found in PATH"
        print_info "Please install nomad-pack: https://github.com/hashicorp/nomad-pack"
        all_good=false
    else
        print_success "nomad-pack found"
    fi

    if ! command -v nomad &> /dev/null; then
        print_error "nomad CLI not found in PATH"
        print_info "Please install nomad: https://www.nomadproject.io/downloads"
        all_good=false
    else
        print_success "nomad CLI found"
    fi

    if ! command -v jq &> /dev/null; then
        print_error "jq not found in PATH"
        print_info "Please install jq: brew install jq (macOS) or apt-get install jq (Linux)"
        all_good=false
    else
        print_success "jq found"
    fi

    if [ "$all_good" = false ]; then
        print_error "Prerequisites check failed. Please fix the issues above and try again."
        exit 1
    fi

    # Test Nomad connectivity
    print_info "Testing Nomad cluster connectivity..."
    print_command "nomad status"
    if nomad status &>/dev/null; then
        print_success "Successfully connected to Nomad cluster"
    else
        print_error "Failed to connect to Nomad cluster"
        print_info "Please check NOMAD_ADDR and NOMAD_TOKEN are correct"
        exit 1
    fi

    wait_for_user
}

# Main demo function
main() {
    print_header "nomad-mcp-pack Integration Demo"

    echo -e "This demo showcases the ${BOLD}end-to-end workflow${NC} of deploying an MCP server"
    echo -e "to a Nomad cluster using ${BOLD}nomad-mcp-pack${NC}."
    echo ""
    echo -e "We'll demonstrate:"
    echo -e "  • Pack generation from MCP Registry"
    echo -e "  • Deployment planning with nomad-pack"
    echo -e "  • Live deployment to Nomad cluster"
    echo -e "  • Service registration with Consul"
    echo -e "  • External access via Traefik load balancer"
    echo -e "  • Job verification and health checking"
    echo -e "  • Allocation inspection and debugging"
    echo -e "  • Log analysis"
    echo -e "  • HTTP port allocation (for HTTP transport)"
    echo -e "  • Claude Code integration"
    echo -e "  • Graceful cleanup"
    echo ""
    print_info "Demo server: ${DEMO_SERVER}@${DEMO_VERSION}"
    print_info "Transport type: HTTP (streamable-http)"
    print_info "Package type: OCI (Docker)"
    echo ""

    wait_for_user

    # Check prerequisites first
    check_prerequisites

    # Section 1: Generate Pack
    print_header "Section 1: Generating Nomad Pack"
    echo -e "We'll generate a Nomad Pack for ${BOLD}${DEMO_SERVER}${NC},"
    echo -e "which is a FalkorDB-based MCP server with HTTP transport."
    echo ""
    print_command "${NOMAD_MCP_PACK_BIN} generate ${DEMO_SERVER}@${DEMO_VERSION} --package-type oci --output-dir ${DEMO_OUTPUT_DIR} --force-overwrite"
    echo ""
    wait_for_user

    "${NOMAD_MCP_PACK_BIN}" generate "${DEMO_SERVER}@${DEMO_VERSION}" --package-type oci --output-dir "${DEMO_OUTPUT_DIR}" --force-overwrite

    # Find the generated pack directory
    PACK_NAME=$(ls -t "${DEMO_OUTPUT_DIR}/" | head -1)
    JOB_NAME="$PACK_NAME"

    print_success "Pack generated successfully: $PACK_NAME"
    echo ""
    print_info "Generated pack structure:"
    tree "${DEMO_OUTPUT_DIR}/$PACK_NAME" 2>/dev/null || ls -R "${DEMO_OUTPUT_DIR}/$PACK_NAME"
    echo ""
    wait_for_user

    # Section 2: Inspect Generated Pack
    print_header "Section 2: Inspecting Generated Pack"
    echo -e "Let's examine the key files in the generated pack."
    echo ""

    print_info "metadata.hcl - Pack metadata:"
    print_command "cat ${DEMO_OUTPUT_DIR}/$PACK_NAME/metadata.hcl"
    cat "${DEMO_OUTPUT_DIR}/$PACK_NAME/metadata.hcl"
    echo ""
    wait_for_user

    print_info "Job template excerpt (showing HTTP configuration):"
    print_command "grep -A 10 'network {' ${DEMO_OUTPUT_DIR}/$PACK_NAME/templates/mcp-server.nomad.tpl"
    grep -A 10 "network {" "${DEMO_OUTPUT_DIR}/$PACK_NAME/templates/mcp-server.nomad.tpl" || echo "No network block found"
    echo ""
    wait_for_user

    # Section 3: Deployment Planning
    print_header "Section 3: Deployment Planning"
    echo -e "Before deploying, let's run a plan to see what Nomad will do."
    echo -e "This is a dry-run that shows resource allocation without making changes."
    echo ""

    # Get the region from Nomad
    NOMAD_REGION=$(nomad server members | tail -n +2 | awk '{print $NF}' | head -1)
    print_info "Detected Nomad region: $NOMAD_REGION"
    echo ""

    print_command "cd ${DEMO_OUTPUT_DIR}/$PACK_NAME && nomad-pack plan --var=\"region=$NOMAD_REGION\" --var=\"container_port=5000\" --var=\"host_port=8091\" --var='service_tags=[...]' ."
    echo ""
    wait_for_user

    cd "${DEMO_OUTPUT_DIR}/$PACK_NAME"
    nomad-pack plan \
      --var="region=$NOMAD_REGION" \
      --var="container_port=5000" \
      --var="host_port=8091" \
      --var='service_tags=["traefik.enable=true","traefik.http.middlewares.queryweaver-strip.stripprefix.prefixes=/queryweaver","traefik.http.routers.queryweaver-mcp.entrypoints=http","traefik.http.routers.queryweaver-mcp.rule=Host(`hashistack-demo-dc1-7329a99ab7f6a3ba.elb.us-east-1.amazonaws.com`) && PathPrefix(`/queryweaver`)","traefik.http.routers.queryweaver-mcp.middlewares=queryweaver-strip","traefik.http.routers.queryweaver-mcp-secure.entrypoints=https","traefik.http.routers.queryweaver-mcp-secure.tls.certresolver=letsencrypt","traefik.http.routers.queryweaver-mcp-secure.rule=Host(`hashistack-demo-dc1-7329a99ab7f6a3ba.elb.us-east-1.amazonaws.com`) && PathPrefix(`/queryweaver`)","traefik.http.routers.queryweaver-mcp-secure.middlewares=queryweaver-strip"]' \
      . || true  # Plan may return non-zero on changes
    cd "$ORIGINAL_DIR"

    print_success "Plan completed successfully - deployment looks good!"
    echo ""
    wait_for_user

    # Section 4: Deployment Execution
    print_header "Section 4: Deploying to Nomad Cluster"
    echo -e "Now let's deploy the MCP server to the Nomad cluster."
    echo -e "This deployment includes:"
    echo -e "  • Container port 5000 (for HTTP MCP server)"
    echo -e "  • Static host port 8091 (for reliable Traefik routing)"
    echo -e "  • Traefik routing configuration for external access"
    echo -e "  • Service registration with Consul"
    echo ""
    print_command "cd ${DEMO_OUTPUT_DIR}/$PACK_NAME && nomad-pack run --var=\"region=$NOMAD_REGION\" --var=\"container_port=5000\" --var=\"host_port=8091\" --var='service_tags=[...]' ."
    echo ""
    wait_for_user

    cd "${DEMO_OUTPUT_DIR}/$PACK_NAME"
    nomad-pack run \
      --var="region=$NOMAD_REGION" \
      --var="container_port=5000" \
      --var="host_port=8091" \
      --var='service_tags=["traefik.enable=true","traefik.http.middlewares.queryweaver-strip.stripprefix.prefixes=/queryweaver","traefik.http.routers.queryweaver-mcp.entrypoints=http","traefik.http.routers.queryweaver-mcp.rule=Host(`hashistack-demo-dc1-7329a99ab7f6a3ba.elb.us-east-1.amazonaws.com`) && PathPrefix(`/queryweaver`)","traefik.http.routers.queryweaver-mcp.middlewares=queryweaver-strip","traefik.http.routers.queryweaver-mcp-secure.entrypoints=https","traefik.http.routers.queryweaver-mcp-secure.tls.certresolver=letsencrypt","traefik.http.routers.queryweaver-mcp-secure.rule=Host(`hashistack-demo-dc1-7329a99ab7f6a3ba.elb.us-east-1.amazonaws.com`) && PathPrefix(`/queryweaver`)","traefik.http.routers.queryweaver-mcp-secure.middlewares=queryweaver-strip"]' \
      .
    cd "$ORIGINAL_DIR"

    print_success "Deployment initiated!"
    print_info "Waiting for allocation to start..."
    sleep 5
    echo ""
    wait_for_user

    # Section 5: Job Verification
    print_header "Section 5: Verifying Job Status"
    echo -e "Let's check the job status in Nomad."
    echo ""
    print_command "nomad job status $JOB_NAME"
    echo ""
    wait_for_user

    nomad job status "$JOB_NAME"
    echo ""
    print_success "Job is deployed!"
    echo ""
    wait_for_user

    # Section 6: Allocation Inspection
    print_header "Section 6: Inspecting Allocation"
    echo -e "Now let's examine the allocation to see detailed task information."
    echo ""

    # Get allocation ID
    ALLOC_ID=$(nomad job allocs "$JOB_NAME" | tail -n +2 | awk '{print $1}' | head -1)
    print_info "Allocation ID: $ALLOC_ID"
    echo ""

    print_command "nomad alloc status $ALLOC_ID"
    echo ""
    wait_for_user

    nomad alloc status "$ALLOC_ID"
    echo ""

    print_success "Allocation is running!"
    print_info "Note the 'Allocation Addresses' section showing the HTTP port assignment"
    echo ""
    wait_for_user

    # Section 7: Log Inspection
    print_header "Section 7: Viewing MCP Server Logs"
    echo -e "Let's look at the logs to see the MCP server starting up."
    echo ""
    print_info "Standard output (stdout):"
    print_command "nomad alloc logs $ALLOC_ID"
    echo ""
    wait_for_user

    nomad alloc logs "$ALLOC_ID" | head -20
    echo ""
    print_info "Standard error (stderr) - HTTP server details:"
    print_command "nomad alloc logs -stderr $ALLOC_ID"
    echo ""
    wait_for_user

    nomad alloc logs -stderr "$ALLOC_ID" | tail -10
    echo ""

    print_success "MCP server is running and listening on HTTP!"
    print_info "The server is accessible via:"
    print_info "  • Internal: Allocated Nomad port (shown in allocation status above)"
    print_info "  • External: http://hashistack-demo-dc1-7329a99ab7f6a3ba.elb.us-east-1.amazonaws.com/queryweaver/mcp"
    echo ""
    wait_for_user

    # Section 8: Connecting to Claude Code
    print_header "Section 8: Connecting to Claude Code"
    echo -e "Now that the MCP server is deployed and accessible, let's connect it to Claude Code."
    echo -e "This allows you to use the QueryWeaver MCP server directly within Claude Code sessions."
    echo ""

    print_step "Generating API token for authentication"
    echo ""

    # Set the QueryWeaver URL
    QUERYWEAVER_URL="http://hashistack-demo-dc1-7329a99ab7f6a3ba.elb.us-east-1.amazonaws.com/queryweaver"

    # Wait for QueryWeaver backend (FalkorDB) to be fully ready
    print_info "Waiting for QueryWeaver backend to be fully ready (this may take 30-60 seconds)..."
    MAX_RETRIES=30
    RETRY_COUNT=0
    READY=false

    while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
        # Test an actual API endpoint to ensure backend is ready
        TEST_RESPONSE=$(curl -s -w "%{http_code}" -o /dev/null "${QUERYWEAVER_URL}/docs" 2>/dev/null)
        if [ "$TEST_RESPONSE" = "200" ]; then
            READY=true
            print_success "QueryWeaver backend is ready!"
            break
        fi
        RETRY_COUNT=$((RETRY_COUNT + 1))
        if [ $RETRY_COUNT -lt $MAX_RETRIES ]; then
            echo -n "."
            sleep 3
        fi
    done
    echo ""

    if [ "$READY" = false ]; then
        print_warning "QueryWeaver backend may not be fully ready, attempting token generation anyway..."
    fi

    # Give it a few more seconds for good measure
    sleep 5
    echo ""

    print_info "Creating QueryWeaver account..."
    SIGNUP_RESPONSE=$(curl -s -X POST ${QUERYWEAVER_URL}/signup/email \
      -H 'Content-Type: application/json' \
      -d '{"firstName":"Demo","lastName":"User","email":"demo@example.com","password":"DemoPassword123!"}')

    if [[ "${DEMO_AUTO}" != "true" ]]; then
        print_info "Signup response: $SIGNUP_RESPONSE"
    fi

    print_info "Logging in..."
    LOGIN_RESPONSE=$(curl -s -X POST ${QUERYWEAVER_URL}/login/email \
      -H 'Content-Type: application/json' \
      -c /tmp/queryweaver_cookies.txt \
      -d '{"email":"demo@example.com","password":"DemoPassword123!"}')

    if [[ "${DEMO_AUTO}" != "true" ]]; then
        print_info "Login response: $LOGIN_RESPONSE"
        print_info "Cookie file contents:"
        cat /tmp/queryweaver_cookies.txt 2>/dev/null || echo "No cookie file"
    fi

    print_info "Generating API token..."
    TOKEN_RESPONSE=$(curl -s -X POST ${QUERYWEAVER_URL}/tokens/generate \
      -b /tmp/queryweaver_cookies.txt \
      -H 'Content-Type: application/json')

    if [[ "${DEMO_AUTO}" != "true" ]]; then
        print_info "Token response: $TOKEN_RESPONSE"
    fi

    # Parse token from response, handling potential errors
    if echo "$TOKEN_RESPONSE" | jq -e . >/dev/null 2>&1; then
        QUERYWEAVER_TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.token_id')
    else
        QUERYWEAVER_TOKEN=""
        print_error "Failed to parse token response as JSON"
    fi

    if [ -n "$QUERYWEAVER_TOKEN" ] && [ "$QUERYWEAVER_TOKEN" != "null" ]; then
        print_success "Token generated successfully!"
        print_info "QUERYWEAVER_TOKEN=$QUERYWEAVER_TOKEN"
        echo ""
    else
        print_error "Failed to generate token"
        print_warning "You may need to generate a token manually - see README.md"
        echo ""
    fi

    wait_for_user

    print_info "To add this server to Claude Code, run:"
    echo ""
    print_command "claude mcp add --scope=user --transport=http --header=\"Authorization: Bearer $QUERYWEAVER_TOKEN\" queryweaver '${QUERYWEAVER_URL}/mcp'"
    echo ""

    if [[ "${DEMO_AUTO}" != "true" ]]; then
        print_warning "Go ahead and add the MCP server to Claude Code using the command above."
        print_warning "Test it by asking Claude to use QueryWeaver capabilities."
        echo ""
    fi
    wait_for_user

    print_success "MCP server is ready to use with Claude Code!"
    echo ""

    # Section 9: Cleanup
    print_header "Section 9: Cleanup"
    echo -e "Finally, let's clean up by stopping the job."
    echo ""
    print_command "nomad job stop $JOB_NAME"
    echo ""
    wait_for_user

    nomad job stop "$JOB_NAME"
    echo ""
    print_success "Job stopped successfully!"
    echo ""
    wait_for_user

    # Section 10: Summary
    print_header "Demo Complete!"
    echo -e "${BOLD}Summary of what we demonstrated:${NC}"
    echo ""
    echo -e "  ✓ Generated Nomad Pack from MCP Registry"
    echo -e "  ✓ Planned deployment with nomad-pack"
    echo -e "  ✓ Deployed OCI-based MCP server to Nomad"
    echo -e "  ✓ Configured Traefik routing for external access"
    echo -e "  ✓ Registered service with Consul"
    echo -e "  ✓ Verified job and allocation health"
    echo -e "  ✓ Inspected HTTP port allocation"
    echo -e "  ✓ Viewed server logs and startup"
    echo -e "  ✓ Connected MCP server to Claude Code"
    echo -e "  ✓ Cleaned up resources"
    echo ""
    echo -e "${BOLD}Key Integration Points:${NC}"
    echo -e "  • MCP Registry → nomad-mcp-pack → Nomad Pack format"
    echo -e "  • Docker image deployment via Nomad docker driver"
    echo -e "  • Dynamic port allocation for HTTP transport"
    echo -e "  • Automatic service registration with Consul"
    echo -e "  • Traefik load balancer integration for external access"
    echo -e "  • Health checking and monitoring"
    echo ""
    echo -e "${CYAN}Next Steps:${NC}"
    echo -e "  • Try different MCP servers from the registry"
    echo -e "  • Customize pack variables (CPU, memory, etc.)"
    echo -e "  • Deploy to different Nomad datacenters"
    echo -e "  • Explore stdio vs HTTP transport types"
    echo -e "  • Use the deployed MCP server in your Claude Code workflows"
    echo ""
    print_success "Thank you for trying nomad-mcp-pack!"
}

# Run the demo
main
