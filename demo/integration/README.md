# Integration Demo - nomad-mcp-pack

This demo showcases the complete end-to-end workflow of generating a Nomad Pack from an MCP server in the official registry and deploying it to a live Nomad cluster.

## Overview

The integration demo demonstrates:

1. **Pack Generation**: Generating a Nomad Pack from the MCP Registry
2. **Deployment Planning**: Using nomad-pack to preview deployment
3. **Live Deployment**: Deploying the MCP server to a Nomad cluster
4. **Job Verification**: Checking job status and health
5. **Allocation Inspection**: Examining allocation details and resource usage
6. **Network Configuration**: Viewing HTTP port allocation for HTTP-based MCP servers
7. **Log Analysis**: Inspecting server logs for troubleshooting
8. **Cleanup**: Gracefully stopping and removing deployments

## Demo Server

The demo uses **com.falkordb/QueryWeaver** as the example MCP server:

- **Package Type**: OCI (Docker container)
- **Transport Type**: streamable-http (HTTP-based communication)
- **Image**: `docker.io/falkordb/queryweaver:0.0.11`
- **Description**: FalkorDB-based MCP server providing graph database capabilities

This server was chosen because it:
- Demonstrates Docker-based deployment
- Showcases HTTP transport with dynamic port allocation
- Is actively maintained and reliable
- Provides a realistic use case (graph database querying)

## Prerequisites

### Required Environment Variables

```bash
export NOMAD_ADDR="http://your-nomad-cluster:4646"
export NOMAD_TOKEN="your-nomad-acl-token"
```

### Required Executables

1. **nomad-mcp-pack** - Build from this repository:
   ```bash
   make build
   ```

2. **nomad-pack** - Install from HashiCorp:
   ```bash
   # macOS
   brew install hashicorp/tap/nomad-pack

   # Linux/Other
   # See: https://github.com/hashicorp/nomad-pack#installation
   ```

3. **nomad** - Install Nomad CLI:
   ```bash
   # macOS
   brew install nomad

   # Linux/Other
   # See: https://www.nomadproject.io/downloads
   ```

### Nomad Cluster Requirements

- Nomad cluster must be accessible via NOMAD_ADDR
- Your NOMAD_TOKEN must have permissions to:
  - Submit jobs
  - Read job status
  - Read allocation information
  - Stop jobs
- At least one Nomad client node with:
  - Docker driver enabled
  - Sufficient resources (500 MHz CPU, 512 MiB memory)
  - Internet access to pull Docker images

### Network Requirements

- Internet connectivity to:
  - MCP Registry API (`registry.modelcontextprotocol.io`)
  - Docker Hub (`docker.io`) for image pulls

## Usage

### Interactive Mode (Default)

Run the demo with pauses between sections for presentation or learning:

```bash
cd demo/integration
./demo.sh
```

Press Enter at each pause to continue to the next section.

### Automatic Mode

Run the demo without pauses for CI/CD or testing:

```bash
cd demo/integration
./demo.sh auto
```

## What to Expect

### Timing

- **Interactive Mode**: ~5-10 minutes (depending on pause lengths)
- **Automatic Mode**: ~2-3 minutes

### Output

The demo provides color-coded output:

- 🔵 **Blue**: Section headers and important information
- 🟢 **Green**: Success messages and step descriptions
- 🟡 **Yellow**: Warnings and pause prompts
- 🔴 **Red**: Errors (if any)
- 🟣 **Purple**: Commands being executed
- 🔷 **Cyan**: Informational messages

### Demo Sections

1. **Prerequisites Check**: Validates environment and connectivity
2. **Pack Generation**: Creates Nomad Pack from MCP Registry
3. **Pack Inspection**: Shows generated files and configuration
4. **Deployment Planning**: Runs nomad-pack plan (dry-run)
5. **Deployment Execution**: Deploys to Nomad cluster
6. **Job Verification**: Checks job status
7. **Allocation Inspection**: Examines allocation details and ports
8. **Log Inspection**: Views MCP server startup logs
9. **Cleanup**: Stops job and removes artifacts
10. **Summary**: Recaps what was demonstrated

### Expected Outputs

**Job Status Example:**
```
ID            = com-falkordb-QueryWeaver-0-0-11-oci-streamable-http
Status        = running
Type          = service
Datacenters   = dc1
```

**Allocation Addresses Example:**
```
Allocation Addresses:
Label  Dynamic  Address
*http  yes      10.0.2.199:25102
```

**Log Example:**
```
INFO:     Uvicorn running on http://0.0.0.0:5000 (Press CTRL+C to quit)
INFO:     Application startup complete.
```

## Troubleshooting

### Common Issues

#### 1. "No path to region" Error

**Symptom**: nomad-pack plan fails with "No path to region"

**Cause**: The region specified doesn't match your Nomad cluster configuration

**Solution**:
```bash
# Check your cluster's region
nomad server members

# The demo auto-detects region, but you can override:
# Edit the NOMAD_REGION variable in demo.sh
```

#### 2. Missing Environment Variables

**Symptom**: "NOMAD_ADDR environment variable is not set"

