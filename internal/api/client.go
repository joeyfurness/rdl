package api

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	DefaultBaseURL = "https://api.real-debrid.com/rest/1.0"
	maxRetries     = 3
	retryBaseDelay = 1 * time.Second
)

// Client is an HTTP client for the Real-Debrid API.
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

// APIError represents an error response from the Real-Debrid API.
type APIError struct {
	Code       string `json:"error"`
	ErrorCode  int    `json:"error_code"`
	HTTPStatus int    `json:"-"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d (HTTP %d): %s", e.ErrorCode, e.HTTPStatus, e.Code)
}

// NewClient creates a new Real-Debrid API client with the given token.
func NewClient(token string) *Client {
	return &Client{
		BaseURL:    DefaultBaseURL,
		Token:      token,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// Get performs a GET request and decodes the JSON response into result.
func (c *Client) Get(path string, result interface{}) error {
	return c.doRequest(http.MethodGet, path, nil, result)
}

// Post performs a POST request with form-encoded data and decodes the JSON response into result.
func (c *Client) Post(path string, form url.Values, result interface{}) error {
	return c.doRequest(http.MethodPost, path, form, result)
}

// Delete performs a DELETE request.
func (c *Client) Delete(path string) error {
	return c.doRequest(http.MethodDelete, path, nil, nil)
}

func (c *Client) doRequest(method, path string, form url.Values, result interface{}) error {
	fullURL := c.BaseURL + path

	for attempt := 0; attempt <= maxRetries; attempt++ {
		var body io.Reader
		if form != nil {
			body = strings.NewReader(form.Encode())
		}

		req, err := http.NewRequest(method, fullURL, body)
		if err != nil {
			return fmt.Errorf("creating request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+c.Token)
		if form != nil {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			return fmt.Errorf("executing request: %w", err)
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			resp.Body.Close()
			if attempt < maxRetries {
				delay := time.Duration(math.Pow(2, float64(attempt))) * retryBaseDelay
				time.Sleep(delay)
				continue
			}
			return &APIError{
				Code:       "rate_limited",
				HTTPStatus: http.StatusTooManyRequests,
			}
		}

		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNoContent {
			return nil
		}

		if resp.StatusCode >= 400 {
			var apiErr APIError
			if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
				return fmt.Errorf("HTTP %d: failed to decode error response: %w", resp.StatusCode, err)
			}
			apiErr.HTTPStatus = resp.StatusCode
			return &apiErr
		}

		if result != nil {
			if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
				return fmt.Errorf("decoding response: %w", err)
			}
		}

		return nil
	}

	return fmt.Errorf("max retries exceeded")
}
