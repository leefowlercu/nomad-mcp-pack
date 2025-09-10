package registry

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	v0 "github.com/modelcontextprotocol/registry/pkg/api/v0"
	"github.com/modelcontextprotocol/registry/pkg/model"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name      string
		baseURL   string
		expectErr bool
	}{
		{
			name:      "valid URL",
			baseURL:   "https://registry.modelcontextprotocol.io",
			expectErr: false,
		},
		{
			name:      "empty URL",
			baseURL:   "",
			expectErr: true,
		},
		{
			name:      "invalid URL",
			baseURL:   "://invalid-url",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.baseURL)
			if tt.expectErr {
				if err == nil {
					t.Errorf("NewClient() expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("NewClient() unexpected error: %v", err)
				return
			}
			if client == nil {
				t.Errorf("NewClient() returned nil client")
			}
		})
	}
}

func TestListServers(t *testing.T) {
	mockResponse := v0.ServerListResponse{
		Servers: []v0.ServerJSON{
			{
				Name:        "test-server",
				Description: "A test server",
				Version:     "1.0.0",
				Status:      model.StatusActive,
			},
		},
		Metadata: &v0.Metadata{
			Count:      1,
			NextCursor: "",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v0/servers" {
			http.NotFound(w, r)
			return
		}

		q := r.URL.Query()
		if cursor := q.Get("cursor"); cursor != "" {
			t.Logf("cursor parameter: %s", cursor)
		}
		if limit := q.Get("limit"); limit != "" {
			t.Logf("limit parameter: %s", limit)
		}
		if updatedSince := q.Get("updated_since"); updatedSince != "" {
			t.Logf("updated_since parameter: %s", updatedSince)
		}
		if search := q.Get("search"); search != "" {
			t.Logf("search parameter: %s", search)
		}
		if version := q.Get("version"); version != "" {
			t.Logf("version parameter: %s", version)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	tests := []struct {
		name string
		opts *ListServersOptions
	}{
		{
			name: "no options",
			opts: nil,
		},
		{
			name: "with cursor and limit",
			opts: &ListServersOptions{
				Cursor: "test-cursor",
				Limit:  50,
			},
		},
		{
			name: "with all query parameters",
			opts: &ListServersOptions{
				Cursor:       "test-cursor",
				Limit:        25,
				UpdatedSince: "2025-08-07T13:15:04.280Z",
				Search:       "filesystem",
				Version:      "latest",
			},
		},
		{
			name: "with limit over max",
			opts: &ListServersOptions{
				Limit: 150, // Should be capped at 100
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.ListServers(context.Background(), tt.opts)
			if err != nil {
				t.Errorf("ListServers() error = %v", err)
				return
			}
			if resp == nil {
				t.Error("ListServers() returned nil response")
				return
			}
			if len(resp.Servers) != 1 {
				t.Errorf("ListServers() expected 1 server, got %d", len(resp.Servers))
			}
		})
	}
}

func TestGetServer(t *testing.T) {
	mockServer := v0.ServerJSON{
		Name:        "test-server",
		Description: "A test server",
		Version:     "1.0.0",
		Status:      model.StatusActive,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v0/servers/valid-id" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockServer)
			return
		}
		if r.URL.Path == "/v0/servers/not-found" {
			http.NotFound(w, r)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	tests := []struct {
		name      string
		serverID  string
		expectErr bool
	}{
		{
			name:      "valid server ID",
			serverID:  "valid-id",
			expectErr: false,
		},
		{
			name:      "empty server ID",
			serverID:  "",
			expectErr: true,
		},
		{
			name:      "server not found",
			serverID:  "not-found",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.GetServer(context.Background(), tt.serverID)
			if tt.expectErr {
				if err == nil {
					t.Errorf("GetServer() expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("GetServer() error = %v", err)
				return
			}
			if resp == nil {
				t.Error("GetServer() returned nil response")
				return
			}
			if resp.Name != mockServer.Name {
				t.Errorf("GetServer() expected name %s, got %s", mockServer.Name, resp.Name)
			}
		})
	}
}

func TestSearchServers(t *testing.T) {
	mockResponse := v0.ServerListResponse{
		Servers: []v0.ServerJSON{
			{
				Name:        "filesystem-server",
				Description: "A filesystem server",
				Version:     "1.0.0",
				Status:      model.StatusActive,
			},
		},
		Metadata: &v0.Metadata{
			Count:      1,
			NextCursor: "",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v0/servers" {
			http.NotFound(w, r)
			return
		}

		search := r.URL.Query().Get("search")
		if search == "" {
			t.Error("Expected search parameter to be set")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	tests := []struct {
		name       string
		searchTerm string
		opts       *ListServersOptions
		expectErr  bool
	}{
		{
			name:       "valid search term",
			searchTerm: "filesystem",
			opts:       nil,
			expectErr:  false,
		},
		{
			name:       "empty search term",
			searchTerm: "",
			opts:       nil,
			expectErr:  true,
		},
		{
			name:       "search with additional options",
			searchTerm: "filesystem",
			opts: &ListServersOptions{
				Version: "latest",
				Limit:   50,
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.SearchServers(context.Background(), tt.searchTerm, tt.opts)
			if tt.expectErr {
				if err == nil {
					t.Errorf("SearchServers() expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("SearchServers() error = %v", err)
				return
			}
			if resp == nil {
				t.Error("SearchServers() returned nil response")
			}
		})
	}
}

func TestGetLatestServers(t *testing.T) {
	mockResponse := v0.ServerListResponse{
		Servers: []v0.ServerJSON{
			{
				Name:        "test-server",
				Description: "A test server",
				Version:     "2.0.0",
				Status:      model.StatusActive,
			},
		},
		Metadata: &v0.Metadata{
			Count:      1,
			NextCursor: "",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v0/servers" {
			http.NotFound(w, r)
			return
		}

		version := r.URL.Query().Get("version")
		if version != "latest" {
			t.Errorf("Expected version parameter to be 'latest', got %s", version)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	resp, err := client.GetLatestServers(context.Background(), nil)
	if err != nil {
		t.Errorf("GetLatestServers() error = %v", err)
		return
	}
	if resp == nil {
		t.Error("GetLatestServers() returned nil response")
	}
}

func TestGetUpdatedServers(t *testing.T) {
	mockResponse := v0.ServerListResponse{
		Servers: []v0.ServerJSON{
			{
				Name:        "updated-server",
				Description: "An updated server",
				Version:     "1.1.0",
				Status:      model.StatusActive,
			},
		},
		Metadata: &v0.Metadata{
			Count:      1,
			NextCursor: "",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v0/servers" {
			http.NotFound(w, r)
			return
		}

		updatedSince := r.URL.Query().Get("updated_since")
		if updatedSince == "" {
			t.Error("Expected updated_since parameter to be set")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	tests := []struct {
		name         string
		updatedSince string
		expectErr    bool
	}{
		{
			name:         "valid timestamp",
			updatedSince: "2025-08-07T13:15:04.280Z",
			expectErr:    false,
		},
		{
			name:         "empty timestamp",
			updatedSince: "",
			expectErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.GetUpdatedServers(context.Background(), tt.updatedSince, nil)
			if tt.expectErr {
				if err == nil {
					t.Errorf("GetUpdatedServers() expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("GetUpdatedServers() error = %v", err)
				return
			}
			if resp == nil {
				t.Error("GetUpdatedServers() returned nil response")
			}
		})
	}
}

func TestGetServerByNameAndVersion(t *testing.T) {
	mockResponse := v0.ServerListResponse{
		Servers: []v0.ServerJSON{
			{
				Name:        "test-server",
				Description: "A test server",
				Version:     "1.0.0",
				Status:      model.StatusActive,
			},
		},
		Metadata: &v0.Metadata{
			Count:      1,
			NextCursor: "",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v0/servers" {
			http.NotFound(w, r)
			return
		}

		search := r.URL.Query().Get("search")
		version := r.URL.Query().Get("version")
		
		// Search parameter should always be set
		if search == "" {
			t.Error("Expected search parameter to be set")
		}
		
		// Version parameter should only be set if requesting "latest"
		if version != "" && version != "latest" {
			t.Errorf("Version parameter should only be 'latest' or empty, got: %s", version)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	tests := []struct {
		name       string
		serverName string
		version    string
		expectErr  bool
	}{
		{
			name:       "valid server and version",
			serverName: "test-server",
			version:    "1.0.0",
			expectErr:  false,
		},
		{
			name:       "empty server name",
			serverName: "",
			version:    "1.0.0",
			expectErr:  true,
		},
		{
			name:       "empty version",
			serverName: "test-server",
			version:    "",
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.GetServerByNameAndVersion(context.Background(), tt.serverName, tt.version)
			if tt.expectErr {
				if err == nil {
					t.Errorf("GetServerByNameAndVersion() expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("GetServerByNameAndVersion() error = %v", err)
				return
			}
			if resp == nil {
				t.Error("GetServerByNameAndVersion() returned nil response")
				return
			}
			if resp.Name != tt.serverName {
				t.Errorf("GetServerByNameAndVersion() expected name %s, got %s", tt.serverName, resp.Name)
			}
		})
	}
}

func TestGetLatestActiveServer(t *testing.T) {
	// Test with multiple versions of servers to verify semantic version comparison
	mockResponseMultiVersion := v0.ServerListResponse{
		Servers: []v0.ServerJSON{
			{
				Name:        "test-server",
				Description: "A test server v1.0.0",
				Version:     "1.0.0",
				Status:      model.StatusActive,
			},
			{
				Name:        "test-server",
				Description: "A test server v2.0.0",
				Version:     "2.0.0",
				Status:      model.StatusActive,
			},
			{
				Name:        "test-server",
				Description: "A test server v1.5.0",
				Version:     "1.5.0",
				Status:      model.StatusActive,
			},
		},
		Metadata: &v0.Metadata{
			Count:      3,
			NextCursor: "",
		},
	}

	// Test with deprecated server
	mockResponseDeprecated := v0.ServerListResponse{
		Servers: []v0.ServerJSON{
			{
				Name:        "deprecated-server",
				Description: "A deprecated server",
				Version:     "1.0.0",
				Status:      model.StatusDeprecated,
			},
		},
		Metadata: &v0.Metadata{
			Count:      1,
			NextCursor: "",
		},
	}

	// Test with mixed active and deprecated versions
	mockResponseMixed := v0.ServerListResponse{
		Servers: []v0.ServerJSON{
			{
				Name:        "mixed-server",
				Description: "Mixed server v3.0.0 (deprecated)",
				Version:     "3.0.0",
				Status:      model.StatusDeprecated,
			},
			{
				Name:        "mixed-server",
				Description: "Mixed server v2.0.0 (active)",
				Version:     "2.0.0",
				Status:      model.StatusActive,
			},
			{
				Name:        "mixed-server",
				Description: "Mixed server v1.0.0 (active)",
				Version:     "1.0.0",
				Status:      model.StatusActive,
			},
		},
		Metadata: &v0.Metadata{
			Count:      3,
			NextCursor: "",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v0/servers" {
			http.NotFound(w, r)
			return
		}

		search := r.URL.Query().Get("search")
		version := r.URL.Query().Get("version")
		if version == "latest" {
			t.Error("Should not use 'latest' version parameter anymore")
		}

		w.Header().Set("Content-Type", "application/json")
		switch search {
		case "deprecated-server":
			json.NewEncoder(w).Encode(mockResponseDeprecated)
		case "mixed-server":
			json.NewEncoder(w).Encode(mockResponseMixed)
		default:
			json.NewEncoder(w).Encode(mockResponseMultiVersion)
		}
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	tests := []struct {
		name           string
		serverName     string
		expectErr      bool
		expectedVersion string
	}{
		{
			name:           "multiple active versions - should return highest",
			serverName:     "test-server",
			expectErr:      false,
			expectedVersion: "2.0.0",
		},
		{
			name:       "empty server name",
			serverName: "",
			expectErr:  true,
		},
		{
			name:       "deprecated server only",
			serverName: "deprecated-server",
			expectErr:  true,
		},
		{
			name:           "mixed active and deprecated - should return highest active",
			serverName:     "mixed-server",
			expectErr:      false,
			expectedVersion: "2.0.0", // Should ignore 3.0.0 because it's deprecated
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.GetLatestActiveServer(context.Background(), tt.serverName)
			if tt.expectErr {
				if err == nil {
					t.Errorf("GetLatestActiveServer() expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("GetLatestActiveServer() error = %v", err)
				return
			}
			if resp == nil {
				t.Error("GetLatestActiveServer() returned nil response")
				return
			}
			if resp.Status != model.StatusActive {
				t.Errorf("GetLatestActiveServer() expected active server, got status %s", resp.Status)
			}
			if tt.expectedVersion != "" && resp.Version != tt.expectedVersion {
				t.Errorf("GetLatestActiveServer() expected version %s, got %s", tt.expectedVersion, resp.Version)
			}
		})
	}
}

func TestSetTimeout(t *testing.T) {
	client, err := NewClient("https://example.com")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	originalTimeout := client.httpClient.Timeout
	newTimeout := 60 * time.Second

	client.SetTimeout(newTimeout)

	if client.httpClient.Timeout != newTimeout {
		t.Errorf("SetTimeout() expected timeout %v, got %v", newTimeout, client.httpClient.Timeout)
	}

	if client.httpClient.Timeout == originalTimeout {
		t.Error("SetTimeout() did not change the timeout")
	}
}