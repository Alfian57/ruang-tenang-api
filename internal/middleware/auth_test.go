package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func executeRoleMiddlewareRequest(t *testing.T, role any, middlewareFn gin.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	router := gin.New()

	router.GET("/protected",
		func(c *gin.Context) {
			if role != nil {
				c.Set("user_role", role)
			}
			c.Next()
		},
		middlewareFn,
		func(c *gin.Context) {
			c.Status(http.StatusOK)
		},
	)

	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	router.ServeHTTP(recorder, request)

	return recorder
}

func TestAdminMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		role       any
		expectCode int
	}{
		{name: "admin passes", role: "admin", expectCode: http.StatusOK},
		{name: "user blocked", role: "user", expectCode: http.StatusForbidden},
		{name: "missing role blocked", role: nil, expectCode: http.StatusForbidden},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			response := executeRoleMiddlewareRequest(t, testCase.role, AdminMiddleware())
			if response.Code != testCase.expectCode {
				t.Fatalf("expected status %d, got %d", testCase.expectCode, response.Code)
			}
		})
	}
}

func TestUserMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		role       any
		expectCode int
	}{
		{name: "user passes", role: "user", expectCode: http.StatusOK},
		{name: "mitra blocked", role: "mitra", expectCode: http.StatusForbidden},
		{name: "admin blocked", role: "admin", expectCode: http.StatusForbidden},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			response := executeRoleMiddlewareRequest(t, testCase.role, UserMiddleware())
			if response.Code != testCase.expectCode {
				t.Fatalf("expected status %d, got %d", testCase.expectCode, response.Code)
			}
		})
	}
}

func TestMemberMiddlewareAlias(t *testing.T) {
	tests := []struct {
		name       string
		role       any
		expectCode int
	}{
		{name: "user passes via alias", role: "user", expectCode: http.StatusOK},
		{name: "legacy member is blocked", role: "member", expectCode: http.StatusForbidden},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			response := executeRoleMiddlewareRequest(t, testCase.role, MemberMiddleware())
			if response.Code != testCase.expectCode {
				t.Fatalf("expected status %d, got %d", testCase.expectCode, response.Code)
			}
		})
	}
}

func TestMitraMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		role       any
		expectCode int
	}{
		{name: "mitra passes", role: "mitra", expectCode: http.StatusOK},
		{name: "user blocked", role: "user", expectCode: http.StatusForbidden},
		{name: "missing role blocked", role: nil, expectCode: http.StatusForbidden},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			response := executeRoleMiddlewareRequest(t, testCase.role, MitraMiddleware())
			if response.Code != testCase.expectCode {
				t.Fatalf("expected status %d, got %d", testCase.expectCode, response.Code)
			}
		})
	}
}

func TestGetUserRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns role when valid", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		context, _ := gin.CreateTestContext(recorder)
		context.Set("user_role", "mitra")

		role, ok := GetUserRole(context)
		if !ok {
			t.Fatalf("expected GetUserRole to return ok=true")
		}
		if role != "mitra" {
			t.Fatalf("expected role mitra, got %s", role)
		}
	})

	t.Run("returns false when missing", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		context, _ := gin.CreateTestContext(recorder)

		_, ok := GetUserRole(context)
		if ok {
			t.Fatalf("expected GetUserRole to return ok=false when role missing")
		}
	})

	t.Run("returns false when invalid type", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		context, _ := gin.CreateTestContext(recorder)
		context.Set("user_role", 123)

		_, ok := GetUserRole(context)
		if ok {
			t.Fatalf("expected GetUserRole to return ok=false when role type is invalid")
		}
	})
}
