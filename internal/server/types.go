package server

import (
	registryv0 "github.com/modelcontextprotocol/registry/pkg/api/v0"
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
	JSON *registryv0.ServerJSON
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
	return s.JSON.Status == "active"
}

func (s *Spec) IsDeprecated() bool {
	return s.JSON.Status == "deprecated"
}

func (s *Spec) IsDeleted() bool {
	return s.JSON.Status == "deleted"
}

func (s *Spec) Name() string {
	return s.JSON.Name
}

func (s *Spec) Version() string {
	return s.JSON.Version
}
