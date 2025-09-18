package generator

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func (g *Generator) createArchive(ctx context.Context, generateDir string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	archivePath := filepath.Join(g.options.OutputDir, g.packName+".zip")

	if _, err := os.Stat(archivePath); err == nil && !g.options.ForceOverwrite {
		return fmt.Errorf("pack archive %s already exists", archivePath)
	}

	archiveFile, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("failed to create archive file; %w", err)
	}
	defer archiveFile.Close()

	zipWriter := zip.NewWriter(archiveFile)
	defer zipWriter.Close()

	err = filepath.Walk(generateDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filePath == generateDir {
			return nil
		}

		relPath, err := filepath.Rel(generateDir, filePath)
		if err != nil {
			return err
		}

		zipPath := strings.ReplaceAll(relPath, string(filepath.Separator), "/")

		if info.IsDir() {
			_, err := zipWriter.Create(zipPath + "/")
			return err
		}

		zipFile, err := zipWriter.Create(zipPath)
		if err != nil {
			return err
		}

		srcFile, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		_, err = io.Copy(zipFile, srcFile)
		return err
	})

	if err != nil {
		return fmt.Errorf("failed to create archive; %w", err)
	}

	return nil
}
