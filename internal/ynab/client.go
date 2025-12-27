package ynab

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const (
	baseURL        = "https://api.ynab.com/v1"
	requestTimeout = 30 * time.Second
	maxRetries     = 3
)

// Client is the YNAB API HTTP client
type Client struct {
	accessToken string
	httpClient  *http.Client
}

// NewClient creates a new YNAB API client
func NewClient(accessToken string) *Client {
	return &Client{
		accessToken: accessToken,
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
	}
}

// doRequest executes an HTTP request with retry logic and rate limit handling
func (c *Client) doRequest(method, path string, body interface{}, result interface{}) error {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			slog.Debug("Retrying request after backoff", "attempt", attempt, "backoff", backoff)
			time.Sleep(backoff)
		}

		// Prepare request body
		var bodyReader io.Reader
		if body != nil {
			jsonBody, err := json.Marshal(body)
			if err != nil {
				return fmt.Errorf("failed to marshal request body: %w", err)
			}
			bodyReader = bytes.NewReader(jsonBody)
		}

		// Create HTTP request
		url := baseURL + path
		req, err := http.NewRequest(method, url, bodyReader)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		// Add headers
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		req.Header.Set("Accept", "application/json")

		// Execute request
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			slog.Warn("HTTP request failed", "error", err, "attempt", attempt+1)
			continue
		}
		defer resp.Body.Close()

		// Read response body
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("failed to read response: %w", err)
			slog.Warn("Failed to read response body", "error", err)
			continue
		}

		// Handle rate limiting (429 Too Many Requests)
		if resp.StatusCode == http.StatusTooManyRequests {
			lastErr = fmt.Errorf("rate limit exceeded")
			slog.Warn("Rate limit exceeded, will retry", "attempt", attempt+1)
			continue
		}

		// Handle other HTTP errors
		if resp.StatusCode >= 400 {
			var apiErr APIErrorResponse
			if err := json.Unmarshal(respBody, &apiErr); err == nil && apiErr.Error.Detail != "" {
				return fmt.Errorf("YNAB API error (%d): %s", resp.StatusCode, apiErr.Error.Detail)
			}
			return fmt.Errorf("YNAB API error: status %d", resp.StatusCode)
		}

		// Parse successful response
		if result != nil {
			if err := json.Unmarshal(respBody, result); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
		}

		return nil
	}

	return fmt.Errorf("request failed after %d attempts: %w", maxRetries, lastErr)
}

// get performs a GET request
func (c *Client) get(path string, result interface{}) error {
	return c.doRequest("GET", path, nil, result)
}

// post performs a POST request
func (c *Client) post(path string, body interface{}, result interface{}) error {
	return c.doRequest("POST", path, body, result)
}

// put performs a PUT request
func (c *Client) put(path string, body interface{}, result interface{}) error {
	return c.doRequest("PUT", path, body, result)
}
