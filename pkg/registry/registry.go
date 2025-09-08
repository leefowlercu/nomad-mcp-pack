package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Masterminds/semver/v3"
	v0 "github.com/modelcontextprotocol/registry/pkg/api/v0"
	"github.com/modelcontextprotocol/registry/pkg/model"
)

const (
	defaultTimeout = 30 * time.Second
	maxRetries     = 3
	retryDelay     = 1 * time.Second
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) (*Client, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("registry base URL is required")
	}

	if _, err := url.Parse(baseURL); err != nil {
		return nil, fmt.Errorf("invalid registry base URL: %w", err)
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}, nil
}

type ListServersOptions struct {
	Cursor string
	Limit  int
}

func (c *Client) ListServers(ctx context.Context, opts *ListServersOptions) (*v0.ServerListResponse, error) {
	u, err := url.Parse(fmt.Sprintf("%s/v0/servers", c.baseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	q := u.Query()
	if opts != nil {
		if opts.Cursor != "" {
			q.Set("cursor", opts.Cursor)
		}
		if opts.Limit > 0 {
			if opts.Limit > 100 {
				opts.Limit = 100 // Max limit
			}
			q.Set("limit", strconv.Itoa(opts.Limit))
		}
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	var resp *http.Response
	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err = c.httpClient.Do(req)
		if err == nil && resp.StatusCode < 500 {
			break // Success or Client error (no retry needed)
		}

		if attempt < maxRetries-1 {
			if resp != nil {
				resp.Body.Close()
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryDelay * time.Duration(attempt+1)): // Exponential backoff
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute request after %d attempts: %w", maxRetries, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var result v0.ServerListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetServer retrieves detailed information about a specific server
func (c *Client) GetServer(ctx context.Context, serverID string) (*v0.ServerJSON, error) {
	if serverID == "" {
		return nil, fmt.Errorf("server ID is required")
	}

	u := fmt.Sprintf("%s/v0/servers/%s", c.baseURL, url.PathEscape(serverID))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	var resp *http.Response
	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err = c.httpClient.Do(req)
		if err == nil && resp.StatusCode < 500 {
			break // Success or Client error (no retry needed)
		}

		if attempt < maxRetries-1 {
			if resp != nil {
				resp.Body.Close()
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryDelay * time.Duration(attempt+1)): // Exponential backoff
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute request after %d attempts: %w", maxRetries, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("server not found: %s", serverID)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var result v0.ServerJSON
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetLatestActiveServer retrieves the latest non-deprecated, non-deleted version of a server
// by collecting all servers with the matching name and selecting the one with the
// highest semantic version number
func (c *Client) GetLatestActiveServer(ctx context.Context, serverName string) (*v0.ServerJSON, error) {
	opts := &ListServersOptions{
		Limit: 100, // Max limit to reduce pagination
	}

	// Collect all matching servers
	var matchingServers []v0.ServerJSON

	for {
		resp, err := c.ListServers(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list servers: %w", err)
		}

		// Collect servers with matching name that are not deprecated or deleted
		for _, server := range resp.Servers {
			if server.Name == serverName {
				// Only include active servers (not deprecated or deleted)
				if server.Status == "" || server.Status == model.StatusActive {
					matchingServers = append(matchingServers, server)
				}
			}
		}

		// Check if there are more pages
		if resp.Metadata == nil || resp.Metadata.NextCursor == "" {
			break
		}

		opts.Cursor = resp.Metadata.NextCursor
	}

	if len(matchingServers) == 0 {
		return nil, fmt.Errorf("no active servers found with name: %s", serverName)
	}

	// Find the server with the latest semantic version
	var latestServer *v0.ServerJSON
	var latestVersion *semver.Version

	for i := range matchingServers {
		server := &matchingServers[i]

		// Parse the version from version_detail
		versionStr := server.VersionDetail.Version
		if versionStr == "" {
			continue // Skip servers without version information
		}

		// Try to parse as semantic version
		version, err := semver.NewVersion(versionStr)
		if err != nil {
			// If not a valid semver, skip this server
			// Log this as debug info if needed
			continue
		}

		// Check if this is the latest version we've seen
		if latestVersion == nil || version.GreaterThan(latestVersion) {
			latestVersion = version
			latestServer = server
		}
	}

	if latestServer == nil {
		return nil, fmt.Errorf("no valid semantic version found for servers with name: %s", serverName)
	}

	return latestServer, nil
}

func (c *Client) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
}
