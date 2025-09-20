package watcher

import (
	"errors"
	"fmt"
)

var ErrGracefulShutdown = errors.New("graceful shutdown")

type PackGenerationErrors struct {
	CriticalErrs []error
}

func (p PackGenerationErrors) Error() string {
	return fmt.Sprintf("pack generation completed with %d critical errors", len(p.CriticalErrs))
}

func (p PackGenerationErrors) Unwrap() []error {
	return p.CriticalErrs
}
