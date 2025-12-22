package server

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/mark3labs/mcp-go/server"
)

// ServeHTTP starts the MCP server in HTTP mode with optional authentication
// This mode is used for remote deployment and cloud hosting
func ServeHTTP(mcpServer *server.MCPServer, port int, authToken string) error {
	// Create the streamable HTTP server (implements http.Handler)
	httpServer := server.NewStreamableHTTPServer(mcpServer)

	// Create custom mux with additional endpoints
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", healthCheckHandler)

	// Root handler for information
	mux.HandleFunc("/", rootHandler)

	// MCP endpoint - the streamable HTTP server implements http.Handler
	var mcpHandler http.Handler = httpServer

	// Apply auth middleware if token is provided
	if authToken != "" {
		slog.Info("HTTP authentication enabled")
		mcpHandler = authMiddleware(mcpHandler, authToken)
	} else {
		slog.Warn("HTTP authentication disabled - server is open to all requests")
	}

	// Mount at /mcp (the streamable HTTP server expects this path)
	mux.Handle("/mcp/", http.StripPrefix("/mcp", mcpHandler))

	addr := fmt.Sprintf(":%d", port)
	slog.Info("Starting HTTP server", "address", addr, "auth_enabled", authToken != "")

	return http.ListenAndServe(addr, mux)
}

// healthCheckHandler handles health check requests
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"ynab-mcp-server"}`))
}

// rootHandler provides basic server information
func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`YNAB MCP Server

This is a Model Context Protocol (MCP) server for YNAB (You Need A Budget).

Endpoints:
  POST /mcp/v1/messages - MCP protocol endpoint
  GET  /health          - Health check

For more information, visit: https://github.com/jeff-french/ynab-mcp-server
`))
}

// authMiddleware validates Bearer token authentication
func authMiddleware(next http.Handler, expectedToken string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		// Check for Bearer token
		expectedAuth := "Bearer " + expectedToken
		if authHeader != expectedAuth {
			slog.Warn("Unauthorized request", "remote_addr", r.RemoteAddr, "path", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"Unauthorized","message":"Valid Bearer token required"}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}
