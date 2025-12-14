package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	MaxUploadSize = 10 << 20 // 10MB
	UploadDir     = "uploads"
)

var AllowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

var AllowedAudioTypes = map[string]bool{
	"audio/mpeg": true,
	"audio/mp3":  true,
	"audio/wav":  true,
	"audio/ogg":  true,
}

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
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("No file uploaded"))
		return
	}
	defer file.Close()

	// Check file size
	if header.Size > MaxUploadSize {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("File size exceeds 10MB limit"))
		return
	}

	// Check file type
	contentType := header.Header.Get("Content-Type")
	if !AllowedImageTypes[contentType] {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid file type. Allowed: jpg, png, gif, webp"))
		return
	}

	// Create upload directory if not exists
	uploadPath := filepath.Join(UploadDir, "images")
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to create upload directory"))
		return
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = getExtensionFromMime(contentType)
	}
	filename := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)
	filePath := filepath.Join(uploadPath, filename)

	// Save file
	if err := c.SaveUploadedFile(header, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to save file"))
		return
	}

	// Return file URL
	fileURL := fmt.Sprintf("/uploads/images/%s", filename)
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"url":      fileURL,
		"filename": filename,
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
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("No file uploaded"))
		return
	}
	defer file.Close()

	// Check file size
	if header.Size > MaxUploadSize {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("File size exceeds 10MB limit"))
		return
	}

	// Check file type
	contentType := header.Header.Get("Content-Type")
	if !AllowedAudioTypes[contentType] {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid file type. Allowed: mp3, wav, ogg"))
		return
	}

	// Create upload directory if not exists
	uploadPath := filepath.Join(UploadDir, "audio")
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to create upload directory"))
		return
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = getExtensionFromMime(contentType)
	}
	filename := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)
	filePath := filepath.Join(uploadPath, filename)

	// Save file
	if err := c.SaveUploadedFile(header, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to save file"))
		return
	}

	// Return file URL
	fileURL := fmt.Sprintf("/uploads/audio/%s", filename)
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"url":      fileURL,
		"filename": filename,
	}, "File uploaded successfully"))
}

func getExtensionFromMime(mimeType string) string {
	mimeToExt := map[string]string{
		"image/jpeg": ".jpg",
		"image/jpg":  ".jpg",
		"image/png":  ".png",
		"image/gif":  ".gif",
		"image/webp": ".webp",
		"audio/mpeg": ".mp3",
		"audio/mp3":  ".mp3",
		"audio/wav":  ".wav",
		"audio/ogg":  ".ogg",
	}
	if ext, ok := mimeToExt[strings.ToLower(mimeType)]; ok {
		return ext
	}
	return ""
}
