package development

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func withTempWorkingDir(t *testing.T) string {
	t.Helper()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir tmp: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(oldWD)
	})
	return tmp
}

func TestFileExists(t *testing.T) {
	tmp := withTempWorkingDir(t)

	if fileExists(filepath.Join(tmp, "missing.txt")) {
		t.Fatal("expected missing file to return false")
	}

	p := filepath.Join(tmp, "exists.txt")
	if err := os.WriteFile(p, []byte("ok"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if !fileExists(p) {
		t.Fatal("expected existing file to return true")
	}
}

func TestDownloadAssetAndCopyToUploads(t *testing.T) {
	withTempWorkingDir(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("asset-data"))
	}))
	defer ts.Close()

	storagePath, err := downloadAsset(ts.URL, "sample.txt", "images")
	if err != nil {
		t.Fatalf("downloadAsset failed: %v", err)
	}
	if !strings.Contains(storagePath, filepath.Join("storage", "images", "sample.txt")) {
		t.Fatalf("unexpected storage path: %s", storagePath)
	}

	uploadURL := copyToUploads(storagePath, "images")
	if !strings.HasPrefix(uploadURL, "/uploads/images/") {
		t.Fatalf("unexpected upload URL: %s", uploadURL)
	}
}

func TestDownloadAssetNon200(t *testing.T) {
	withTempWorkingDir(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer ts.Close()

	if _, err := downloadAsset(ts.URL, "bad.txt", "images"); err == nil {
		t.Fatal("expected error for non-200 response")
	}
}

func TestGetSeedAssetAndImageFromLocal(t *testing.T) {
	withTempWorkingDir(t)

	storageImage := filepath.Join("storage", "images")
	if err := os.MkdirAll(storageImage, 0755); err != nil {
		t.Fatalf("mkdir storage image: %v", err)
	}
	if err := os.WriteFile(filepath.Join(storageImage, "avatar-1.jpg"), []byte("img"), 0644); err != nil {
		t.Fatalf("write storage image: %v", err)
	}

	p := getSeedAsset("avatar-1.jpg", "images")
	if !strings.HasPrefix(p, "/uploads/images/") {
		t.Fatalf("expected upload url from storage asset, got %s", p)
	}

	p2 := getOrDownloadImage("", "avatar-1.jpg")
	if !strings.HasPrefix(p2, "/uploads/images/") {
		t.Fatalf("expected upload url from getOrDownloadImage local fallback, got %s", p2)
	}

	if miss := getSeedAudio("missing.mp3"); miss != "" {
		t.Fatalf("expected empty string for missing audio, got %s", miss)
	}
}

func TestGetOrDownloadAssetAndImage_DownloadPath(t *testing.T) {
	withTempWorkingDir(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("downloaded"))
	}))
	defer ts.Close()

	assetURL := getOrDownloadAsset(ts.URL, "remote.jpg", "images")
	if !strings.HasPrefix(assetURL, "/uploads/images/") {
		t.Fatalf("expected uploaded image url, got %s", assetURL)
	}

	imgURL := getOrDownloadImage(ts.URL, "remote2.jpg")
	if !strings.HasPrefix(imgURL, "/uploads/images/") {
		t.Fatalf("expected uploaded image url from getOrDownloadImage, got %s", imgURL)
	}
}

func TestGetOrDownloadAsset_ErrorBranches(t *testing.T) {
	withTempWorkingDir(t)

	t.Run("existing storage file but copy to uploads fails", func(t *testing.T) {
		if err := os.MkdirAll(filepath.Join("storage", "images"), 0755); err != nil {
			t.Fatalf("mkdir storage/images: %v", err)
		}
		if err := os.WriteFile(filepath.Join("storage", "images", "already.jpg"), []byte("img"), 0644); err != nil {
			t.Fatalf("write storage image: %v", err)
		}
		if err := os.WriteFile("uploads", []byte("not-a-dir"), 0644); err != nil {
			t.Fatalf("prepare uploads as file: %v", err)
		}

		if got := getOrDownloadAsset("http://invalid.local/should-not-be-used", "already.jpg", "images"); got != "" {
			t.Fatalf("expected empty result when copyToUploads fails, got %s", got)
		}
		if err := os.Remove("uploads"); err != nil {
			t.Fatalf("cleanup uploads file: %v", err)
		}
	})

	t.Run("missing file and download fails", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
		}))
		defer ts.Close()

		if got := getOrDownloadAsset(ts.URL, "missing.jpg", "images"); got != "" {
			t.Fatalf("expected empty result when downloadAsset fails, got %s", got)
		}
	})
}

func TestDownloadAsset_ErrorBranches(t *testing.T) {
	withTempWorkingDir(t)

	if _, err := downloadAsset("http://127.0.0.1:0/unreachable", "x.txt", "images"); err == nil {
		t.Fatal("expected network error for invalid url")
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer ts.Close()

	if _, err := downloadAsset(ts.URL, "nested/path.txt", "images"); err == nil {
		t.Fatal("expected create file error for nested filename path")
	}
}

func TestCopyToUploads_ErrorBranches(t *testing.T) {
	withTempWorkingDir(t)

	if err := os.WriteFile("uploads", []byte("not-a-dir"), 0644); err != nil {
		t.Fatalf("prepare uploads as file: %v", err)
	}
	if got := copyToUploads("missing.txt", "images"); got != "" {
		t.Fatalf("expected empty when mkdirAll fails, got %s", got)
	}

	if err := os.Remove("uploads"); err != nil {
		t.Fatalf("remove uploads file: %v", err)
	}
	if got := copyToUploads("missing.txt", "images"); got != "" {
		t.Fatalf("expected empty when source open fails, got %s", got)
	}
}

func TestGetSeedAsset_BundledAndPlaceholderBranches(t *testing.T) {
	withTempWorkingDir(t)

	if err := os.MkdirAll(filepath.Join("assets", "images"), 0755); err != nil {
		t.Fatalf("mkdir bundled assets: %v", err)
	}
	if err := os.WriteFile(filepath.Join("assets", "images", "bundled.jpg"), []byte("img"), 0644); err != nil {
		t.Fatalf("write bundled asset: %v", err)
	}

	bundled := getSeedAsset("bundled.jpg", "images")
	if !strings.HasPrefix(bundled, "/uploads/images/") {
		t.Fatalf("expected upload url from bundled asset, got %s", bundled)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("placeholder"))
	}))
	defer ts.Close()

	oldURL, hadOld := placeholderImages["test-placeholder.jpg"]
	placeholderImages["test-placeholder.jpg"] = ts.URL
	t.Cleanup(func() {
		if hadOld {
			placeholderImages["test-placeholder.jpg"] = oldURL
		} else {
			delete(placeholderImages, "test-placeholder.jpg")
		}
	})

	fromPlaceholder := getSeedAsset("test-placeholder.jpg", "images")
	if !strings.HasPrefix(fromPlaceholder, "/uploads/images/") {
		t.Fatalf("expected upload url from placeholder download, got %s", fromPlaceholder)
	}
}

func TestGetOrDownloadImage_EmptyURLAndNoLocal(t *testing.T) {
	withTempWorkingDir(t)
	if got := getOrDownloadImage("", "definitely-missing.jpg"); got != "" {
		t.Fatalf("expected empty result when local missing and URL empty, got %s", got)
	}
}