**Solution**:
```bash
export NOMAD_ADDR="http://your-nomad-cluster:4646"
export NOMAD_TOKEN="your-token"
```

#### 3. Image Pull Failures

**Symptom**: Allocation fails with "Failed to pull image"

**Possible Causes**:
- Nomad client nodes can't reach Docker Hub
- Rate limiting from Docker Hub
- Network connectivity issues

**Solution**:
- Check Docker daemon is running on Nomad clients
- Verify internet connectivity from clients
- Consider using Docker Hub authentication for rate limits

#### 4. Port Allocation Issues

**Symptom**: No port allocated or port conflicts

**Cause**: Network block not generated or ports exhausted

**Solution**:
- Verify pack template has network block for HTTP transport
- Check available ports on Nomad clients
- Review client configuration for port ranges

#### 5. Pack Already Exists

**Symptom**: Generation fails because pack directory exists

**Solution**:
The demo cleans up automatically, but if needed:
```bash
rm -rf packs/com-falkordb-QueryWeaver-*
```

## Advanced Usage

### Customizing Pack Variables

You can modify the demo to pass custom variables:

```bash
# In demo.sh, change the nomad-pack run command:
nomad-pack run \
  --var="region=$NOMAD_REGION" \
  --var="cpu=1000" \
  --var="memory=1024" \
  --var="count=2" \
  .
```

### Customizing Service Registration

Generated packs for HTTP-based MCP servers automatically include service registration with Consul. You can customize the service configuration:

```bash
# Custom service name
nomad-pack run --var="service_name=my-mcp-server" .

# Custom container port
nomad-pack run --var="container_port=5000" .

# Static host port (recommended for load balancer integration)
# Use host_port for static port allocation, or 0 for dynamic (default)
nomad-pack run \
  --var="container_port=5000" \
  --var="host_port=8091" \
  .

# Add Traefik routing tags for external access
nomad-pack run \
  --var='service_tags=["traefik.enable=true","traefik.http.middlewares.myserver-strip.stripprefix.prefixes=/myserver","traefik.http.routers.myserver.entrypoints=http","traefik.http.routers.myserver.rule=Host(`example.com`) && PathPrefix(`/myserver`)","traefik.http.routers.myserver.middlewares=myserver-strip"]' \
  .

# Custom health check intervals
nomad-pack run \
  --var="health_check_interval=60s" \
  --var="health_check_timeout=10s" \
  .
```

**Port Allocation Modes:**

- **Dynamic ports (default)**: Set `host_port=0` or omit the variable. Nomad allocates a random available port. Requires load balancers to use Consul service discovery for port resolution.
- **Static ports**: Set `host_port` to a specific port number (e.g., `8091`). The MCP server will always bind to that host port. Simpler for load balancer integration when service discovery is not configured.

See the generated pack's README for more details on service registration options.

### Testing Different MCP Servers

To try different servers, modify `DEMO_SERVER` in demo.sh:

```bash
# Other OCI + HTTP servers found in the registry:
DEMO_SERVER="io.github.eat-pray-ai/yutu"           # Transport: streamable-http
DEMO_SERVER="io.github.andrasfe/vulnicheck"       # Transport: streamable-http
DEMO_SERVER="io.github.jgador/websharp"           # Transport: streamable-http
```

### Deploying to Different Datacenters

```bash
# Modify the nomad-pack commands:
nomad-pack plan --var="region=us-west" --var="datacenters=[\"dc2\",\"dc3\"]" .
```

## Integration Points Demonstrated

### 1. MCP Registry → nomad-mcp-pack

- Queries MCP Registry API for server metadata
- Resolves `@latest` versions automatically
- Validates server status (active vs deprecated)
- Extracts package and transport configuration

### 2. nomad-mcp-pack → Nomad Pack Format

- Generates HCL templates compatible with nomad-pack
- Creates variables.hcl with configurable parameters
- Produces metadata.hcl with pack information
- Includes helper templates for reusable logic
- Automatically includes service registration for HTTP/SSE transport
- Infers sensible defaults for service names and container ports

### 3. Nomad Pack → Nomad Job

- nomad-pack renders templates with variable substitution
- Validates job specification before submission
- Submits job to Nomad API with proper authentication

### 4. Docker Driver Integration

- Nomad pulls Docker images from registry
- Configures container with environment variables
- Maps dynamic ports for HTTP transport
- Monitors container health and logs

### 5. Network Configuration & Service Registration

- Dynamic port allocation for HTTP-based servers
- Automatic container port mapping (configurable via variables)
- Port mapping visible in allocation addresses
- Automatic Consul service registration with TCP health checks
- Customizable service tags for load balancer integration (Traefik, Fabio, etc.)
- Configurable health check intervals and timeouts

## Key Differences: HTTP vs stdio Transport

### stdio Transport
- No network port required
- Process communication via stdin/stdout
- Simpler for local/embedded use cases
- Example: Most NPM-based MCP servers

### HTTP Transport (streamable-http)
- Requires network port allocation
- Client-server communication over HTTP
- Better for distributed deployments
- Example: This demo (QueryWeaver)
- Nomad allocates dynamic port automatically
- Port visible in allocation addresses

