package middleware

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// imageExtensions are file types that must remain inline so <img>/<picture>
// elements can render them. All other served files are forced to download
// (Content-Disposition: attachment) to prevent stored XSS via HTML/SVG/etc.
var imageExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
}

// SafeStatic serves files from rootDir while forcing a download
// (Content-Disposition: attachment) for any file whose extension is not an
// allowed inline image type. This closes the stored-XSS vector where an
// attacker uploads an HTML/SVG file that the browser then renders as a page.
//
// Note: it also sets X-Content-Type-Options: nosniff (already applied globally
// by SecurityHeadersMiddleware) so the browser honors the declared Content-Type.
func SafeStatic(urlPrefix, rootDir string) gin.HandlerFunc {
	fileServer := http.StripPrefix(urlPrefix, http.FileServer(http.Dir(rootDir)))

	return func(c *gin.Context) {
		// Strip the route prefix to get the file path within rootDir.
		relPath := strings.TrimPrefix(c.Request.URL.Path, urlPrefix)
		relPath = strings.TrimPrefix(relPath, "/")

		// Reject path traversal defensively (InputValidationMiddleware also covers this).
		cleaned := filepath.Clean("/" + relPath)
		if strings.Contains(cleaned, "..") {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		fsPath := filepath.Join(rootDir, relPath)
		info, err := os.Stat(fsPath)
		if err != nil || info.IsDir() {
			c.Status(http.StatusNotFound)
			return
		}

		ext := strings.ToLower(filepath.Ext(relPath))
		if !imageExtensions[ext] {
			// Force download for non-image uploads to prevent browser execution
			// of HTML/SVG/scripts served from the uploads directory.
			c.Header("Content-Disposition", "attachment; filename=\""+filepath.Base(relPath)+"\"")
		}

		fileServer.ServeHTTP(c.Writer, c.Request)
	}
}
