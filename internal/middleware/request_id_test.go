package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/pkg/ctxutil"
	"github.com/gin-gonic/gin"
)

func TestRequestIDMiddlewareGeneratesAndReuses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestIDMiddleware())
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, ctxutil.GetRequestID(c))
	})

	t.Run("generates when missing", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		head := w.Header().Get(RequestIDHeader)
		if head == "" {
			t.Fatal("expected request id header")
		}
		if w.Body.String() != head {
			t.Fatalf("request id in context mismatch: body=%s header=%s", w.Body.String(), head)
		}
	})

	t.Run("reuses client header", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(RequestIDHeader, "client-id-123")
		r.ServeHTTP(w, req)

		if got := w.Header().Get(RequestIDHeader); got != "client-id-123" {
			t.Fatalf("expected client header to be reused, got %s", got)
		}
	})
}
