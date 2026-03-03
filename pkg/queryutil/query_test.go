package queryutil

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func makeContext(rawURL string) *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", rawURL, nil)
	c.Request = req
	return c
}

func TestGetPagination(t *testing.T) {
	c := makeContext("/items?page=0&limit=1000")
	p := GetPagination(c)

	if p.Page != 1 {
		t.Fatalf("expected page=1, got %d", p.Page)
	}
	if p.Limit != MaxLimit {
		t.Fatalf("expected limit=%d, got %d", MaxLimit, p.Limit)
	}
	if p.Offset != 0 {
		t.Fatalf("expected offset=0, got %d", p.Offset)
	}
}

func TestGetPaginationWithOffset(t *testing.T) {
	c := makeContext("/items?limit=10&offset=30")
	p := GetPagination(c)

	if p.Page != 4 {
		t.Fatalf("expected page=4, got %d", p.Page)
	}
	if p.Offset != 30 {
		t.Fatalf("expected offset=30, got %d", p.Offset)
	}
}

func TestOptionalAndParamHelpers(t *testing.T) {
	c := makeContext("/items?uid=7&enabled=true&name=abc")
	c.Params = gin.Params{{Key: "id", Value: "42"}}

	if v := GetUintQuery(c, "uid", 0); v != 7 {
		t.Fatalf("unexpected uint query: %d", v)
	}
	if v := GetBoolQuery(c, "enabled", false); !v {
		t.Fatal("expected enabled=true")
	}
	if v := GetStringQuery(c, "name", "x"); v != "abc" {
		t.Fatalf("unexpected name: %s", v)
	}
	if v := GetOptionalUint(c, "uid"); v == nil || *v != 7 {
		t.Fatalf("unexpected optional uint: %v", v)
	}
	if v := GetOptionalString(c, "name"); v == nil || *v != "abc" {
		t.Fatalf("unexpected optional string: %v", v)
	}

	if id, ok := GetIntParam(c, "id"); !ok || id != 42 {
		t.Fatalf("unexpected int param: %d, %v", id, ok)
	}
	if id, ok := GetUintParam(c, "id"); !ok || id != 42 {
		t.Fatalf("unexpected uint param: %d, %v", id, ok)
	}
}

func TestGetSort(t *testing.T) {
	allowed := []string{"created_at", "title"}

	c := makeContext("/items?sort=title&order=desc")
	s := GetSort(c, "created_at", "asc", allowed)
	if s.Field != "title" || s.Order != "desc" {
		t.Fatalf("unexpected sort: %+v", s)
	}

	c2 := makeContext("/items?sort=hack&order=invalid")
	s2 := GetSort(c2, "created_at", "asc", allowed)
	if s2.Field != "created_at" || s2.Order != "asc" {
		t.Fatalf("expected defaults, got %+v", s2)
	}
}

func TestAdditionalQueryHelpers(t *testing.T) {
	c := makeContext("/items?count64=123456789&opt_int=15&bad_int=oops")
	c.Params = gin.Params{{Key: "id", Value: "not-a-number"}}

	if v := GetInt64Query(c, "count64", 9); v != 123456789 {
		t.Fatalf("unexpected int64 query: %d", v)
	}
	if v := GetInt64Query(c, "missing", 9); v != 9 {
		t.Fatalf("expected int64 default fallback, got %d", v)
	}

	if v := GetOptionalInt(c, "opt_int"); v == nil || *v != 15 {
		t.Fatalf("unexpected optional int: %v", v)
	}
	if v := GetOptionalInt(c, "bad_int"); v != nil {
		t.Fatalf("expected nil optional int for invalid value, got %v", v)
	}
	if v := GetOptionalInt(c, "none"); v != nil {
		t.Fatalf("expected nil optional int for missing value, got %v", v)
	}

	if id := MustGetUintParam(c, "id"); id != 0 {
		t.Fatalf("expected MustGetUintParam fallback 0, got %d", id)
	}
	if id := MustGetIntParam(c, "id"); id != 0 {
		t.Fatalf("expected MustGetIntParam fallback 0, got %d", id)
	}
}

func TestQueryFallbackBranches(t *testing.T) {
	c := makeContext("/items?uid=bad&enabled=badbool")

	if v := GetUintQuery(c, "uid", 55); v != 55 {
		t.Fatalf("expected uint fallback 55, got %d", v)
	}
	if v := GetUintQuery(c, "missing_uid", 11); v != 11 {
		t.Fatalf("expected missing uint fallback 11, got %d", v)
	}

	if v := GetBoolQuery(c, "enabled", true); v != true {
		t.Fatalf("expected bool fallback true, got %v", v)
	}
	if v := GetBoolQuery(c, "missing_enabled", false); v != false {
		t.Fatalf("expected missing bool fallback false, got %v", v)
	}

	if v := GetOptionalUint(c, "uid"); v != nil {
		t.Fatalf("expected nil optional uint for invalid value, got %v", v)
	}
	if v := GetOptionalUint(c, "missing_uid"); v != nil {
		t.Fatalf("expected nil optional uint for missing value, got %v", v)
	}

	c2 := makeContext("/items")
	if id, ok := GetIntParam(c2, "id"); ok || id != 0 {
		t.Fatalf("expected empty int param to fail, got id=%d ok=%v", id, ok)
	}
	if id, ok := GetUintParam(c2, "id"); ok || id != 0 {
		t.Fatalf("expected empty uint param to fail, got id=%d ok=%v", id, ok)
	}
}

func TestMoreDefaultAndInvalidBranches(t *testing.T) {
	c := makeContext("/items?page=2&limit=0&offset=4&count64=bad")

	p := GetPagination(c)
	if p.Limit != DefaultLimit {
		t.Fatalf("expected limit fallback=%d, got %d", DefaultLimit, p.Limit)
	}
	if p.Page != 1 {
		t.Fatalf("expected page recalculated to 1, got %d", p.Page)
	}

	if v := GetIntQuery(c, "bad", 77); v != 77 {
		t.Fatalf("expected GetIntQuery fallback 77, got %d", v)
	}
	if v := GetInt64Query(c, "count64", 88); v != 88 {
		t.Fatalf("expected GetInt64Query fallback 88, got %d", v)
	}
	if v := GetStringQuery(c, "missing", "def"); v != "def" {
		t.Fatalf("expected GetStringQuery default def, got %q", v)
	}
	if v := GetOptionalString(c, "missing"); v != nil {
		t.Fatalf("expected nil optional string, got %v", v)
	}
}
