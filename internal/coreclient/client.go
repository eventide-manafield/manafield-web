package coreclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Module struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Version     string  `json:"version"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New(baseURL string, httpClient *http.Client) (*Client, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil, errors.New("Core URL is empty")
	}

	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse Core URL: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("unsupported Core URL scheme %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return nil, errors.New("Core URL must include a host")
	}

	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{baseURL: baseURL, httpClient: httpClient}, nil
}

func (c *Client) ListModules(ctx context.Context) ([]Module, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/modules", nil)
	if err != nil {
		return nil, fmt.Errorf("create modules request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request Core registry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Core registry returned %s", resp.Status)
	}

	var modules []Module
	if err := json.NewDecoder(resp.Body).Decode(&modules); err != nil {
		return nil, fmt.Errorf("decode Core registry response: %w", err)
	}

	return modules, nil
}
