Create an MCP (Model Context Protocol) server for YNAB (You Need A Budget) using their OpenAPI specification. Build this in Go so it can be compiled and distributed as a single binary, with support for BOTH stdio (local) and HTTP transport (remote/hosted).

PROJECT REQUIREMENTS:
1. Build a Go-based MCP server that implements the YNAB API
2. Use the YNAB OpenAPI spec as the source of truth for API endpoints
3. Support BOTH transport modes:
   - stdio: For local execution (Claude Desktop, command line)
   - HTTP: For remote deployment using the streamable HTTP transport (cloud hosting)
4. Compile to a single, distributable binary
5. Include proper error handling and authentication

SETUP:
1. Initialize a new Go module (e.g., github.com/yourusername/ynab-mcp-server)
2. Fetch and parse the YNAB OpenAPI spec from: https://api.ynab.com/papi/open_api_spec.yaml
3. Use the official Go MCP SDK or implement both stdio and HTTP transports
4. Set up the project structure following Go conventions

DEPENDENCIES:
- MCP server library for Go (with stdio and HTTP transport support)
- HTTP client for YNAB API calls
- YAML parser for OpenAPI spec (if generating tools dynamically)
- Structured logging library
- HTTP router for MCP HTTP endpoints (e.g., chi, gorilla/mux, or stdlib)

IMPLEMENTATION:
1. Transport Layer:
   - Implement a flag/env var to choose transport mode: --transport=stdio or --transport=http
   - stdio mode: Read from stdin, write to stdout (for local use)
   - HTTP mode: Start HTTP server with MCP HTTP transport endpoints (for remote deployment)
   - Default to stdio if no flag provided

2. HTTP Transport Mode Specifics:
   - HTTP server with configurable port (default 8080)
   - POST /mcp/v1/messages endpoint for MCP protocol
   - Streaming responses using chunked transfer encoding
   - Health check endpoint at /health
   - Optional: CORS support for browser-based clients
   - Optional: Authentication middleware for remote access

3. Create MCP tools for key YNAB operations:
   - Budget operations (list_budgets, get_budget_details)
   - Account operations (list_accounts, get_account_details)
   - Transaction operations (list_transactions, create_transaction, update_transaction)
   - Category operations (list_categories, get_category_details)
   - Payee operations (list_payees)

4. Each tool should:
   - Have clear descriptions from the OpenAPI spec
   - Include proper input schemas with validation
   - Handle authentication via YNAB personal access token
   - Return well-formatted JSON responses
   - Include comprehensive error handling
   - Work identically in both stdio and HTTP modes

5. Authentication:
   - YNAB API token via:
     * Environment variable (YNAB_ACCESS_TOKEN)
     * Config file (~/.config/ynab-mcp/config.json)
   - Optional: MCP server authentication for HTTP mode
     * Bearer token or API key for remote access
     * Environment variable (MCP_AUTH_TOKEN)
     * Validate on each request to HTTP endpoints

6. Configuration:
   - Support both environment variables and config file
   - Create config directory and file on first run if needed
   - Configuration options:
     * YNAB_ACCESS_TOKEN (required)
     * TRANSPORT_MODE (stdio|http, default: stdio)
     * HTTP_PORT (default: 8080)
     * HTTP_HOST (default: 0.0.0.0)
     * MCP_AUTH_TOKEN (optional, for HTTP auth)
     * CORS_ENABLED (optional, for HTTP mode)

7. Build and Distribution:
   - Create Makefile or build script for cross-compilation
   - Target common platforms: linux/amd64, darwin/amd64, darwin/arm64, windows/amd64
   - Use Go's embed feature for any static resources if needed
   - Create Dockerfile for easy cloud deployment

DELIVERABLES:
- Working MCP server in Go with dual transport support
- go.mod and go.sum
- README.md with:
  - Installation instructions (including pre-built binary downloads)
  - How to get a YNAB API token from https://app.ynab.com/settings/developer
  - LOCAL USAGE: How to configure in Claude Desktop (stdio mode)
  - REMOTE USAGE: How to deploy and connect (HTTP mode)
    * Example deployment on common platforms (fly.io, railway, render, etc.)
    * How to configure MCP client to connect to remote HTTP server
    * Authentication setup for remote access
  - List of available tools with examples
  - Building from source instructions
- config.json.example template
- Dockerfile for containerized deployment
- docker-compose.yml example for easy local testing
- Makefile or build script for cross-compilation
- .gitignore for Go projects
- Example systemd service file for Linux deployment

PROJECT STRUCTURE SUGGESTION:

```
ynab-mcp-server/
├── main.go                    # Entry point, transport selection
├── go.mod
├── go.sum
├── README.md
├── Makefile
├── Dockerfile
├── docker-compose.yml
├── config.json.example
├── ynab-mcp.service          # systemd service example
├── internal/
│   ├── server/               # MCP server with stdio + HTTP
│   ├── transport/            # Transport implementations
│   ├── ynab/                 # YNAB API client
│   ├── tools/                # MCP tool implementations
│   └── config/               # Configuration handling
└── .gitignore
```
EXAMPLE USAGE PATTERNS TO DOCUMENT:

Local (stdio):
```bash
# Direct execution
./ynab-mcp-server --transport=stdio

# In Claude Desktop config
{
  "mcpServers": {
    "ynab": {
      "command": "/path/to/ynab-mcp-server",
      "args": ["--transport=stdio"],
      "env": {
        "YNAB_ACCESS_TOKEN": "your-token-here"
      }
    }
  }
}
```

Remote (HTTP):
```bash
# Start server
./ynab-mcp-server --transport=http --port=8080

# Or with Docker
docker run -p 8080:8080 -e YNAB_ACCESS_TOKEN=xxx ynab-mcp-server

# In Claude Desktop config (or Claude.ai MCP settings)
{
  "mcpServers": {
    "ynab": {
      "url": "https://your-server.com",
      "transport": "http",
      "headers": {
        "Authorization": "Bearer your-mcp-auth-token"
      }
    }
  }
}
```

Start by:
1. Setting up the Go module and project structure
2. Implementing both MCP transports (stdio and HTTP with streaming)
3. Creating the YNAB API client with authentication
4. Implementing 2-3 core tools to validate both transport modes work
5. Then expand to all major YNAB operations
6. Add deployment configurations and documentation

## COMMIT MESSAGE CONVENTION

This project uses [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) for all commit messages. This enables automated changelog generation and semantic versioning.

### Format
```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Types
- **feat**: New feature (minor version bump)
- **fix**: Bug fix (patch version bump)
- **docs**: Documentation changes
- **style**: Code formatting (no functional changes)
- **refactor**: Code restructuring (no functional changes)
- **perf**: Performance improvements
- **test**: Test additions/updates
- **build**: Build system/dependency changes
- **ci**: CI configuration changes
- **chore**: Maintenance tasks
- **revert**: Revert previous commit

### Breaking Changes
Use `!` after type or add `BREAKING CHANGE:` in footer for major version bumps.

### Examples
```
feat(transactions): add bulk import support
fix(auth): correct token refresh logic
docs(readme): update installation steps
test(ynab-client): add retry logic tests
```

**IMPORTANT**: Always use Conventional Commits format. The release process depends on this for changelogs and versioning.