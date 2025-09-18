package server

import "fmt"

type PackageTypeNotFoundError struct {
	PackageType           string
	AvailablePackageTypes []string
}

func (e *PackageTypeNotFoundError) Error() string {
	return fmt.Sprintf("no packages of type %q found (available types: %s)", e.PackageType, e.AvailablePackageTypes)
}

type TransportTypeNotFoundError struct {
	PackageType             string
	TransportType           string
	AvailableTransportTypes []string
}

func (e *TransportTypeNotFoundError) Error() string {
	return fmt.Sprintf("no packages of type %q with transport type %q found (available transports: %s)", e.PackageType, e.TransportType, e.AvailableTransportTypes)
}
