package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/shared/filetype"
	"github.com/Alfian57/ruang-tenang-api/internal/shared/imageproc"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	MaxUploadSize = 10 << 20 // 10MB
	UploadDir     = "uploads"
)

type UploadHandler struct{}

func NewUploadHandler() *UploadHandler {
	return &UploadHandler{}
}

// UploadImage godoc
// @Summary Upload an image file
// @Description Upload an image file (jpg, png, gif, webp) with max size 10MB
// @Tags Upload
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "Image file to upload"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Router /upload/image [post]
func (h *UploadHandler) UploadImage(c *gin.Context) {
	result, ok := saveImageUpload(c)
	if !ok {
		return // error response already written
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"url":           result.URL,
		"thumbnail_url": result.ThumbnailURL,
		"filename":      result.Filename,
	}, "File uploaded successfully"))
}

// UploadAudio godoc
// @Summary Upload an audio file
// @Description Upload an audio file (mp3, wav, ogg) with max size 10MB
// @Tags Upload
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "Audio file to upload"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Router /upload/audio [post]
func (h *UploadHandler) UploadAudio(c *gin.Context) {
	fileURL, filename, ok := saveUpload(c, "audio", filetype.IsAudio)
	if !ok {
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"url":      fileURL,
		"filename": filename,
	}, "File uploaded successfully"))
}

// saveUpload validates a multipart file by sniffing its magic bytes (not the
// client-supplied Content-Type header, which is trivially spoofable), assigns a
// safe extension derived from the detected MIME type, and persists it. The
// allowed predicate gates which MIME types are accepted for the category.
//
// Note: gin's SaveUploadedFile re-opens the FileHeader, so sniffing from the
// multipart.File reader first does not interfere with the subsequent save.
func saveUpload(c *gin.Context, category string, allowed func(string) bool) (string, string, bool) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("No file uploaded"))
		return "", "", false
	}
	defer file.Close()

	if header.Size > MaxUploadSize {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("File size exceeds 10MB limit"))
		return "", "", false
	}

	// Sniff magic bytes instead of trusting the client Content-Type header,
	// which previously allowed stored XSS via HTML/SVG files served from /uploads.
	mimeType, _, err := filetype.DetectContentType(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Failed to read file"))
		return "", "", false
	}
	if !allowed(mimeType) {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid file type"))
		return "", "", false
	}

	uploadPath := filepath.Join(UploadDir, category)
	if err := os.MkdirAll(uploadPath, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to create upload directory"))
		return "", "", false
	}

	// Always derive the extension from the sniffed MIME type; never from the
	// user-supplied filename, which could carry a dangerous extension (.html, .svg).
	ext := filetype.ExtensionFor(mimeType)
	filename := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)
	filePath := filepath.Join(uploadPath, filename)

	if err := c.SaveUploadedFile(header, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to save file"))
		return "", "", false
	}

	fileURL := fmt.Sprintf("/uploads/%s/%s", category, filename)
	return fileURL, filename, true
}


// imageUploadResult memuat URL hasil upload gambar.
type imageUploadResult struct {
	URL          string
	ThumbnailURL string
	Filename     string
}

// saveImageUpload memvalidasi & menyimpan gambar. Untuk foto besar, gambar
// utama di-downscale (hemat bandwidth) dan sebuah thumbnail JPEG dibuat untuk
// daftar/preview. Bila pemrosesan tidak berlaku (mis. GIF animasi), berkas
// asli disimpan apa adanya.
func saveImageUpload(c *gin.Context) (*imageUploadResult, bool) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("No file uploaded"))
		return nil, false
	}
	defer file.Close()

	if header.Size > MaxUploadSize {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("File size exceeds 10MB limit"))
		return nil, false
	}

	// Baca seluruh berkas (≤10MB) untuk disniff sekaligus diproses.
	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Failed to read file"))
		return nil, false
	}

	mimeType, _, err := filetype.DetectContentType(strings.NewReader(string(data)))
	if err != nil || !filetype.IsImage(mimeType) {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid file type"))
		return nil, false
	}

	uploadPath := filepath.Join(UploadDir, "images")
	if err := os.MkdirAll(uploadPath, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to create upload directory"))
		return nil, false
	}

	base := fmt.Sprintf("%s_%d", uuid.New().String(), time.Now().Unix())
	ext := filetype.ExtensionFor(mimeType)

	// Coba proses (downscale + thumbnail). Murni Go, tanpa CGO.
	processed, ok := imageproc.Process(data)

	// Tentukan byte & ekstensi gambar utama.
	mainBytes := data
	mainExt := ext
	if ok && processed.Main != nil {
		mainBytes = processed.Main
		mainExt = ".jpg" // hasil downscale di-encode JPEG
	}

	mainName := base + mainExt
	if err := os.WriteFile(filepath.Join(uploadPath, mainName), mainBytes, 0o644); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to save file"))
		return nil, false
	}

	res := &imageUploadResult{
		URL:      fmt.Sprintf("/uploads/images/%s", mainName),
		Filename: mainName,
	}

	// Simpan thumbnail bila tersedia.
	if ok && processed.Thumbnail != nil {
		thumbName := base + "_thumb.jpg"
		if err := os.WriteFile(filepath.Join(uploadPath, thumbName), processed.Thumbnail, 0o644); err == nil {
			res.ThumbnailURL = fmt.Sprintf("/uploads/images/%s", thumbName)
		}
	}
	// Fallback: bila thumbnail gagal, gunakan gambar utama.
	if res.ThumbnailURL == "" {
		res.ThumbnailURL = res.URL
	}

	return res, true
}
