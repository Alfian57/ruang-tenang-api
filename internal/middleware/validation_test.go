package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSanitizeAndDetectionHelpers(t *testing.T) {
	if got := SanitizeString(" <script>alert(1)</script> "); !strings.Contains(got, "&lt;script&gt;") {
		t.Fatalf("expected escaped string, got %s", got)
	}

	if !DetectSQLInjection("SELECT * FROM users") {
		t.Fatal("expected SQL injection pattern detection")
	}
	if !DetectXSS("<script>alert(1)</script>") {
		t.Fatal("expected XSS pattern detection")
	}
	if !DetectPathTraversal("../../etc/passwd") {
		t.Fatal("expected path traversal detection")
	}

	if !ValidateEmail("user@example.com") || ValidateEmail("bad-email") {
		t.Fatal("email validator mismatch")
	}
	if !ValidateUsername("user_name-1") || ValidateUsername("bad name") {
		t.Fatal("username validator mismatch")
	}
	if !ValidateUUID("123e4567-e89b-12d3-a456-426614174000") || ValidateUUID("invalid") {
		t.Fatal("uuid validator mismatch")
	}
}

func TestValidateAndSanitize(t *testing.T) {
	config := DefaultSanitizationConfig()
	config.MaxLength = 5

	if _, err := ValidateAndSanitize("123456", config); err == nil {
		t.Fatal("expected max length validation error")
	}

	config.MaxLength = 100
	if _, err := ValidateAndSanitize("select * from users", config); err == nil {
		t.Fatal("expected injection validation error")
	}

	config.StripNewlines = true
	got, err := ValidateAndSanitize("hello\nworld", config)
	if err != nil {
		t.Fatalf("unexpected sanitize error: %v", err)
	}
	if strings.Contains(got, "\n") {
		t.Fatalf("expected newlines to be removed, got %q", got)
	}
}

func TestInputValidationMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(InputValidationMiddleware())
	r.GET("/*any", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test?q=select+*+from+users", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for injected query, got %d", w.Code)
	}
}

func TestContentTypeValidationMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(ContentTypeValidationMiddleware())
	r.POST("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodPost, "/", nil)
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing content-type, got %d", w1.Code)
	}

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/", nil)
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 for valid content-type, got %d", w2.Code)
	}

	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodPost, "/", nil)
	req3.Header.Set("Content-Type", "text/plain")
	r.ServeHTTP(w3, req3)
	if w3.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected 415 for unsupported content-type, got %d", w3.Code)
	}

	w4 := httptest.NewRecorder()
	req4 := httptest.NewRequest(http.MethodPost, "/", nil)
	req4.Header.Set("Content-Type", "multipart/form-data; boundary=abc")
	r.ServeHTTP(w4, req4)
	if w4.Code != http.StatusOK {
		t.Fatalf("expected 200 for multipart content-type, got %d", w4.Code)
	}
}

func TestValidationHelpers_AdditionalBranches(t *testing.T) {
	htmlIn := `<p>halo</p><script>alert(1)</script>`
	htmlOut := SanitizeHTML(htmlIn, nil)
	if strings.Contains(strings.ToLower(htmlOut), "<script") {
		t.Fatalf("expected script tag removed, got %q", htmlOut)
	}

	name := " <b>Alice</b> "
	empty := ""
	fields := map[string]*string{"name": &name, "empty": &empty}
	SanitizeRequestBody(fields)
	if !strings.Contains(name, "&lt;b&gt;") {
		t.Fatalf("expected sanitized body field, got %q", name)
	}

	if !DetectPathTraversal("..\\windows\\system32") {
		t.Fatal("expected windows-style path traversal detection")
	}

	ve := &ValidationError{Field: "name", Message: "invalid"}
	if ve.Error() != "invalid" {
		t.Fatalf("unexpected validation error message: %q", ve.Error())
	}
}

func TestInputValidation_PathTraversalAndPass(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(InputValidationMiddleware())
	r.GET("/*any", func(c *gin.Context) { c.Status(http.StatusOK) })

	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/../../etc/passwd", nil)
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for path traversal, got %d", w1.Code)
	}

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/safe-path?q=hello", nil)
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 for safe request, got %d", w2.Code)
	}
}

func TestMaxBodySizeMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(MaxBodySizeMiddleware(5))
	r.POST("/", func(c *gin.Context) {
		if _, err := io.ReadAll(c.Request.Body); err != nil {
			c.Status(http.StatusRequestEntityTooLarge)
			return
		}
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("123456789"))
	r.ServeHTTP(w, req)
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413 for oversized body, got %d", w.Code)
	}
}
