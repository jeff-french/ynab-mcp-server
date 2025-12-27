# GitHub Copilot Instructions

This file provides context and guidance for GitHub Copilot when working with this repository.

## Project Overview

This is a Go-based MCP (Model Context Protocol) server for YNAB (You Need A Budget) that enables natural language interaction with YNAB budgets through Claude and other MCP-compatible clients. The server supports both stdio (local) and HTTP (remote) transport modes and compiles to a single distributable binary.

**Key Features:**
- Dual transport support (stdio for local, HTTP for remote)
- Comprehensive YNAB API integration (budgets, accounts, transactions, categories, payees)
- Production-ready with retry logic, rate limiting, and error handling
- Cross-platform single binary with no dependencies
- Optional authentication for HTTP mode

## Architecture

### Transport Design

The server implements two MCP transport modes sharing the same tool implementations:

**stdio mode**: For local execution (Claude Desktop, CLI)
- JSON-RPC 2.0 over stdin/stdout
- Newline-delimited JSON messages
- Default mode, used in Claude Desktop configuration

**HTTP mode**: For remote deployment (cloud hosting)
- HTTP POST endpoint at `/mcp/v1/messages`
- Streaming responses using chunked transfer encoding
- Optional authentication via `MCP_AUTH_TOKEN`
- Health check at `/health`

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
cmd/             # CLI commands using cobra
```

## Coding Conventions

### Go Best Practices

1. **Error Handling**: Always return errors, don't panic
   - Use wrapped errors for context: `fmt.Errorf("failed to %s: %w", action, err)`
   - Handle errors at the appropriate level

2. **Naming Conventions**:
   - Use camelCase for unexported identifiers
   - Use PascalCase for exported identifiers
   - Use descriptive names (e.g., `budgetID` not `bid`)

3. **Code Organization**:
   - Keep files focused on single responsibilities
   - Group related functionality in packages
   - Use interfaces for testability

4. **Testing**:
   - Write table-driven tests where applicable
   - Use meaningful test names: `TestFunctionName_Scenario_ExpectedBehavior`
   - Mock external dependencies (YNAB API calls)
   - Tests should be in `*_test.go` files

### MCP Tool Implementation Pattern

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

**Important**: Tools must work identically in both stdio and HTTP modes.

### Authentication

1. **YNAB API Authentication**: Required for all modes
   - Bearer token from https://app.ynab.com/settings/developer
   - Loaded from `YNAB_ACCESS_TOKEN` env var or `~/.config/ynab-mcp/config.json`
   - Passed in Authorization header to YNAB API

2. **MCP Server Authentication**: Optional for HTTP mode only
   - Bearer token for remote server access (`MCP_AUTH_TOKEN`)
   - Validated via middleware on `/mcp/v1/messages` endpoint
   - Not used in stdio mode (local trust model)

## Development Workflow

### Building and Testing

```bash
# Build for current platform
make build

# Run tests with race detector
make test

# Run tests with coverage
make test-coverage

# Format code
make fmt

# Lint code (requires golangci-lint)
make lint
```

### Running the Server

```bash
# stdio mode (local/development)
make run-stdio
# or
go run . serve --transport=stdio

# HTTP mode (remote/deployment)
make run-http
# or
go run . serve --transport=http --port=8080
```

### Testing Locally

When developing, you can test changes by:
1. Building: `make build`
2. Updating Claude Desktop config to point to your local binary
3. Restarting Claude Desktop
4. Testing tools through conversations

## YNAB API Integration

**API Base URL**: `https://api.ynab.com/v1`

**OpenAPI Spec**: https://api.ynab.com/papi/open_api_spec.yaml

### Key Endpoints

- `GET /budgets` - List all budgets
- `GET /budgets/{budget_id}` - Get budget details
- `GET /budgets/{budget_id}/accounts` - List accounts
- `GET /budgets/{budget_id}/transactions` - List transactions
- `POST /budgets/{budget_id}/transactions` - Create transaction
- `PUT /budgets/{budget_id}/transactions/{transaction_id}` - Update transaction
- `GET /budgets/{budget_id}/categories` - List categories
- `GET /budgets/{budget_id}/payees` - List payees

