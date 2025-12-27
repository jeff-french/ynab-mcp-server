# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based MCP (Model Context Protocol) server for YNAB (You Need A Budget) that supports both stdio (local) and HTTP (remote) transport modes. The server implements YNAB API operations as MCP tools and compiles to a single distributable binary.

## Key Commands

### Development
```bash
# Run in stdio mode (local/Claude Desktop)
go run main.go --transport=stdio

# Run in HTTP mode (remote/hosted)
go run main.go --transport=http --port=8080

# Build binary
go build -o ynab-mcp-server

# Run tests
go test ./...

# Run specific package tests
go test ./internal/tools -v

# Run with race detector
go test -race ./...
```

### Building for Distribution
```bash
# Build for current platform
make build

# Cross-compile for all platforms
make build-all

# Build targets: linux/amd64, darwin/amd64, darwin/arm64, windows/amd64
```

### Releasing

The project uses GoReleaser and GitHub Actions for automated releases.

#### Creating a Release

1. Ensure all changes are committed and pushed to `main`
2. Create and push a new tag:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```
3. GitHub Actions will automatically:
   - Run tests across all platforms
   - Build binaries for Linux, macOS, and Windows (amd64 and arm64)
   - Create Docker images for Linux (amd64 and arm64)
   - Push Docker images to GitHub Container Registry
   - Generate changelog from commits
   - Create GitHub release with all artifacts
   - Update Homebrew tap (if HOMEBREW_TAP_TOKEN is configured)

#### Release Artifacts

Each release includes:
- Pre-built binaries for all platforms (tar.gz/zip)
- SHA256 checksums
- Docker images: `ghcr.io/jeff-french/ynab-mcp-server:VERSION`
- Docker multi-arch manifest
- Homebrew formula (optional)

#### Required GitHub Secrets

- `GITHUB_TOKEN` - Automatically provided by GitHub Actions
- `HOMEBREW_TAP_TOKEN` - (Optional) Personal access token for updating Homebrew tap
  - Create at: https://github.com/settings/tokens
  - Needs `repo` scope
  - Add to repository secrets at: Settings → Secrets and variables → Actions

#### Testing Locally

Test the release process without publishing:
```bash
# Install GoReleaser
go install github.com/goreleaser/goreleaser/v2@latest

# Run in snapshot mode (doesn't publish)
goreleaser release --snapshot --clean

# Check dist/ directory for artifacts
ls -lh dist/
```

### Docker
```bash
# Build image
docker build -t ynab-mcp-server .

# Run with docker-compose (includes example config)
docker-compose up

# Run standalone
docker run -p 8080:8080 -e YNAB_ACCESS_TOKEN=xxx ynab-mcp-server
```

## Architecture

### Dual Transport Design

The server supports two MCP transport modes that share the same tool implementations:

**stdio mode**: For local execution (Claude Desktop, CLI)
- Reads JSON-RPC from stdin, writes to stdout
- Default mode if no flag specified
- Used in Claude Desktop `mcpServers` config

**HTTP mode**: For remote deployment (cloud hosting)
- HTTP server with POST /mcp/v1/messages endpoint
- Streaming responses using chunked transfer encoding
- Optional authentication middleware (MCP_AUTH_TOKEN)
- Health check at /health

### Project Structure

```
internal/
├── server/      # Core MCP server logic (protocol handling)
├── transport/   # Transport layer implementations
│   ├── stdio/   # stdin/stdout transport
│   └── http/    # HTTP streaming transport
├── ynab/        # YNAB API client wrapper
│   ├── client.go    # HTTP client with auth
│   └── types.go     # YNAB API models
├── tools/       # MCP tool implementations
│   ├── budgets.go      # list_budgets, get_budget_details
│   ├── accounts.go     # list_accounts, get_account_details
│   ├── transactions.go # list/create/update transactions
│   ├── categories.go   # list_categories, get_category_details
│   └── payees.go       # list_payees
└── config/      # Configuration loading (env vars + config file)
```

### Authentication Flow

1. **YNAB API Authentication**: Required for all modes
   - Bearer token from https://app.ynab.com/settings/developer
   - Loaded from YNAB_ACCESS_TOKEN env var or ~/.config/ynab-mcp/config.json
   - Passed in Authorization header to YNAB API

2. **MCP Server Authentication**: Optional for HTTP mode
   - Bearer token for remote server access (MCP_AUTH_TOKEN)
   - Validated via middleware on /mcp/v1/messages endpoint
   - Not used in stdio mode (local trust model)

### Tool Implementation Pattern

Each MCP tool follows this structure:
```go
type Tool struct {
    Name        string
    Description string
    InputSchema map[string]interface{}
}

