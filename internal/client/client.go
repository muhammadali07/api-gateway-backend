package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"api-gateway-backend/internal/config"
)

// ExternalAPIClient handles external API requests
type ExternalAPIClient struct {
	client  *http.Client
	baseURL string
}

// PostResponse represents a post from JSONPlaceholder API
type PostResponse struct {
	UserID int    `json:"userId"`
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

// New creates a new external API client
func New(cfg config.ExternalAPIConfig) *ExternalAPIClient {
	return &ExternalAPIClient{
		client: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     30 * time.Second,
				DisableCompression:  true,
				MaxIdleConnsPerHost: 10,
			},
		},
		baseURL: cfg.BaseURL,
	}
}

// FetchPosts fetches posts from external API with retry logic
func (c *ExternalAPIClient) FetchPosts(ctx context.Context) ([]PostResponse, error) {
	url := fmt.Sprintf("%s/posts", c.baseURL)

	var posts []PostResponse
	err := c.retryRequest(ctx, url, &posts, 3)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch posts: %w", err)
	}

	return posts, nil
}

// retryRequest performs HTTP request with exponential backoff retry
func (c *ExternalAPIClient) retryRequest(ctx context.Context, url string, dest interface{}, maxRetries int) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s, 8s...
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "API-Gateway-Backend/1.0")

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		defer resp.Body.Close()

		// Check for successful status codes
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				lastErr = fmt.Errorf("failed to read response body: %w", err)
				continue
			}

			if err := json.Unmarshal(body, dest); err != nil {
				lastErr = fmt.Errorf("failed to unmarshal response: %w", err)
				continue
			}

			return nil
		}

		// Handle different HTTP error codes
		switch {
		case resp.StatusCode >= 500:
			// Server errors - retry
			lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
		case resp.StatusCode == 429:
			// Rate limited - retry with longer backoff
			lastErr = fmt.Errorf("rate limited: %d", resp.StatusCode)
		case resp.StatusCode >= 400:
			// Client errors - don't retry
			return fmt.Errorf("client error: %d", resp.StatusCode)
		default:
			lastErr = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
	}

	return fmt.Errorf("max retries exceeded, last error: %w", lastErr)
}