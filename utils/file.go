package utils

import (
	"fmt"
	"io"
	"os"
)

// CopyFile copies a file from src to dst in a cross-platform way
func CopyFile(src, dst string) error {
	// Try rename first (fast and atomic)
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	// If rename fails (e.g., cross-partition), do a manual copy
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("CopyFile: failed to open source '%s': %w", src, err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("CopyFile: failed to create destination '%s': %w", dst, err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("CopyFile: failed to copy '%s' -> '%s': %w", src, dst, err)
	}

	// Ensure data is written to disk
	if err := destFile.Sync(); err != nil {
		return fmt.Errorf("CopyFile: failed to sync '%s': %w", dst, err)
	}

	// Remove source file after successful copy
	if err := os.Remove(src); err != nil {
		return fmt.Errorf("CopyFile: failed to remove source '%s': %w", src, err)
	}

	return nil
}
