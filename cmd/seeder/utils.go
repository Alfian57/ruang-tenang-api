package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// copyFile copies a file from src to dst, creating directories as needed
func copyFile(src, dst string) error {
	// Create destination directory
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// copyAsset copies an asset from assets/ to uploads/ and returns the URL path
func copyAsset(assetPath, subDir string) string {
	// Generate unique filename with timestamp
	ext := filepath.Ext(assetPath)
	baseName := filepath.Base(assetPath)
	baseName = baseName[:len(baseName)-len(ext)]
	timestamp := time.Now().UnixNano()
	newFileName := fmt.Sprintf("%s_%d%s", baseName, timestamp, ext)

	dstPath := filepath.Join("uploads", subDir, newFileName)
	if err := copyFile(assetPath, dstPath); err != nil {
		log.Printf("  ⚠️ Failed to copy %s: %v", assetPath, err)
		return ""
	}

	// Return URL path (relative to server root)
	return fmt.Sprintf("/uploads/%s/%s", subDir, newFileName)
}
