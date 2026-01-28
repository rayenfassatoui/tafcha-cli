// Package cli provides the HTTP client for the Tafcha CLI.
package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client is the HTTP client for interacting with the Tafcha API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// CreateResponse matches the API response for snippet creation.
type CreateResponse struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}

// APIError represents an error from the API.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse wraps an API error.
type ErrorResponse struct {
	Error APIError `json:"error"`
}

// NewClient creates a new Tafcha API client.
func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Create uploads content and returns the snippet URL.
func (c *Client) Create(content []byte, expiry string) (*CreateResponse, error) {
	// Build URL with optional expiry query parameter
	apiURL := c.baseURL
	if expiry != "" {
		apiURL = fmt.Sprintf("%s?expiry=%s", c.baseURL, url.QueryEscape(expiry))
	}

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "text/plain")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		var errResp ErrorResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Error.Message != "" {
			return nil, fmt.Errorf("API error (%s): %s", errResp.Error.Code, errResp.Error.Message)
		}
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result CreateResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &result, nil
}

// Get retrieves a snippet's content by ID.
func (c *Client) Get(id string) ([]byte, error) {
	apiURL := fmt.Sprintf("%s/%s", c.baseURL, id)

	resp, err := c.httpClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("snippet not found or expired")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}
