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
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	// Ensure data is written to disk
	if err := destFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}

	// Remove source file after successful copy
	if err := os.Remove(src); err != nil {
		return fmt.Errorf("failed to remove source file: %w", err)
	}

	return nil
}
