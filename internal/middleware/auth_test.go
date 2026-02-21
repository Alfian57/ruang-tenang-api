package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/Alfian57/ruang-tenang-api/pkg/utils"
	"github.com/gin-gonic/gin"
)

func issueTestToken(t *testing.T, role string) string {
	t.Helper()
	config.AppConfig = &config.Config{JWTSecret: "test-secret"}
	token, err := utils.GenerateToken(99, "tester@example.com", role, time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	return token
}

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing authorization header", func(t *testing.T) {
		r := gin.New()
		r.Use(AuthMiddleware())
		r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})

	t.Run("invalid header format", func(t *testing.T) {
		r := gin.New()
		r.Use(AuthMiddleware())
		r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Token abc")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		config.AppConfig = &config.Config{JWTSecret: "test-secret"}
		r := gin.New()
		r.Use(AuthMiddleware())
		r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer not-a-valid-token")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})

	t.Run("valid token sets user context", func(t *testing.T) {
		token := issueTestToken(t, "member")
		r := gin.New()
		r.Use(AuthMiddleware())
		r.GET("/", func(c *gin.Context) {
			uid, okUID := GetUserID(c)
			role, okRole := GetUserRole(c)
			if !okUID || !okRole || uid != 99 || role != "member" {
				c.Status(http.StatusInternalServerError)
				return
			}
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})
}

func TestOptionalAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing header continues", func(t *testing.T) {
		r := gin.New()
		r.Use(OptionalAuthMiddleware())
		r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("invalid header continues", func(t *testing.T) {
		r := gin.New()
		r.Use(OptionalAuthMiddleware())
		r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Invalid")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("invalid token continues", func(t *testing.T) {
		config.AppConfig = &config.Config{JWTSecret: "test-secret"}
		r := gin.New()
		r.Use(OptionalAuthMiddleware())
		r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer invalid")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("valid token sets context", func(t *testing.T) {
		token := issueTestToken(t, "admin")
		r := gin.New()
		r.Use(OptionalAuthMiddleware())
		r.GET("/", func(c *gin.Context) {
			uid, okUID := GetUserID(c)
			role, okRole := GetUserRole(c)
			if !okUID || !okRole || uid != 99 || role != "admin" {
				c.Status(http.StatusInternalServerError)
				return
			}
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})
}

func TestRoleMiddlewaresAndHelpers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("admin middleware deny and allow", func(t *testing.T) {
		r := gin.New()
		r.GET("/deny", func(c *gin.Context) {
			AdminMiddleware()(c)
		}, func(c *gin.Context) { c.Status(http.StatusOK) })
		r.GET("/allow", func(c *gin.Context) {
			c.Set("user_role", "admin")
			AdminMiddleware()(c)
		}, func(c *gin.Context) { c.Status(http.StatusOK) })

		w1 := httptest.NewRecorder()
		r.ServeHTTP(w1, httptest.NewRequest(http.MethodGet, "/deny", nil))
		if w1.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w1.Code)
		}

		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/allow", nil))
		if w2.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w2.Code)
		}
	})

	t.Run("member middleware deny and allow", func(t *testing.T) {
		r := gin.New()
		r.GET("/deny", func(c *gin.Context) {
			MemberMiddleware()(c)
		}, func(c *gin.Context) { c.Status(http.StatusOK) })
		r.GET("/allow", func(c *gin.Context) {
			c.Set("user_role", "member")
			MemberMiddleware()(c)
		}, func(c *gin.Context) { c.Status(http.StatusOK) })

		w1 := httptest.NewRecorder()
		r.ServeHTTP(w1, httptest.NewRequest(http.MethodGet, "/deny", nil))
		if w1.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w1.Code)
		}

		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/allow", nil))
		if w2.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w2.Code)
		}
	})

	t.Run("moderator middleware branches", func(t *testing.T) {
		r := gin.New()
		r.GET("/missing", func(c *gin.Context) {
			ModeratorMiddleware()(c)
		}, func(c *gin.Context) { c.Status(http.StatusOK) })
		r.GET("/wrongtype", func(c *gin.Context) {
			c.Set("user_role", 123)
			ModeratorOrAdminMiddleware()(c)
		}, func(c *gin.Context) { c.Status(http.StatusOK) })
		r.GET("/member", func(c *gin.Context) {
			c.Set("user_role", "member")
			ModeratorMiddleware()(c)
		}, func(c *gin.Context) { c.Status(http.StatusOK) })
		r.GET("/moderator", func(c *gin.Context) {
			c.Set("user_role", "moderator")
			ModeratorMiddleware()(c)
		}, func(c *gin.Context) { c.Status(http.StatusOK) })
		r.GET("/admin", func(c *gin.Context) {
			c.Set("user_role", "admin")
			ModeratorMiddleware()(c)
		}, func(c *gin.Context) { c.Status(http.StatusOK) })

		paths := map[string]int{
			"/missing":   http.StatusForbidden,
			"/wrongtype": http.StatusForbidden,
			"/member":    http.StatusForbidden,
			"/moderator": http.StatusOK,
			"/admin":     http.StatusOK,
		}

		for path, expected := range paths {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
			if w.Code != expected {
				t.Fatalf("path %s expected %d, got %d", path, expected, w.Code)
			}
		}
	})

	t.Run("get user helpers", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		if uid, ok := GetUserID(c); ok || uid != 0 {
			t.Fatalf("expected no user id, got %v %v", uid, ok)
		}
		if role, ok := GetUserRole(c); ok || role != "" {
			t.Fatalf("expected no role, got %q %v", role, ok)
		}

		c.Set("user_id", uint(7))
		c.Set("user_role", 123)
		if uid, ok := GetUserID(c); !ok || uid != 7 {
			t.Fatalf("expected user id 7, got %v %v", uid, ok)
		}
		if role, ok := GetUserRole(c); ok || role != "" {
			t.Fatalf("expected invalid role type, got %q %v", role, ok)
		}

		c.Set("user_role", "member")
		if role, ok := GetUserRole(c); !ok || role != "member" {
			t.Fatalf("expected member role, got %q %v", role, ok)
		}
	})
}
