// Copyright (C) 2025 blubskye
  //
  // This program is free software: you can redistribute it and/or modify
  // it under the terms of the GNU Affero General Public License as published by
  // the Free Software Foundation, either version 3 of the License, or
  // (at your option) any later version.
  //
  // This program is distributed in the hope that it will be useful,
  // but WITHOUT ANY WARRANTY; without even the implied warranty of
  // MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
  // GNU Affero General Public License for more details.
  //
  // You should have received a copy of the GNU Affero General Public License
  // along with this program.  If not, see <https://www.gnu.org/licenses/>.
  //
  // Source code: https://github.com/blubskye/godiscordmobileclient

  
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	BaseURL    = "https://discord.com/api/v10"
	CDNBaseURL = "https://cdn.discordapp.com"
)

// Client is a Discord REST API client
type Client struct {
	token      string
	httpClient *http.Client
	userAgent  string

	// Rate limiting
	mu          sync.Mutex
	globalReset time.Time
	buckets     map[string]*bucket
}

type bucket struct {
	remaining int
	reset     time.Time
	mu        sync.Mutex
}

// NewClient creates a new API client
func NewClient(token string) *Client {
	return &Client{
		token: token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		userAgent: "DiscordMobile (https://github.com/blubskye/godiscordmobileclient, 1.0.0)",
		buckets:   make(map[string]*bucket),
	}
}

// SetToken updates the authentication token
func (c *Client) SetToken(token string) {
	c.mu.Lock()
	c.token = token
	c.mu.Unlock()
}

// Request makes an authenticated request to the Discord API
func (c *Client) Request(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	return c.doRequest(ctx, method, path, body, path)
}

// RequestWithBucket makes a request using a specific rate limit bucket
func (c *Client) RequestWithBucket(ctx context.Context, method, path, bucket string, body interface{}) ([]byte, error) {
	return c.doRequest(ctx, method, path, body, bucket)
}

func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, bucketKey string) ([]byte, error) {
	// Check global rate limit
	c.mu.Lock()
	if time.Now().Before(c.globalReset) {
		wait := time.Until(c.globalReset)
		c.mu.Unlock()
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		c.mu.Lock()
	}
	c.mu.Unlock()

	// Check bucket rate limit
	b := c.getBucket(bucketKey)
	b.mu.Lock()
	if b.remaining == 0 && time.Now().Before(b.reset) {
		wait := time.Until(b.reset)
		b.mu.Unlock()
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		b.mu.Lock()
	}
	b.mu.Unlock()

	// Build request
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, BaseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	c.mu.Lock()
	token := c.token
	c.mu.Unlock()

	req.Header.Set("Authorization", token)
	req.Header.Set("User-Agent", c.userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Update rate limits from headers
	c.updateRateLimits(resp, b)

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle errors
	if resp.StatusCode >= 400 {
		return nil, c.handleError(resp.StatusCode, respBody)
	}

	return respBody, nil
}

func (c *Client) getBucket(key string) *bucket {
	c.mu.Lock()
	defer c.mu.Unlock()

	if b, ok := c.buckets[key]; ok {
		return b
	}

	b := &bucket{remaining: 1}
	c.buckets[key] = b
	return b
}

func (c *Client) updateRateLimits(resp *http.Response, b *bucket) {
	// Global rate limit
	if resp.Header.Get("X-RateLimit-Global") == "true" {
		if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
			if seconds, err := strconv.ParseFloat(retryAfter, 64); err == nil {
				c.mu.Lock()
				c.globalReset = time.Now().Add(time.Duration(seconds*1000) * time.Millisecond)
				c.mu.Unlock()
			}
		}
	}

	// Bucket rate limit
	b.mu.Lock()
	defer b.mu.Unlock()

	if remaining := resp.Header.Get("X-RateLimit-Remaining"); remaining != "" {
		if n, err := strconv.Atoi(remaining); err == nil {
			b.remaining = n
		}
	}

	if resetAfter := resp.Header.Get("X-RateLimit-Reset-After"); resetAfter != "" {
		if seconds, err := strconv.ParseFloat(resetAfter, 64); err == nil {
			b.reset = time.Now().Add(time.Duration(seconds*1000) * time.Millisecond)
		}
	}
}

// APIError represents a Discord API error
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("discord api error %d: %s (status %d)", e.Code, e.Message, e.Status)
}

func (c *Client) handleError(status int, body []byte) error {
	var apiErr APIError
	if err := json.Unmarshal(body, &apiErr); err != nil {
		return &APIError{
			Status:  status,
			Message: string(body),
		}
	}
	apiErr.Status = status
	return &apiErr
}

// GET is a convenience method for GET requests
func (c *Client) GET(ctx context.Context, path string) ([]byte, error) {
	return c.Request(ctx, http.MethodGet, path, nil)
}

// POST is a convenience method for POST requests
func (c *Client) POST(ctx context.Context, path string, body interface{}) ([]byte, error) {
	return c.Request(ctx, http.MethodPost, path, body)
}

// PATCH is a convenience method for PATCH requests
func (c *Client) PATCH(ctx context.Context, path string, body interface{}) ([]byte, error) {
	return c.Request(ctx, http.MethodPatch, path, body)
}

// PUT is a convenience method for PUT requests
func (c *Client) PUT(ctx context.Context, path string, body interface{}) ([]byte, error) {
	return c.Request(ctx, http.MethodPut, path, body)
}

// DELETE is a convenience method for DELETE requests
func (c *Client) DELETE(ctx context.Context, path string) ([]byte, error) {
	return c.Request(ctx, http.MethodDelete, path, nil)
}