## Connecting to the Deployed MCP Server

After successfully deploying QueryWeaver to your Nomad cluster, you can connect Claude Code (or other MCP clients) to use the server.

### Step 1: Create a QueryWeaver Account

The QueryWeaver MCP server requires authentication. Create an account using the API:

```bash
# Replace with your Traefik/load balancer URL
QUERYWEAVER_URL="http://your-nomad-cluster/queryweaver"

# Create account
curl -X POST ${QUERYWEAVER_URL}/signup/email \
  -H "Content-Type: application/json" \
  -d '{
    "firstName": "Your",
    "lastName": "Name",
    "email": "your.email@example.com",
    "password": "YourSecurePassword123!"
  }'
```

### Step 2: Login and Generate API Token

After creating an account, login and generate an API token:

```bash
# Login to get session cookie
curl -X POST ${QUERYWEAVER_URL}/login/email \
  -H "Content-Type: application/json" \
  -c /tmp/queryweaver_cookies.txt \
  -d '{
    "email": "your.email@example.com",
    "password": "YourSecurePassword123!"
  }'

# Generate API token using session
curl -X POST ${QUERYWEAVER_URL}/tokens/generate \
  -b /tmp/queryweaver_cookies.txt \
  -H "Content-Type: application/json"
```

This will return a token in the format:
```json
{"token_id":"your-token-here","created_at":0}
```

### Step 3: Configure Claude Code

Create or update your Claude Code MCP configuration file at `~/.config/claude-code/mcp.json`:

```json
{
  "servers": {
    "queryweaver": {
      "type": "http",
      "url": "http://your-nomad-cluster/queryweaver/mcp",
      "headers": {
        "Authorization": "Bearer your-token-here"
      }
    }
  },
  "inputs": []
}
```

Replace:
- `http://your-nomad-cluster/queryweaver` with your actual Traefik/load balancer URL
- `your-token-here` with the token_id from step 2

### Step 4: Restart Claude Code

For Claude Code to pick up the new MCP server configuration:

```bash
# If running in a terminal session, restart claude
# The MCP server will be available in the next session
```

### Step 5: Test the Connection

Once Claude Code restarts, you can ask it to use QueryWeaver's capabilities:

- **list_databases** - Retrieve available databases
- **connect_database** - Establish database connections
- **database_schema** - Fetch schema information
- **query_database** - Execute Text2SQL queries

Example: "Use QueryWeaver to list available databases"

### Troubleshooting MCP Connection

#### MCP Server Not Responding

If Claude Code cannot connect to QueryWeaver:

1. **Verify the deployment is running**:
   ```bash
   nomad job status com-falkordb-QueryWeaver-0-0-11-oci-streamable-http
   ```

2. **Test the MCP endpoint manually**:
   ```bash
   curl -i -H 'Accept: text/event-stream' \
     -H 'Authorization: Bearer your-token' \
     'http://your-nomad-cluster/queryweaver/mcp'
   ```

   You should get a response indicating the MCP endpoint is active.

3. **Check Traefik routing**:
   ```bash
   nomad job status traefik  # or your load balancer job name
   ```

4. **Verify token is valid**:
   ```bash
   # List your tokens
   curl -X GET ${QUERYWEAVER_URL}/tokens/list \
     -b /tmp/queryweaver_cookies.txt
   ```

#### Token Expired or Invalid

If you get authentication errors, generate a new token:

```bash
# Login again
curl -X POST ${QUERYWEAVER_URL}/login/email \
  -H "Content-Type: application/json" \
  -c /tmp/queryweaver_cookies.txt \
  -d '{"email":"your.email@example.com","password":"YourPassword"}'

# Generate new token
curl -X POST ${QUERYWEAVER_URL}/tokens/generate \
  -b /tmp/queryweaver_cookies.txt
```

Update the token in `~/.config/claude-code/mcp.json` and restart Claude Code.

## Next Steps

After running this demo, you can:

1. **Connect and Use the MCP Server**:
   - Follow the steps above to configure Claude Code
   - Test QueryWeaver's database capabilities
   - Connect to FalkorDB instances for graph queries

2. **Explore Other Demos**:
   - `../generate/` - Pack generation features
   - `../watch/` - Continuous monitoring and generation

3. **Deploy Your Own MCP Servers**:
   - Browse the MCP Registry
   - Generate packs for servers you need
   - Customize variables for your environment

4. **Production Deployment**:
   - Add resource constraints
   - Configure health checks
   - Set up service discovery
   - Enable monitoring and alerting
   - Implement backup/restore procedures
   - Configure Traefik for external access

5. **Contribute**:
   - Report issues on GitHub
   - Submit pull requests
   - Share your MCP server deployments

## Resources

- [Nomad Documentation](https://www.nomadproject.io/docs)
- [Nomad Pack Documentation](https://github.com/hashicorp/nomad-pack)
- [MCP Registry](https://registry.modelcontextprotocol.io)
- [MCP Protocol Specification](https://modelcontextprotocol.io)
- [FalkorDB Documentation](https://www.falkordb.com/docs)

## License

This demo is part of the nomad-mcp-pack project and follows the same license.
