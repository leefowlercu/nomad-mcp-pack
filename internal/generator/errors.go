package generator

import "errors"

var (
	ErrPackDirectoryExists = errors.New("pack directory already exists")
	ErrPackArchiveExists   = errors.New("pack archive already exists")
)
