package presentation

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	storageDir = "storage"
	uploadsDir = "uploads"
)

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return !info.IsDir()
}

// copyToUploads copies a file from storage to uploads directory with unique name
func copyToUploads(storagePath, subDir string) string {
	// Create uploads directory if not exists
	uploadDir := filepath.Join(uploadsDir, subDir)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Printf("⚠️ Failed to create upload dir: %v", err)
		return ""
	}

	// Generate unique filename
	ext := filepath.Ext(storagePath)
	baseName := filepath.Base(storagePath)
	baseName = baseName[:len(baseName)-len(ext)]
	timestamp := time.Now().UnixNano()
	newFileName := fmt.Sprintf("%s_%d%s", baseName, timestamp, ext)

	// Copy file
	src, err := os.Open(storagePath)
	if err != nil {
		log.Printf("⚠️ Failed to open source file: %v", err)
		return ""
	}
	defer src.Close()

	dstPath := filepath.Join(uploadDir, newFileName)
	dst, err := os.Create(dstPath)
	if err != nil {
		log.Printf("⚠️ Failed to create destination file: %v", err)
		return ""
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		log.Printf("⚠️ Failed to copy file: %v", err)
		return ""
	}

	// Return relative URL path so it works behind any public domain/proxy.
	// Matches the runtime upload handler (/uploads/<category>/<file>) instead of
	// hardcoding http://localhost:8080, which is unreachable from the browser.
	return fmt.Sprintf("/uploads/%s/%s", subDir, newFileName)
}

// getSeedAsset copies a bundled seed asset into uploads and returns its URL.
func getSeedAsset(filename, assetType string) string {
	storagePath := filepath.Join(storageDir, assetType, filename)
	if fileExists(storagePath) {
		return copyToUploads(storagePath, assetType)
	}

	log.Printf("⚠️ Missing local seed physical file locally for %s/%s", assetType, filename)

	return ""
}

func getSeedAudio(filename string) string {
	return getSeedAsset(filename, "audio")
}
