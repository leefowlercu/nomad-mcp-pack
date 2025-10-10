package server

import (
	registryv0 "github.com/modelcontextprotocol/registry/pkg/api/v0"
	"github.com/modelcontextprotocol/registry/pkg/model"
)

type NameSpec struct {
	Namespace string
	Name      string
}

type SearchSpec struct {
	*NameSpec
	VersionSpec string
}

type Spec struct {
	*SearchSpec
	JSON     *registryv0.ServerJSON
	Response *registryv0.ServerResponse // Contains metadata including Status
}

func (n *NameSpec) String() string {
	return n.FullName()
}

func (n *NameSpec) FullName() string {
	return n.Namespace + "/" + n.Name
}

func (s *SearchSpec) String() string {
	return s.NameSpec.String() + "@" + s.VersionSpec
}

func (s *SearchSpec) IsLatest() bool {
	return s.VersionSpec == "latest"
}

func (s *Spec) String() string {
	return s.JSON.Name + "@" + s.JSON.Version
}

func (s *Spec) IsLatest() bool {
	return s.JSON.Version == "latest"
}

func (s *Spec) IsActive() bool {
	return s.getStatus() == model.StatusActive
}

func (s *Spec) IsDeprecated() bool {
	return s.getStatus() == model.StatusDeprecated
}

func (s *Spec) IsDeleted() bool {
	return s.getStatus() == model.StatusDeleted
}

func (s *Spec) Name() string {
	return s.JSON.Name
}

func (s *Spec) Version() string {
	return s.JSON.Version
}

// getStatus extracts status from ServerResponse metadata.
// If ServerResponse is not available, defaults to StatusActive.
func (s *Spec) getStatus() model.Status {
	if s.Response != nil && s.Response.Meta.Official != nil {
		return s.Response.Meta.Official.Status
	}
	// Default to active if no metadata is available
	return model.StatusActive
}
