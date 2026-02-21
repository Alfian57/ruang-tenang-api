package service

import (
	"testing"
	"time"
)

func TestCacheService_SetGetDelete(t *testing.T) {
	c := NewCacheService()
	c.Set("k1", "v1")

	if got := c.Get("k1"); got != "v1" {
		t.Fatalf("expected v1, got %v", got)
	}

	c.Delete("k1")
	if got := c.Get("k1"); got != nil {
		t.Fatalf("expected nil after delete, got %v", got)
	}
}

func TestCacheService_ExpiryPrefixClear(t *testing.T) {
	c := NewCacheService()
	c.SetWithTTL("pref:1", "a", 5*time.Millisecond)
	c.SetWithTTL("pref:2", "b", time.Minute)
	c.SetWithTTL("other:1", "c", time.Minute)

	time.Sleep(10 * time.Millisecond)
	if got := c.Get("pref:1"); got != nil {
		t.Fatalf("expected expired entry to be nil, got %v", got)
	}

	c.DeletePrefix("pref:")
	if c.Get("pref:2") != nil {
		t.Fatal("expected pref:2 to be deleted by prefix")
	}
	if c.Get("other:1") == nil {
		t.Fatal("expected other:1 to remain")
	}

	c.Clear()
	if c.Get("other:1") != nil {
		t.Fatal("expected cache clear to remove all entries")
	}
}
