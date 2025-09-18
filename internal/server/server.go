package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/leefowlercu/go-mcp-registry/mcp"
	"github.com/leefowlercu/nomad-mcp-pack/internal/utils"
	v0 "github.com/modelcontextprotocol/registry/pkg/api/v0"
	"github.com/modelcontextprotocol/registry/pkg/model"
)

func ParseNameSpec(s string) (*NameSpec, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, errors.New("invalid server name format; server name stanza must not be empty")
	}

	if !strings.Contains(s, "/") {
		return nil, fmt.Errorf("invalid server name format %q; expected format like 'io.github.example/server'", s)
	}

	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid server name format %q; expected exactly one '/' separator", s)
	}

	namespace := strings.TrimSpace(parts[0])
	name := strings.TrimSpace(parts[1])

	if namespace == "" {
		return nil, fmt.Errorf("invalid server name format %q; namespace stanza must not be empty", s)
	}

	if name == "" {
		return nil, fmt.Errorf("invalid server name format %q; name stanza must not be empty", s)
	}

	return &NameSpec{
		Namespace: namespace,
		Name:      name,
	}, nil
}

func ParseSearchSpec(s string) (*SearchSpec, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, errors.New("invalid server search format; server search argument must not be empty")
	}

	parts := strings.Split(s, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid server search format %q; expected exactly one '@' separator", s)
	}

	name := strings.TrimSpace(parts[0])
	version := strings.TrimSpace(parts[1])

	if name == "" {
		return nil, errors.New("invalid server name format; server name stanza must not be empty")
	}

	if version == "" {
		return nil, errors.New("invalid server version format; server version stanza must not be empty")
	}

	nameSpec, err := ParseNameSpec(name)
	if err != nil {
		return nil, err
	}

	return &SearchSpec{
		NameSpec:    nameSpec,
		VersionSpec: version,
	}, nil
}

func Find(ctx context.Context, searchSpec *SearchSpec, client *mcp.Client) (*Spec, error) {
	if searchSpec == nil {
		return nil, errors.New("search spec must not be nil")
	}

	if client == nil {
		return nil, errors.New("mcp client must not be nil")
	}

	var srv *v0.ServerJSON
	var err error

	if searchSpec.IsLatest() {
		slog.Info("resolving latest version", "server", searchSpec.FullName())
		srv, err = client.Servers.GetByNameLatest(ctx, searchSpec.FullName())
	} else {
		srv, err = client.Servers.GetByNameExactVersion(ctx, searchSpec.FullName(), searchSpec.VersionSpec)
	}
	if err != nil {
		return nil, fmt.Errorf("failure reading from registry; %w", err)
	}

	if srv == nil {
		return nil, fmt.Errorf("no server %q found with version %q", searchSpec.FullName(), searchSpec.VersionSpec)
	}

	if searchSpec.IsLatest() {
		slog.Info("resolved latest version", "server", srv.Name, "version", srv.Version)
	}

	return &Spec{
		SearchSpec: searchSpec,
		JSON:       srv,
	}, nil
}

func FindPackageWithTransport(server *v0.ServerJSON, packageType, transportType string) (*model.Package, error) {
	if server == nil {
		return nil, fmt.Errorf("server must not be nil")
	}

	registryTransportType := utils.MapToRegistryTransportType(transportType)

	var matchingPackages []model.Package
	var availablePackageTypes []string
	var availableTransportTypes []string

	for _, pkg := range server.Packages {
		if !slices.Contains(availablePackageTypes, pkg.RegistryType) {
			availablePackageTypes = append(availablePackageTypes, pkg.RegistryType)
		}
		if pkg.RegistryType == packageType {
			matchingPackages = append(matchingPackages, pkg)
			if pkg.Transport.Type != "" {
				if !slices.Contains(availableTransportTypes, pkg.Transport.Type) {
					availableTransportTypes = append(availableTransportTypes, pkg.Transport.Type)
				}
			}
		}
	}

	if len(matchingPackages) == 0 {
		return nil, &PackageTypeNotFoundError{
			PackageType:           packageType,
			AvailablePackageTypes: availablePackageTypes,
		}
	}

	for _, pkg := range matchingPackages {
		if pkg.Transport.Type == registryTransportType {
			pkgCopy := pkg
			return &pkgCopy, nil
		}
	}

	uniqueUserTransports := make([]string, 0, len(availableTransportTypes))
	seen := make(map[string]bool)
	for _, transport := range availableTransportTypes {
		userTransport := utils.MapFromRegistryTransportType(transport)
		if !seen[userTransport] {
			uniqueUserTransports = append(uniqueUserTransports, userTransport)
			seen[userTransport] = true
		}
	}

	return nil, &TransportTypeNotFoundError{
		PackageType:             packageType,
		TransportType:           transportType,
		AvailableTransportTypes: uniqueUserTransports,
	}
}
