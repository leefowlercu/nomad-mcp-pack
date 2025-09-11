package generate

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func (g *Generator) createArchive(tempPackDir string) error {
	packName := sanitizePackName(g.serverSpec.FullName()) + "-" + g.serverSpec.Version + "-" + g.pkg.RegistryType
	archivePath := filepath.Join(g.options.OutputDir, packName+".zip")

	if _, err := os.Stat(archivePath); err == nil && !g.options.Force {
		return fmt.Errorf("pack archive %s already exists (use --force to overwrite)", archivePath)
	}

	archiveFile, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("failed to create archive file: %w", err)
	}
	defer archiveFile.Close()

	zipWriter := zip.NewWriter(archiveFile)
	defer zipWriter.Close()

	err = filepath.Walk(tempPackDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filePath == tempPackDir {
			return nil
		}

		relPath, err := filepath.Rel(tempPackDir, filePath)
		if err != nil {
			return err
		}

		// Create zip entry with forward slashes for cross-platform compatibility
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
		return fmt.Errorf("failed to create archive: %w", err)
	}

	return nil
}