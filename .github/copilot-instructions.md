# GitHub Copilot Instructions

This file provides context and guidance for GitHub Copilot when working with this repository.

Read all primary instructions from [AGENTS.md](../AGENTS.md) file.

## Copilot-Specific Notes

### Troubleshooting

#### Common Issues

- **"YNAB access token is required"**: Set `YNAB_ACCESS_TOKEN` env var
- **"Rate limit exceeded"**: Reduce request frequency, client has retry logic
- **"Unauthorized" on HTTP**: Check `MCP_AUTH_TOKEN` matches client config
- **Build failures**: Ensure Go 1.23+ is installed

### Additional Resources

- [YNAB API Documentation](https://api.ynab.com)
- [MCP Specification](https://modelcontextprotocol.io)
- [Go Documentation](https://go.dev/doc/)
- [Project README](../README.md)
- [Claude-specific instructions](../CLAUDE.md)
