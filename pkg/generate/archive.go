package generate

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// createArchive creates a ZIP archive from the generated pack files
func (g *Generator) createArchive(tempPackDir string) error {
	// Create the archive file path
	packName := sanitizePackName(g.serverSpec.ServerName) + "-" + g.serverSpec.Version + "-" + g.pkg.RegistryType
	archivePath := filepath.Join(g.options.OutputDir, packName+".zip")

	// Check if archive already exists
	if _, err := os.Stat(archivePath); err == nil && !g.options.Force {
		return fmt.Errorf("pack archive %s already exists (use --force to overwrite)", archivePath)
	}

	// Create the archive file
	archiveFile, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("failed to create archive file: %w", err)
	}
	defer archiveFile.Close()

	// Create zip writer
	zipWriter := zip.NewWriter(archiveFile)
	defer zipWriter.Close()

	// Walk the temporary pack directory and add files to zip
	err = filepath.Walk(tempPackDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if filePath == tempPackDir {
			return nil
		}

		// Get relative path within the pack
		relPath, err := filepath.Rel(tempPackDir, filePath)
		if err != nil {
			return err
		}

		// Create zip entry with forward slashes for cross-platform compatibility
		zipPath := strings.ReplaceAll(relPath, string(filepath.Separator), "/")

		if info.IsDir() {
			// Create directory entry in zip
			_, err := zipWriter.Create(zipPath + "/")
			return err
		}

		// Create file entry in zip
		zipFile, err := zipWriter.Create(zipPath)
		if err != nil {
			return err
		}

		// Open source file
		srcFile, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		// Copy file contents to zip
		_, err = io.Copy(zipFile, srcFile)
		return err
	})

	if err != nil {
		return fmt.Errorf("failed to create archive: %w", err)
	}

	return nil
}