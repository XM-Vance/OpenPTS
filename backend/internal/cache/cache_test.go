package cache

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	c := New("test-smoke")
	if c == nil {
		t.Fatal("New() returned nil")
	}

	// Basic Set/Get cycle.
	c.Set("key1", "value1", 5*time.Second)
	v, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected cache hit for key1")
	}
	if v != "value1" {
		t.Fatalf("expected value1, got %v", v)
	}

	// Miss.
	_, ok = c.Get("nonexistent")
	if ok {
		t.Fatal("expected cache miss for nonexistent key")
	}
}
