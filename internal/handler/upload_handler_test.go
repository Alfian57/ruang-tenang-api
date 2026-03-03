package handler

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func buildMultipartRequest(t *testing.T, fieldName, fileName, contentType string, payload []byte) *http.Request {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	var (
		part io.Writer
		err  error
	)
	if contentType == "" {
		part, err = writer.CreateFormFile(fieldName, fileName)
	} else {
		header := make(textproto.MIMEHeader)
		header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldName, fileName))
		header.Set("Content-Type", contentType)
		part, err = writer.CreatePart(header)
	}
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(payload); err != nil {
		t.Fatalf("write payload: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req
}

func withTempWorkingDir(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(oldWd)
	})
}

func TestUploadHandler_ImageAndAudio(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUploadHandler()

	t.Run("image-no-file", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/upload/image", nil)
		h.UploadImage(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("image-too-large", func(t *testing.T) {
		withTempWorkingDir(t)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := buildMultipartRequest(t, "file", "big.png", "image/png", bytes.Repeat([]byte("a"), MaxUploadSize+1))

		c.Request = req
		h.UploadImage(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("image-invalid-type", func(t *testing.T) {
		withTempWorkingDir(t)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := buildMultipartRequest(t, "file", "a.txt", "text/plain", []byte("abc"))

		c.Request = req
		h.UploadImage(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("image-success", func(t *testing.T) {
		withTempWorkingDir(t)
		AllowedImageTypes[""] = true
		AllowedImageTypes["application/octet-stream"] = true
		defer delete(AllowedImageTypes, "")
		defer delete(AllowedImageTypes, "application/octet-stream")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := buildMultipartRequest(t, "file", "a.png", "image/png", []byte("img"))

		c.Request = req
		h.UploadImage(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
		}

		if _, err := os.Stat(filepath.Join(UploadDir, "images")); err != nil {
			t.Fatalf("expected image upload dir: %v", err)
		}
	})

	t.Run("audio-invalid-type", func(t *testing.T) {
		withTempWorkingDir(t)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := buildMultipartRequest(t, "file", "a.txt", "text/plain", []byte("abc"))

		c.Request = req
		h.UploadAudio(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("audio-no-file", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/upload/audio", nil)
		h.UploadAudio(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("audio-too-large", func(t *testing.T) {
		withTempWorkingDir(t)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := buildMultipartRequest(t, "file", "big.mp3", "audio/mpeg", bytes.Repeat([]byte("a"), MaxUploadSize+1))

		c.Request = req
		h.UploadAudio(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("audio-success", func(t *testing.T) {
		withTempWorkingDir(t)
		AllowedAudioTypes[""] = true
		AllowedAudioTypes["application/octet-stream"] = true
		defer delete(AllowedAudioTypes, "")
		defer delete(AllowedAudioTypes, "application/octet-stream")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := buildMultipartRequest(t, "file", "a.mp3", "audio/mpeg", []byte("aud"))

		c.Request = req
		h.UploadAudio(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
		}

		if _, err := os.Stat(filepath.Join(UploadDir, "audio")); err != nil {
			t.Fatalf("expected audio upload dir: %v", err)
		}
	})
}

func TestGetExtensionFromMime(t *testing.T) {
	if getExtensionFromMime("image/jpeg") != ".jpg" {
		t.Fatalf("expected .jpg for jpeg")
	}
	if getExtensionFromMime("AUDIO/WAV") != ".wav" {
		t.Fatalf("expected case-insensitive .wav")
	}
	if getExtensionFromMime("application/octet-stream") != "" {
		t.Fatalf("expected empty for unknown mime")
	}
}

func TestUploadHandler_MkdirAndSaveErrorBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUploadHandler()

	t.Run("image mkdir fail when uploads is file", func(t *testing.T) {
		withTempWorkingDir(t)
		if err := os.WriteFile(UploadDir, []byte("x"), 0644); err != nil {
			t.Fatalf("write uploads file: %v", err)
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := buildMultipartRequest(t, "file", "a.png", "image/png", []byte("img"))
		c.Request = req
		h.UploadImage(c)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d body=%s", w.Code, w.Body.String())
		}
	})

	t.Run("audio mkdir fail when uploads is file", func(t *testing.T) {
		withTempWorkingDir(t)
		if err := os.WriteFile(UploadDir, []byte("x"), 0644); err != nil {
			t.Fatalf("write uploads file: %v", err)
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := buildMultipartRequest(t, "file", "a.mp3", "audio/mpeg", []byte("aud"))
		c.Request = req
		h.UploadAudio(c)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d body=%s", w.Code, w.Body.String())
		}
	})

	t.Run("image save fail due read-only directory", func(t *testing.T) {
		withTempWorkingDir(t)
		if err := os.MkdirAll(UploadDir, 0755); err != nil {
			t.Fatalf("mkdir uploads: %v", err)
		}
		imageDir := filepath.Join(UploadDir, "images")
		if err := os.MkdirAll(imageDir, 0755); err != nil {
			t.Fatalf("mkdir images: %v", err)
		}
		if err := os.Chmod(imageDir, 0555); err != nil {
			t.Fatalf("chmod images: %v", err)
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := buildMultipartRequest(t, "file", "a.png", "image/png", []byte("img"))
		c.Request = req
		h.UploadImage(c)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d body=%s", w.Code, w.Body.String())
		}
	})

	t.Run("audio save fail due read-only directory", func(t *testing.T) {
		withTempWorkingDir(t)
		if err := os.MkdirAll(UploadDir, 0755); err != nil {
			t.Fatalf("mkdir uploads: %v", err)
		}
		audioDir := filepath.Join(UploadDir, "audio")
		if err := os.MkdirAll(audioDir, 0755); err != nil {
			t.Fatalf("mkdir audio: %v", err)
		}
		if err := os.Chmod(audioDir, 0555); err != nil {
			t.Fatalf("chmod audio: %v", err)
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := buildMultipartRequest(t, "file", "a.mp3", "audio/mpeg", []byte("aud"))
		c.Request = req
		h.UploadAudio(c)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d body=%s", w.Code, w.Body.String())
		}
	})
}

func TestUploadHandler_ExtensionFallbackWhenFilenameHasNoExt(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUploadHandler()

	t.Run("image ext fallback", func(t *testing.T) {
		withTempWorkingDir(t)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := buildMultipartRequest(t, "file", "imagefile", "image/png", []byte("img"))
		c.Request = req
		h.UploadImage(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), ".png") {
			t.Fatalf("expected generated filename to include .png, body=%s", w.Body.String())
		}
	})

	t.Run("audio ext fallback", func(t *testing.T) {
		withTempWorkingDir(t)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := buildMultipartRequest(t, "file", "audiofile", "audio/mpeg", []byte("aud"))
		c.Request = req
		h.UploadAudio(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), ".mp3") {
			t.Fatalf("expected generated filename to include .mp3, body=%s", w.Body.String())
		}
	})
}
