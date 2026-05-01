package presentation

import (
	"fmt"
	"io"
	"log"
	"net/http"
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

// downloadAsset downloads a file from URL and saves it to storage directory under the specified type
func downloadAsset(url, filename, assetType string) (string, error) {
	// Create storage directory if not exists
	assetDir := filepath.Join(storageDir, assetType)
	if err := os.MkdirAll(assetDir, 0755); err != nil {
		return "", err
	}

	// Download file
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download %s: %s", assetType, resp.Status)
	}

	// Create file
	filePath := filepath.Join(assetDir, filename)
	out, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return filePath, nil
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

	// Return URL path
	return fmt.Sprintf("/uploads/%s/%s", subDir, newFileName)
}

// getOrDownloadAsset checks if asset exists in storage, downloads if not
func getOrDownloadAsset(url, filename, assetType string) string {
	storagePath := filepath.Join(storageDir, assetType, filename)

	// Check if file exists
	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		log.Printf("    📥 Downloading %s...", filename)
		path, err := downloadAsset(url, filename, assetType)
		if err != nil {
			log.Printf("    ⚠️ Download failed: %v", err)
			return ""
		}
		storagePath = path
	}

	// Copy to uploads and return URL
	return copyToUploads(storagePath, assetType)
}

// getSeedAsset checks storage first, then falls back to download URL when available.
func getSeedAsset(filename, assetType string) string {
	storagePath := filepath.Join(storageDir, assetType, filename)
	if fileExists(storagePath) {
		return copyToUploads(storagePath, assetType)
	}

	// Wait, if it's missing from local but it's in placeholder, we could download it.
	// But since we commented out placeholderImages, this might fail for generated images
	// if they are not in `storage/`.
	// We should just return empty if it doesn't exist locally at this point,
	// except for those still in the placeholder map.
	if url, ok := placeholderImages[filename]; ok {
		return getOrDownloadAsset(url, filename, assetType)
	}

	log.Printf("⚠️ Missing local seed physical file locally for %s/%s", assetType, filename)

	return ""
}

// Helper for existing image logic
func getOrDownloadImage(url, filename string) string {
	// Our generated files don't have URLs, they only have filenames.
	// So we must try to find them locally first regardless of whether url is empty.
	if localOrBundled := getSeedAsset(filename, "images"); localOrBundled != "" {
		return localOrBundled
	}

	if url == "" {
		return ""
	}

	return getOrDownloadAsset(url, filename, "images")
}

func getSeedAudio(filename string) string {
	return getSeedAsset(filename, "audio")
}

// Placeholder images from Unsplash for development
var placeholderImages = map[string]string{
	// Avatars (we have generated these)
	// "avatar-1.jpg": "https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=200&h=200&fit=crop",
	// "avatar-2.jpg": "https://images.unsplash.com/photo-1494790108377-be9c29b29330?w=200&h=200&fit=crop",
	// "avatar-3.jpg": "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=200&h=200&fit=crop",
	// "avatar-4.jpg": "https://images.unsplash.com/photo-1438761681033-6461ffad8d80?w=200&h=200&fit=crop",

	// Articles (we have generated these)
	// "article-mental.jpg":     "https://images.unsplash.com/photo-1544027993-37dbfe43562a?w=800&h=400&fit=crop",
	// "article-meditation.jpg": "https://images.unsplash.com/photo-1506126613408-eca07ce68773?w=800&h=400&fit=crop",
	// "article-tips.jpg":       "https://images.unsplash.com/photo-1499750310107-5fef28a66643?w=800&h=400&fit=crop",
	// "article-sleep.jpg":      "https://images.unsplash.com/photo-1541781774459-bb2af2f05b55?w=800&h=400&fit=crop",
	// "article-stress.jpg":     "https://images.unsplash.com/photo-1515377905703-c4788e51af15?w=800&h=400&fit=crop",

	// Songs (we have generated these)
	// "song-forest.jpg":     "https://images.unsplash.com/photo-1448375240586-882707db888b?w=400&h=400&fit=crop",
	// "song-ocean.jpg":      "https://images.unsplash.com/photo-1505142468610-359e7d316be0?w=400&h=400&fit=crop",
	// "song-rain.jpg":       "https://images.unsplash.com/photo-1515694346937-94d85e41e6f0?w=400&h=400&fit=crop",
	// "song-piano.jpg":      "https://images.unsplash.com/photo-1520523839897-bd0b52f945a0?w=400&h=400&fit=crop",
	// "song-meditation.jpg": "https://images.unsplash.com/photo-1528715471579-d1bcf0ba5e83?w=400&h=400&fit=crop",

	// Categories (we have generated these)
	// "cat-alam.jpg":     "https://images.unsplash.com/photo-1441974231531-c6227db76b6e?w=400&h=400&fit=crop",
	// "cat-piano.jpg":    "https://images.unsplash.com/photo-1552422535-c45813c61732?w=400&h=400&fit=crop",
	// "cat-hujan.jpg":    "https://images.unsplash.com/photo-1534274988757-a28bf1a57c17?w=400&h=400&fit=crop",
	// "cat-laut.jpg":     "https://images.unsplash.com/photo-1507525428034-b723cf961d3e?w=400&h=400&fit=crop",
	// "cat-meditasi.jpg": "https://images.unsplash.com/photo-1593811167562-9cef47bfc4d7?w=400&h=400&fit=crop",

	// Audio
	"song-placeholder.mp3": "https://github.com/rafaelreis-hotmart/Audio-Sample-files/raw/master/sample.mp3", // Small sample mp3
}