**Authentication**: All requests require `Authorization: Bearer <token>` header

**Rate Limiting**: YNAB API has rate limits. The client implements automatic retry with exponential backoff.

### YNAB Data Models

- Amounts are in milliunits (e.g., $10.00 = 10000)
- Dates are in ISO 8601 format (YYYY-MM-DD)
- Budget IDs can be actual IDs or "last-used"
- Transaction amounts: negative = outflow, positive = inflow

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

## Deployment

### Docker

```bash
# Build image
docker build -t ynab-mcp-server .

# Run with docker-compose (includes example config)
docker-compose up

# Run standalone
docker run -p 8080:8080 -e YNAB_ACCESS_TOKEN=xxx ynab-mcp-server
```

### Cross-Platform Builds

```bash
# Build for all platforms
make build-all

# Creates binaries in dist/:
# - ynab-mcp-server-linux-amd64
# - ynab-mcp-server-darwin-amd64
# - ynab-mcp-server-darwin-arm64
# - ynab-mcp-server-windows-amd64.exe
```

### Release Process

The project uses GoReleaser and GitHub Actions for automated releases:

1. Tag a new version: `git tag -a v1.0.0 -m "Release v1.0.0"`
2. Push the tag: `git push origin v1.0.0`
3. GitHub Actions will automatically:
   - Run tests across all platforms
   - Build binaries for Linux, macOS, Windows (amd64 and arm64)
   - Create Docker images for Linux (amd64 and arm64)
   - Push images to GitHub Container Registry
   - Generate changelog and create GitHub release

## Common Tasks

### Adding a New MCP Tool

1. Create tool implementation in `internal/tools/`
2. Define tool schema (name, description, input schema)
3. Implement `Execute` method with YNAB API call
4. Add tool to server's tool registry
5. Write tests in `*_test.go` file
6. Update documentation

### Adding YNAB API Endpoint Support

1. Add types to `internal/ynab/types.go` if needed
2. Add client method to `internal/ynab/client.go`
3. Use the client method in tool implementation
4. Write tests with mocked responses

### Modifying Transport Layer

Transport implementations are in `internal/transport/`:
- `stdio/` - Standard input/output transport
- `http/` - HTTP streaming transport

Changes should maintain compatibility with existing MCP clients.

## Dependencies

- `github.com/mark3labs/mcp-go` - MCP protocol implementation
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management

Use `go mod tidy` to clean up dependencies after changes.

## Security Considerations

1. **Never commit secrets**: Use environment variables or config files
2. **Validate all inputs**: Especially in tool parameters
3. **Use authentication**: Required for HTTP mode in production
4. **HTTPS for remote**: Always use HTTPS in production deployments
5. **Rate limiting**: Respect YNAB API rate limits

## Testing Strategy

1. **Unit tests**: Test individual functions and methods
2. **Integration tests**: Test YNAB client with mocked API
3. **Transport tests**: Test stdio and HTTP modes
4. **End-to-end tests**: Test complete MCP workflows

Mock YNAB API responses for reproducible tests.

## Troubleshooting

### Common Issues

- **"YNAB access token is required"**: Set `YNAB_ACCESS_TOKEN` env var
- **"Rate limit exceeded"**: Reduce request frequency, client has retry logic
- **"Unauthorized" on HTTP**: Check `MCP_AUTH_TOKEN` matches client config
- **Build failures**: Ensure Go 1.23+ is installed

## Additional Resources

- [YNAB API Documentation](https://api.ynab.com)
- [MCP Specification](https://modelcontextprotocol.io)
- [Go Documentation](https://go.dev/doc/)
- [Project README](../README.md)
- [Claude-specific instructions](../CLAUDE.md)
