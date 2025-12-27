# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

@AGENTS.md

## Claude-Specific Instructions

### Key Commands

#### Development
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

#### Building for Distribution
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
