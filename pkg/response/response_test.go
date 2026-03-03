package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/pkg/apperror"
	"github.com/Alfian57/ruang-tenang-api/pkg/ctxutil"
	"github.com/gin-gonic/gin"
)

func parseBody(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse json: %v", err)
	}
	return body
}

func makeResponseContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	ctxutil.SetRequestID(c, "req-1")
	return c, w
}

func TestOKResponse(t *testing.T) {
	c, w := makeResponseContext()
	OK(c, gin.H{"hello": "world"}, "")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse json: %v", err)
	}
	if body["requestId"] != "req-1" {
		t.Fatalf("unexpected requestId: %v", body["requestId"])
	}
}

func TestPaginatedResponse(t *testing.T) {
	c, w := makeResponseContext()
	Paginated(c, []string{"a", "b"}, 2, 10, 25)

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse json: %v", err)
	}
	meta, ok := body["meta"].(map[string]any)
	if !ok {
		t.Fatalf("missing meta in response: %+v", body)
	}
	if int(meta["total_pages"].(float64)) != 3 {
		t.Fatalf("expected total_pages=3, got %v", meta["total_pages"])
	}
}

func TestErrorAndAbort(t *testing.T) {
	c, w := makeResponseContext()
	Error(c, apperror.BadRequest("invalid"))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	c2, w2 := makeResponseContext()
	AbortUnauthorized(c2, "")
	if w2.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w2.Code)
	}
	if !c2.IsAborted() {
		t.Fatal("expected context to be aborted")
	}
}

func TestSuccessHelpers(t *testing.T) {
	t.Run("created", func(t *testing.T) {
		c, w := makeResponseContext()
		Created(c, gin.H{"id": 1}, "")
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", w.Code)
		}
	})

	t.Run("no content", func(t *testing.T) {
		c, w := makeResponseContext()
		NoContent(c)
		c.Writer.WriteHeaderNow()
		if w.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d", w.Code)
		}
	})

	t.Run("deleted custom and default", func(t *testing.T) {
		c1, w1 := makeResponseContext()
		Deleted(c1, "removed")
		if w1.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w1.Code)
		}
		body1 := parseBody(t, w1)
		data1 := body1["data"].(map[string]any)
		if data1["message"] != "removed" {
			t.Fatalf("unexpected custom delete message: %v", data1["message"])
		}

		c2, w2 := makeResponseContext()
		Deleted(c2, "")
		body2 := parseBody(t, w2)
		data2 := body2["data"].(map[string]any)
		if data2["message"] != "Deleted successfully" {
			t.Fatalf("unexpected default delete message: %v", data2["message"])
		}
	})

	t.Run("updated", func(t *testing.T) {
		c, w := makeResponseContext()
		Updated(c, gin.H{"ok": true}, "")
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})
}

func TestPaginationHelpers(t *testing.T) {
	t.Run("paginated total pages minimum 1", func(t *testing.T) {
		c, w := makeResponseContext()
		Paginated(c, []int{}, 1, 10, 0)
		body := parseBody(t, w)
		meta := body["meta"].(map[string]any)
		if int(meta["total_pages"].(float64)) != 1 {
			t.Fatalf("expected total_pages=1, got %v", meta["total_pages"])
		}
	})

	t.Run("paginated with extra meta", func(t *testing.T) {
		c, w := makeResponseContext()
		PaginatedWithMeta(c, []string{"x"}, 2, 2, 5, map[string]any{"foo": "bar"})
		body := parseBody(t, w)
		if body["foo"] != "bar" {
			t.Fatalf("expected extra key foo=bar, got %v", body["foo"])
		}
		if body["requestId"] != "req-1" {
			t.Fatalf("unexpected requestId: %v", body["requestId"])
		}
	})
}

func TestErrorHelpers(t *testing.T) {
	t.Run("error with message", func(t *testing.T) {
		c, w := makeResponseContext()
		ErrorWithMessage(c, http.StatusTeapot, "brew")
		if w.Code != http.StatusTeapot {
			t.Fatalf("expected 418, got %d", w.Code)
		}
	})

	t.Run("bad request wrapper", func(t *testing.T) {
		c, w := makeResponseContext()
		BadRequest(c, "bad")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("unauthorized default", func(t *testing.T) {
		c, w := makeResponseContext()
		Unauthorized(c, "")
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})

	t.Run("forbidden default", func(t *testing.T) {
		c, w := makeResponseContext()
		Forbidden(c, "")
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		c, w := makeResponseContext()
		NotFound(c, "article")
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})

	t.Run("internal error default", func(t *testing.T) {
		c, w := makeResponseContext()
		InternalError(c, "")
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		c, w := makeResponseContext()
		ValidationError(c, []apperror.FieldError{{Field: "email", Message: "invalid"}})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})
}

func TestAbortHelpers(t *testing.T) {
	t.Run("abort with error", func(t *testing.T) {
		c, w := makeResponseContext()
		AbortWithError(c, apperror.NotFound("x"))
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
		if !c.IsAborted() {
			t.Fatal("expected context aborted")
		}
	})

	t.Run("abort forbidden default", func(t *testing.T) {
		c, w := makeResponseContext()
		AbortForbidden(c, "")
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w.Code)
		}
		if !c.IsAborted() {
			t.Fatal("expected context aborted")
		}
	})
}

func TestPaginatedWithMeta_MinTotalPagesAndNoExtraFields(t *testing.T) {
	c, w := makeResponseContext()
	PaginatedWithMeta(c, []string{}, 1, 10, 0, map[string]any{})

	body := parseBody(t, w)
	meta := body["meta"].(map[string]any)
	if int(meta["total_pages"].(float64)) != 1 {
		t.Fatalf("expected total_pages=1, got %v", meta["total_pages"])
	}
	if _, exists := body["foo"]; exists {
		t.Fatalf("did not expect extra fields, got %+v", body)
	}
}
