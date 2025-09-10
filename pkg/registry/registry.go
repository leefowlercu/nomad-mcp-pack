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
	Cursor      string
	Limit       int
	UpdatedSince string
	Search      string
	Version     string
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
				opts.Limit = 100
			}
			q.Set("limit", strconv.Itoa(opts.Limit))
		}
		if opts.UpdatedSince != "" {
			q.Set("updated_since", opts.UpdatedSince)
		}
		if opts.Search != "" {
			q.Set("search", opts.Search)
		}
		if opts.Version != "" {
			q.Set("version", opts.Version)
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
			break
		}

		if attempt < maxRetries-1 {
			if resp != nil {
				resp.Body.Close()
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryDelay * time.Duration(attempt+1)):
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
			break
		}

		if attempt < maxRetries-1 {
			if resp != nil {
				resp.Body.Close()
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryDelay * time.Duration(attempt+1)):
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

// GetLatestActiveServer retrieves the latest active version of a server
// by collecting all servers with the matching name and selecting the one with the
// highest semantic version number that has Status == model.StatusActive
func (c *Client) GetLatestActiveServer(ctx context.Context, serverName string) (*v0.ServerJSON, error) {
	if serverName == "" {
		return nil, fmt.Errorf("server name is required")
	}

	opts := &ListServersOptions{
		Search: serverName,
		Limit:  100,
	}

	// Collect all matching active servers
	var matchingServers []v0.ServerJSON

	for {
		resp, err := c.ListServers(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list servers: %w", err)
		}

		for _, server := range resp.Servers {
			if server.Name == serverName {
				// Only include servers with explicit active status
				if server.Status == model.StatusActive {
					matchingServers = append(matchingServers, server)
				}
			}
		}

		if resp.Metadata == nil || resp.Metadata.NextCursor == "" {
			break
		}

		opts.Cursor = resp.Metadata.NextCursor
	}

	if len(matchingServers) == 0 {
		return nil, fmt.Errorf("no active servers found with name: %s", serverName)
	}

	var latestServer *v0.ServerJSON
	var latestVersion *semver.Version

	for i := range matchingServers {
		server := &matchingServers[i]

		versionStr := server.Version
		if versionStr == "" {
			continue // Skip servers without version information
		}

		version, err := semver.NewVersion(versionStr)
		if err != nil {
			// If not a valid semver, skip this server
			continue
		}
		if latestVersion == nil || version.GreaterThan(latestVersion) {
			latestVersion = version
			latestServer = server
		}
	}

	if latestServer == nil {
		return nil, fmt.Errorf("no valid semantic version found for active servers with name: %s", serverName)
	}

	return latestServer, nil
}

func (c *Client) SearchServers(ctx context.Context, searchTerm string, opts *ListServersOptions) (*v0.ServerListResponse, error) {
	if searchTerm == "" {
		return nil, fmt.Errorf("search term is required")
	}

	searchOpts := &ListServersOptions{
		Search: searchTerm,
	}

	if opts != nil {
		searchOpts.Cursor = opts.Cursor
		searchOpts.Limit = opts.Limit
		searchOpts.UpdatedSince = opts.UpdatedSince
		searchOpts.Version = opts.Version
		// Don't override the Search field
	}

	return c.ListServers(ctx, searchOpts)
}

func (c *Client) GetLatestServers(ctx context.Context, opts *ListServersOptions) (*v0.ServerListResponse, error) {
	latestOpts := &ListServersOptions{
		Version: "latest",
	}

	if opts != nil {
		latestOpts.Cursor = opts.Cursor
		latestOpts.Limit = opts.Limit
		latestOpts.UpdatedSince = opts.UpdatedSince
		latestOpts.Search = opts.Search
		// Don't override the Version field
	}

	return c.ListServers(ctx, latestOpts)
}

func (c *Client) GetUpdatedServers(ctx context.Context, updatedSince string, opts *ListServersOptions) (*v0.ServerListResponse, error) {
	if updatedSince == "" {
		return nil, fmt.Errorf("updated_since timestamp is required")
	}

	updatedOpts := &ListServersOptions{
		UpdatedSince: updatedSince,
	}

	if opts != nil {
		updatedOpts.Cursor = opts.Cursor
		updatedOpts.Limit = opts.Limit
		updatedOpts.Search = opts.Search
		updatedOpts.Version = opts.Version
		// Don't override the UpdatedSince field
	}

	return c.ListServers(ctx, updatedOpts)
}

func (c *Client) GetServerByNameAndVersion(ctx context.Context, serverName, version string) (*v0.ServerJSON, error) {
	if serverName == "" {
		return nil, fmt.Errorf("server name is required")
	}
	if version == "" {
		return nil, fmt.Errorf("version is required")
	}

	opts := &ListServersOptions{
		Search: serverName,
		Limit:  100,
	}
	
	// Only use version filter if it's "latest" - the API doesn't support filtering by specific versions
	if version == "latest" {
		opts.Version = "latest"
	}

	for {
		resp, err := c.ListServers(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to search servers: %w", err)
		}

		for _, server := range resp.Servers {
			if server.Name == serverName {
				// If we're looking for "latest" or the version matches exactly, return it
				if version == "latest" || server.Version == version {
					return &server, nil
				}
			}
		}

		if resp.Metadata == nil || resp.Metadata.NextCursor == "" {
			break
		}

		opts.Cursor = resp.Metadata.NextCursor
	}

	return nil, fmt.Errorf("server not found: %s@%s", serverName, version)
}

func (c *Client) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
}
