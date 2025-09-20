package watcher

import (
	"slices"
	"strings"

	"github.com/leefowlercu/go-mcp-registry/mcp"
	"github.com/leefowlercu/nomad-mcp-pack/internal/generator"
	"github.com/leefowlercu/nomad-mcp-pack/internal/utils"
	v0 "github.com/modelcontextprotocol/registry/pkg/api/v0"
	"github.com/modelcontextprotocol/registry/pkg/model"
)

type WatcherConfig struct {
	PollInterval    int
	StateFilePath   string
	MaxConcurrent   int
	AllowDeprecated bool
	NameFilter      *ServerNameFilter
	PackageFilter   *PackageTypeFilter
	TransportFilter *TransportTypeFilter
}

type Watcher struct {
	client       *mcp.Client
	config       *WatcherConfig
	state        *WatchState
	generateOpts generator.Options
}

type ServerNameFilter struct {
	Names []string
}

type PackageTypeFilter struct {
	Types []string
}

type TransportTypeFilter struct {
	Types []string
}

func (f *ServerNameFilter) Matches(serverName string) bool {
	if len(f.Names) == 0 {
		return true
	}

	return slices.Contains(f.Names, serverName)
}

func (f *PackageTypeFilter) Matches(packageType string) bool {
	packageTypeLower := strings.ToLower(packageType)

	for _, filterType := range f.Types {
		if strings.ToLower(filterType) == packageTypeLower {
			return true
		}
	}
	return false
}

func (f *TransportTypeFilter) Matches(registryTransportType string) bool {
	userTransportType := utils.MapFromRegistryTransportType(registryTransportType)
	userTransportTypeLower := strings.ToLower(userTransportType)

	for _, filterType := range f.Types {
		if strings.ToLower(filterType) == userTransportTypeLower {
			return true
		}
	}
	return false
}

type ServerGenerateTask struct {
	Server  v0.ServerJSON
	Package *model.Package
}

type packGenSemaphore struct {
	sem chan struct{}
}

func newPackGenSemaphore(maxConcurrent int) *packGenSemaphore {
	return &packGenSemaphore{
		sem: make(chan struct{}, maxConcurrent),
	}
}

func (p *packGenSemaphore) Acquire() {
	p.sem <- struct{}{}
}

func (p *packGenSemaphore) Release() {
	<-p.sem
}
