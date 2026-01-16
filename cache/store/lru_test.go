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
	lru := NewLruCache(100)
	lru.Add("key1", String("1234"))
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
	lru := NewLruCache(cap)
	lru.Add(k1, v1)
	lru.Add(k2, v2)
	lru.Add(k3, v3) // 此时总大小 12 > 10，应该淘汰 k1
	if _, ok := lru.Get(k1); ok {
		t.Fatalf("Removeoldest key1 failed")
	}
	if _, ok := lru.Get(k2); !ok {
		t.Fatalf("cache miss key2 failed")
	}
	lru.Get(k2)
	lru.Add(k4, v4) // 此时应该淘汰 k3，因为 k2 刚被访问过
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
	lru := NewLruCache(int64(4)) 
	lru.OnEvicted = callback
	lru.Add("k1", String("v1")) // size 4
	lru.Add("k2", String("v2")) // size 4 + 4 > 4，淘汰 k1
	lru.Delete("k2")            // 主动删除 k2
	expected := []string{"k1", "k2"}
	if len(keys) != 2 {
		t.Fatalf("Call OnEvicted failed, expect len:2, got %d", len(keys))
	}
	if keys[0] != expected[0] || keys[1] != expected[1] {
		t.Fatalf("Call OnEvicted failed, expect %s, got %s", expected, keys)
	}
}

func TestLRU_AddWithExpiration(t *testing.T) {
	lru := NewLruCache(100)
	lru.AddWithExpiration("k1", String("v1"), time.Now().Add(100*time.Millisecond))
	if _, ok := lru.Get("k1"); !ok {
		t.Fatal("key1 should exist immediately")
	}
	time.Sleep(200 * time.Millisecond)
	if _, ok := lru.Get("k1"); ok {
		t.Fatal("key1 should be expired")
	}
	if lru.nbytes != 0 {
		t.Fatalf("nbytes should be 0 after expiration, got %d", lru.nbytes)
	}
}

func TestLRU_Delete(t *testing.T) {
	lru := NewLruCache(100)
	lru.Add("k1", String("v1"))
	if !lru.Delete("k1") {
		t.Fatal("Delete return false")
	}
	if _, ok := lru.Get("k1"); ok {
		t.Fatal("key1 should be deleted")
	}
	if lru.nbytes != 0 {
		t.Fatalf("nbytes should be 0, got %d", lru.nbytes)
	}
}

func TestLRU_Update(t *testing.T) {
	lru := NewLruCache(100)
	lru.Add("k1", String("1")) // size 2+1=3
	if v, _ := lru.Get("k1"); string(v.(String)) != "1" {
		t.Fatal("val check failed")
	}
	lru.Add("k1", String("123")) // size 2+3=5 (增加2字节)
	if v, _ := lru.Get("k1"); string(v.(String)) != "123" {
		t.Fatal("val update failed")
	}
	if lru.nbytes != 5 {
		t.Fatalf("nbytes update failed, expect 5, got %d", lru.nbytes)
	}
}