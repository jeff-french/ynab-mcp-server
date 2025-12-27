# YNAB MCP Server

[![Release](https://img.shields.io/github/v/release/jeff-french/ynab-mcp-server)](https://github.com/jeff-french/ynab-mcp-server/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/jeff-french/ynab-mcp-server)](https://go.dev)
[![License](https://img.shields.io/github/license/jeff-french/ynab-mcp-server)](LICENSE)
[![Docker Pulls](https://img.shields.io/badge/docker-ghcr.io-blue)](https://ghcr.io/jeff-french/ynab-mcp-server)

A [Model Context Protocol (MCP)](https://modelcontextprotocol.io) server for [YNAB (You Need A Budget)](https://www.ynab.com) that provides seamless integration with Claude and other MCP clients. Access your budget data, manage transactions, and analyze spending through natural language conversations.

## Features

- **Dual Transport Support**: stdio (local) and HTTP (remote) modes
- **Comprehensive YNAB Integration**: Budgets, accounts, transactions, categories, and payees
- **Production Ready**: Built-in retry logic, rate limiting, and error handling
- **Single Binary**: Cross-platform distributable with no dependencies
- **Secure**: Optional authentication for HTTP mode
- **Well-Tested**: Comprehensive test coverage

## Quick Start

### Prerequisites

- YNAB account with [Personal Access Token](https://app.ynab.com/settings/developer)
- Claude Desktop or MCP-compatible client

### Installation

Choose your preferred installation method:

**Homebrew (macOS/Linux):**
```bash
brew install jeff-french/tap/ynab-mcp-server
```

**Docker:**
```bash
docker pull ghcr.io/jeff-french/ynab-mcp-server:latest
```

**Download Binary:**

Download the pre-built binary for your platform from the [releases page](https://github.com/jeff-french/ynab-mcp-server/releases).

**Build from Source:**
```bash
git clone https://github.com/jeff-french/ynab-mcp-server.git
cd ynab-mcp-server
make build
```

### Get Your YNAB Token

1. Go to [YNAB Developer Settings](https://app.ynab.com/settings/developer)
2. Click "New Token"
3. Give it a name (e.g., "MCP Server")
4. Copy the generated token (you won't be able to see it again!)

## Usage

### Local Mode (Claude Desktop)

For local use with Claude Desktop, use stdio transport:

**1. Configure Claude Desktop**

Add to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "ynab": {
      "command": "/path/to/ynab-mcp-server",
      "args": ["serve", "--transport=stdio"],
      "env": {
        "YNAB_ACCESS_TOKEN": "your-token-here"
      }
    }
  }
}
```

**2. Restart Claude Desktop**

The YNAB tools will now be available in your conversations!

### Remote Mode (HTTP Server)

For remote deployment or cloud hosting:

**1. Set Environment Variable**

```bash
export YNAB_ACCESS_TOKEN="your-token-here"
export MCP_AUTH_TOKEN="your-auth-token" # Optional but recommended
```

**2. Start Server**

```bash
./ynab-mcp-server serve --transport=http --port=8080
```

**3. Configure MCP Client**

```json
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

## Available Tools

### Budget Operations

- **`list_budgets`**: List all accessible budgets
- **`get_budget_details`**: Get comprehensive budget information

### Account Operations

- **`list_accounts`**: List all accounts in a budget
- **`get_account_details`**: Get detailed account information

### Transaction Operations

- **`list_transactions`**: List transactions with optional filters
- **`get_transaction_details`**: Get detailed transaction information
- **`create_transaction`**: Create a new transaction
- **`update_transaction`**: Update an existing transaction

### Category Operations

- **`list_categories`**: List all category groups and categories
- **`get_category_details`**: Get detailed category information with goals

### Payee Operations

- **`list_payees`**: List all payees in a budget

## Example Conversations

Once configured, you can use natural language with Claude:

```
You: "Show me my YNAB budgets"
Claude: [Lists your budgets using list_budgets tool]

You: "What's my checking account balance?"
Claude: [Shows account details including balance]

You: "Add a transaction: $45.67 at Whole Foods yesterday for groceries"
Claude: [Creates transaction using create_transaction tool]

You: "Show me all uncategorized transactions"
Claude: [Lists uncategorized transactions]
```

## Configuration

The server supports multiple configuration methods (in order of precedence):

1. **CLI Flags**: `--transport=http --port=8080`
2. **Environment Variables**: `YNAB_ACCESS_TOKEN`, `MCP_AUTH_TOKEN`
3. **Config File**: `~/.config/ynab-mcp/config.json`
4. **Defaults**: stdio mode, port 8080

### Environment Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `YNAB_ACCESS_TOKEN` | Your YNAB personal access token | Yes | - |
| `MCP_AUTH_TOKEN` | Authentication token for HTTP mode | No | - |
| `YNAB_MCP_TRANSPORT_MODE` | Transport mode: `stdio` or `http` | No | `stdio` |
| `YNAB_MCP_HTTP_PORT` | Port for HTTP mode | No | `8080` |
| `YNAB_MCP_HTTP_HOST` | Host binding for HTTP mode | No | `0.0.0.0` |
| `YNAB_MCP_LOG_LEVEL` | Log level: `info` or `debug` | No | `info` |

### Config File Example

Create `~/.config/ynab-mcp/config.json`:

```json
{
  "ynab_access_token": "your-token-here",
  "transport_mode": "stdio",
  "http_port": 8080,
  "mcp_auth_token": "",
  "log_level": "info"
}
```

## Deployment

### Docker

**Build and Run:**

```bash
docker build -t ynab-mcp-server .
docker run -p 8080:8080 \
  -e YNAB_ACCESS_TOKEN=your-token \
  ynab-mcp-server
```

**Docker Compose:**

```bash
# Create .env file
echo "YNAB_ACCESS_TOKEN=your-token" > .env

# Start
docker-compose up -d
```

### Cloud Platforms

The server is designed for easy deployment to cloud platforms:

**fly.io:**
```bash
flyctl launch
flyctl secrets set YNAB_ACCESS_TOKEN=your-token
flyctl deploy
```

**Railway:**
```bash
railway init
railway variables set YNAB_ACCESS_TOKEN=your-token
railway up
```

**Render:**

1. Create new Web Service
2. Connect your repository
3. Set environment variable `YNAB_ACCESS_TOKEN`
4. Deploy!

### Linux Systemd Service

1. Copy binary to `/usr/local/bin/ynab-mcp-server`
2. Create user: `sudo useradd -r -s /bin/false ynab-mcp`
3. Copy `ynab-mcp.service` to `/etc/systemd/system/`
4. Edit service file to add your token
5. Enable and start:

```bash
sudo systemctl enable ynab-mcp.service
sudo systemctl start ynab-mcp.service
sudo systemctl status ynab-mcp.service
```

## Building from Source

### Requirements

- Go 1.23 or higher

### Build Commands

```bash
# Install dependencies
make deps

# Build for current platform
make build

# Cross-compile for all platforms
make build-all

# Run tests
make test

# Run in stdio mode (development)
make run-stdio

# Run in HTTP mode (development)
make run-http
```

## Development

### Project Structure

```
ynab-mcp-server/
├── cmd/           # CLI commands (cobra)
├── internal/
│   ├── config/    # Configuration management
│   ├── server/    # MCP server and transports
│   ├── tools/     # MCP tool implementations
│   └── ynab/      # YNAB API client
├── main.go        # Entry point
└── Makefile       # Build automation
```

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage
```

## Troubleshooting

### "Failed to load configuration: YNAB access token is required"

Set your YNAB token via environment variable or config file.

### "Rate limit exceeded"

The YNAB API has rate limits. The server automatically retries with exponential backoff, but if you hit limits frequently, reduce request frequency.

### "Unauthorized" on HTTP endpoint

Ensure your `MCP_AUTH_TOKEN` matches between server and client configuration.

### Claude Desktop doesn't show tools

1. Verify `claude_desktop_config.json` is in the correct location
2. Check that the binary path is absolute
3. Restart Claude Desktop completely
4. Check logs in `~/Library/Logs/Claude/` (macOS) or `%APPDATA%/Claude/logs/` (Windows)

## Security

- **Never commit your YNAB token** to version control
- **Use authentication** (`MCP_AUTH_TOKEN`) when running HTTP mode
- **Use HTTPS** for remote deployments (e.g., behind reverse proxy)
- **Limit network exposure** - bind to `127.0.0.1` for local-only HTTP mode

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

- [YNAB API](https://api.ynab.com) - Official YNAB API documentation
- [Model Context Protocol](https://modelcontextprotocol.io) - MCP specification
- [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) - Go MCP library

## Support

- **Issues**: [GitHub Issues](https://github.com/jeff-french/ynab-mcp-server/issues)
- **Discussions**: [GitHub Discussions](https://github.com/jeff-french/ynab-mcp-server/discussions)
- **YNAB API**: [YNAB API Documentation](https://api.ynab.com)
