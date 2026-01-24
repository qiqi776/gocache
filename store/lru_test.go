package store

import (
	"testing"
	"time"
)

type String string

func (d String) Len() int {
	return len(d)
}

func TestLRU_AddGet(t *testing.T) {
	lru := NewLRUCache(Options{MaxBytes: 100})
	lru.Set("key1", String("1234"))
	if v, ok := lru.Get("key1"); !ok || string(v.(String)) != "1234" {
		t.Fatalf("cache hit key1=1234 failed")
	}
	if _, ok := lru.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}

func TestLRU_Eviction(t *testing.T) {
	k1, k2, k3, k4 := "k1", "k2", "k3", "k4"
	v1, v2, v3, v4 := String("v1"), String("v2"), String("v3"), String("v4")
	cap := int64(10)
	lru := NewLRUCache(Options{MaxBytes: cap})
	lru.Set(k1, v1)
	lru.Set(k2, v2)
	lru.Set(k3, v3) 
	if _, ok := lru.Get(k1); ok {
		t.Fatalf("Removeoldest key1 failed")
	}
	if _, ok := lru.Get(k2); !ok {
		t.Fatalf("cache miss key2 failed")
	}
	lru.Get(k2)
	lru.Set(k4, v4) 
	if _, ok := lru.Get(k3); ok {
		t.Fatalf("Removeoldest key3 failed")
	}
	if _, ok := lru.Get(k2); !ok {
		t.Fatalf("cache miss key2 failed")
	}
}

func TestLRU_OnEvicted(t *testing.T) {
	keys := make([]string, 0)
	callback := func(key string, value Value) {
		keys = append(keys, key)
	}
	lru := NewLRUCache(Options{
		MaxBytes:  int64(4),
		OnEvicted: callback,
	})
	lru.Set("k1", String("v1")) 
	lru.Set("k2", String("v2")) 
	lru.Delete("k2")            
	expected := []string{"k1", "k2"}
	if len(keys) != 2 {
		t.Fatalf("Call OnEvicted failed, expect len:2, got %d", len(keys))
	}
	if keys[0] != expected[0] || keys[1] != expected[1] {
		t.Fatalf("Call OnEvicted failed, expect %s, got %s", expected, keys)
	}
}

func TestLRU_AddWithExpiration(t *testing.T) {
	lru := NewLRUCache(Options{MaxBytes: 100})
	lru.SetWithExpiration("k1", String("v1"), 100*time.Millisecond)
	if _, ok := lru.Get("k1"); !ok {
		t.Fatal("key1 should exist immediately")
	}
	time.Sleep(200 * time.Millisecond)
	if _, ok := lru.Get("k1"); ok {
		t.Fatal("key1 should be expired")
	}
	lru.mu.Lock()
	lru.evict()
	lru.mu.Unlock()
	if lru.usedBytes != 0 {
		t.Fatalf("usedBytes should be 0 after expiration, got %d", lru.usedBytes)
	}
}

func TestLRU_Delete(t *testing.T) {
	lru := NewLRUCache(Options{MaxBytes: 100})
	lru.Set("k1", String("v1"))
	if !lru.Delete("k1") {
		t.Fatal("Delete return false")
	}
	if _, ok := lru.Get("k1"); ok {
		t.Fatal("key1 should be deleted")
	}
	if lru.usedBytes != 0 {
		t.Fatalf("usedBytes should be 0, got %d", lru.usedBytes)
	}
}

func TestLRU_Update(t *testing.T) {
	lru := NewLRUCache(Options{MaxBytes: 100})
	lru.Set("k1", String("1"))
	if v, _ := lru.Get("k1"); string(v.(String)) != "1" {
		t.Fatal("val check failed")
	}
	lru.Set("k1", String("123"))
	if v, _ := lru.Get("k1"); string(v.(String)) != "123" {
		t.Fatal("val update failed")
	}
	if lru.usedBytes != 5 {
		t.Fatalf("usedBytes update failed, expect 5, got %d", lru.usedBytes)
	}
}