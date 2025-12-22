package cmd

import (
	"log"
	"log/slog"
	"os"

	"github.com/jeff-french/ynab-mcp-server/internal/config"
	"github.com/jeff-french/ynab-mcp-server/internal/server"
	"github.com/jeff-french/ynab-mcp-server/internal/ynab"
	"github.com/spf13/cobra"
)

var (
	transport  string
	port       int
	configPath string
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the MCP server",
	Long: `Start the YNAB MCP server in either stdio or HTTP mode.

stdio mode: Reads JSON-RPC from stdin, writes to stdout (for Claude Desktop)
http mode: Runs HTTP server with /mcp/v1/messages endpoint (for remote access)`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load configuration
		cfg, err := config.Load(configPath)
		if err != nil {
			log.Fatalf("Failed to load configuration: %v", err)
		}

		// Override transport from flag if specified
		if cmd.Flags().Changed("transport") {
			cfg.TransportMode = transport
		}
		if cmd.Flags().Changed("port") {
			cfg.HTTPPort = port
		}

		// Setup logging (write to stderr for stdio compatibility)
		logLevel := slog.LevelInfo
		if cfg.LogLevel == "debug" {
			logLevel = slog.LevelDebug
		}
		logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: logLevel,
		}))
		slog.SetDefault(logger)

		// Create YNAB client
		ynabClient := ynab.NewClient(cfg.YNABToken)

		// Create MCP server
		mcpServer, err := server.NewMCPServer(ynabClient)
		if err != nil {
			log.Fatalf("Failed to create MCP server: %v", err)
		}

		// Run appropriate transport
		switch cfg.TransportMode {
		case "stdio":
			slog.Info("Starting YNAB MCP server in stdio mode")
			if err := server.ServeStdio(mcpServer); err != nil {
				log.Fatalf("stdio server error: %v", err)
			}
		case "http":
			slog.Info("Starting YNAB MCP server in HTTP mode", "port", cfg.HTTPPort)
			if err := server.ServeHTTP(mcpServer, cfg.HTTPPort, cfg.MCPAuthToken); err != nil {
				log.Fatalf("HTTP server error: %v", err)
			}
		default:
			log.Fatalf("Invalid transport mode: %s (must be 'stdio' or 'http')", cfg.TransportMode)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringVarP(&transport, "transport", "t", "stdio", "Transport mode: stdio or http")
	serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "HTTP port (http mode only)")
	serveCmd.Flags().StringVarP(&configPath, "config", "c", "", "Config file path")
}
