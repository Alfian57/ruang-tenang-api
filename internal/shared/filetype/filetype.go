package filetype

import (
	"bytes"
	"io"
	"net/http"
)

// DetectContentType reads up to 512 bytes from r and returns the sniffed MIME
// type. It returns the sniffed content type and a reader that replays the
// already-read bytes so callers can reuse the stream.
//
// We sniff the actual file bytes instead of trusting the client-supplied
// Content-Type header, which can be trivially spoofed.
func DetectContentType(r io.Reader) (mimeType string, replay io.Reader, err error) {
	head := make([]byte, 512)
	n, err := io.ReadFull(r, head)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return "", nil, err
	}
	mimeType = http.DetectContentType(head[:n])
	return mimeType, io.MultiReader(bytes.NewReader(head[:n]), r), nil
}

// IsImage reports whether mimeType is an allowed image type.
func IsImage(mimeType string) bool {
	return imageTypes[mimeType]
}

// IsAudio reports whether mimeType is an allowed audio type.
func IsAudio(mimeType string) bool {
	return audioTypes[mimeType]
}

// ExtensionFor returns the safe file extension for an allowed MIME type, or
// empty string when the type is not whitelisted.
func ExtensionFor(mimeType string) string {
	return mimeToExt[mimeType]
}

// AllowedImageMIMEs returns the set of accepted image MIME types.
func AllowedImageMIMEs() []string {
	return []string{
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/webp",
	}
}

// AllowedAudioMIMEs returns the set of accepted audio MIME types.
func AllowedAudioMIMEs() []string {
	return []string{
		"audio/mpeg",
		"audio/wav",
		"audio/ogg",
		"audio/x-wav",
	}
}

var (
	imageTypes = map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	audioTypes = map[string]bool{
		"audio/mpeg": true,
		"audio/wav":  true,
		"audio/ogg":  true,
		// Some browsers/buckets tag wav as audio/x-wav; accept it but normalize.
		"audio/x-wav": true,
	}
	mimeToExt = map[string]string{
		"image/jpeg":  ".jpg",
		"image/png":   ".png",
		"image/gif":   ".gif",
		"image/webp":  ".webp",
		"audio/mpeg":  ".mp3",
		"audio/wav":   ".wav",
		"audio/ogg":   ".ogg",
		"audio/x-wav": ".wav",
	}
)
