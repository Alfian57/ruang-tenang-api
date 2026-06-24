package filetype

import (
	"bytes"
	"io"
	"testing"
)

func TestDetectContentType(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  string
	}{
		{name: "png magic bytes", input: []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, want: "image/png"},
		{name: "jpeg magic bytes", input: []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10}, want: "image/jpeg"},
		{name: "gif magic bytes", input: []byte("GIF89a"), want: "image/gif"},
		{name: "html rejected", input: []byte("<html><script>alert(1)</script></html>"), want: "text/html; charset=utf-8"},
		{name: "empty input", input: []byte{}, want: "text/plain; charset=utf-8"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mime, replay, err := DetectContentType(bytes.NewReader(tt.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if mime != tt.want {
				t.Fatalf("mime = %q, want %q", mime, tt.want)
			}
			// replay must restore the original bytes so callers can reuse the stream.
			got, err := io.ReadAll(replay)
			if err != nil {
				t.Fatalf("replay read error: %v", err)
			}
			if !bytes.Equal(got, tt.input) {
				t.Fatalf("replay lost bytes: got %v, want %v", got, tt.input)
			}
		})
	}
}

func TestIsImageRejectsSpoofedType(t *testing.T) {
	// A client could send a .png Content-Type header while uploading HTML.
	// Sniffing the bytes must classify it as text/html, not an image.
	mime, _, _ := DetectContentType(bytes.NewReader([]byte("<svg/onload=alert(1)>")))
	if IsImage(mime) {
		t.Fatalf("HTML/SVG payload classified as image: %q", mime)
	}
}

func TestExtensionFor(t *testing.T) {
	tests := []struct {
		mime string
		want string
	}{
		{"image/jpeg", ".jpg"},
		{"image/png", ".png"},
		{"image/gif", ".gif"},
		{"image/webp", ".webp"},
		{"audio/mpeg", ".mp3"},
		{"audio/wav", ".wav"},
		{"audio/ogg", ".ogg"},
		{"audio/x-wav", ".wav"},
		{"text/html", ""},     // not whitelisted
		{"image/svg+xml", ""}, // not whitelisted -> forces safe handling
	}

	for _, tt := range tests {
		t.Run(tt.mime, func(t *testing.T) {
			if got := ExtensionFor(tt.mime); got != tt.want {
				t.Fatalf("ExtensionFor(%q) = %q, want %q", tt.mime, got, tt.want)
			}
		})
	}
}
