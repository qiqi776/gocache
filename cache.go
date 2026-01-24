package gocache

import (
	"gocache/store"
	"sync"
	"time"
)

type cache struct {
	lock       sync.RWMutex
	lruCache   store.Store
	cacheBytes int64
}

func (cache *cache) lruCacheLazyLoadIfNeed() {
	if cache.lruCache == nil {
		cache.lock.Lock()
		defer cache.lock.Unlock()
		if cache.lruCache == nil {
			cache.lruCache = store.NewLRUCache(store.Options{
				MaxBytes: cache.cacheBytes,
			})
		}
	}
}

func (cache *cache) add(key string, value ByteView) {
	cache.lruCacheLazyLoadIfNeed()
	cache.lruCache.Set(key, value)
}

func (cache *cache) get(key string) (value ByteView, ok bool) {
	cache.lock.RLock()
	defer cache.lock.RUnlock()
	if cache.lruCache == nil {
		return
	}
	if v, find := cache.lruCache.Get(key); find {
		return v.(ByteView), true
	}
	return
}

func (cache *cache) addWithExpiration(key string, value ByteView, expirationTime time.Time) {
	cache.lruCacheLazyLoadIfNeed()
    ttl := time.Until(expirationTime)
    if ttl < 0 {
        ttl = 0
    }
	cache.lruCache.SetWithExpiration(key, value, ttl)
}

func (cache *cache) delete(key string) bool {
	cache.lock.Lock()
	defer cache.lock.Unlock()
	if cache.lruCache == nil {
		return true
	}
	return cache.lruCache.Delete(key)
}