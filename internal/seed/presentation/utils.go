package presentation

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	billingapp "github.com/Alfian57/ruang-tenang-api/internal/features/billing/application"
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

// copyToUploads copies a bundled asset to a stable upload URL. Reusing the
// original filename keeps presentation seeding idempotent and avoids orphaned
// timestamped copies on every run.
func copyToUploads(storagePath, subDir string) string {
	// Create uploads directory if not exists
	uploadDir := filepath.Join(uploadsDir, subDir)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Printf("⚠️ Failed to create upload dir: %v", err)
		return ""
	}

	newFileName := filepath.Base(storagePath)

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

// seededFreemiumChatUsage leaves exactly one daily chat message available for
// the presentation freemium account, so the next request exercises the real
// quota boundary instead of displaying an arbitrary sample count.
func seededFreemiumChatUsage() int {
	limit := 100 // matches the API/config default
	if configured, err := strconv.Atoi(strings.TrimSpace(os.Getenv("CHAT_DAILY_MESSAGE_LIMIT"))); err == nil && configured > 0 {
		limit = configured
	}
	if limit > 0 {
		return limit - 1
	}
	return 0
}

func seededChatQuotaWindowStart(now time.Time) time.Time {
	return billingapp.ChatQuotaWindowStart(now, os.Getenv("CHAT_QUOTA_RESET_INTERVAL"))
}