func (t *Tool) Execute(params map[string]interface{}, ynabClient *ynab.Client) (interface{}, error) {
    // 1. Validate params against schema
    // 2. Call YNAB API via client
    // 3. Transform response to MCP format
    // 4. Return structured JSON
}
```

Tools must work identically in both stdio and HTTP modes.

## Configuration

### Environment Variables
- `YNAB_ACCESS_TOKEN` (required): YNAB API personal access token
- `TRANSPORT_MODE` (optional): "stdio" or "http", default stdio
- `HTTP_PORT` (optional): Port for HTTP mode, default 8080
- `HTTP_HOST` (optional): Host binding for HTTP mode, default 0.0.0.0
- `MCP_AUTH_TOKEN` (optional): Authentication token for HTTP mode
- `CORS_ENABLED` (optional): Enable CORS headers in HTTP mode

### Config File
Falls back to `~/.config/ynab-mcp/config.json`:
```json
{
  "ynab_access_token": "your-token",
  "transport_mode": "stdio",
  "http_port": 8080
}
```

Config file is created with defaults on first run if it doesn't exist.

## YNAB API Integration

The OpenAPI spec is at https://api.ynab.com/papi/open_api_spec.yaml

### Key API Endpoints
- GET /v1/budgets - List all budgets
- GET /v1/budgets/{budget_id} - Get budget details
- GET /v1/budgets/{budget_id}/accounts - List accounts
- GET /v1/budgets/{budget_id}/transactions - List transactions
- POST /v1/budgets/{budget_id}/transactions - Create transaction
- PUT /v1/budgets/{budget_id}/transactions/{transaction_id} - Update transaction
- GET /v1/budgets/{budget_id}/categories - List categories
- GET /v1/budgets/{budget_id}/payees - List payees

All requests require `Authorization: Bearer <token>` header.

## MCP Protocol Notes

### stdio Transport
- Uses JSON-RPC 2.0 over stdin/stdout
- Each message is newline-delimited JSON
- Server reads continuously until stdin closes

### HTTP Transport
- POST to /mcp/v1/messages with JSON body
- Response uses chunked transfer encoding for streaming
- Each chunk is a JSON-RPC message
- Content-Type: application/json
- Supports SSE-style streaming for long-running operations

## Testing Strategy

1. Unit tests for each tool in `internal/tools/*_test.go`
2. Integration tests for YNAB client (requires test token or mocks)
3. Transport tests for stdio and HTTP modes
4. End-to-end tests using MCP client library

Mock YNAB API responses for reproducible tests.

## Deployment

### Local (Claude Desktop)
Add to `claude_desktop_config.json`:
```json
{
  "mcpServers": {
    "ynab": {
      "command": "/path/to/ynab-mcp-server",
      "args": ["--transport=stdio"],
      "env": {
        "YNAB_ACCESS_TOKEN": "your-token"
      }
    }
  }
}
```

### Remote (HTTP)
Deploy to fly.io, Railway, Render, or any cloud platform:
1. Build Docker image
2. Set YNAB_ACCESS_TOKEN and MCP_AUTH_TOKEN env vars
3. Expose port 8080
4. Configure MCP client with server URL and auth token

Example systemd service included in `ynab-mcp.service` for Linux servers.
