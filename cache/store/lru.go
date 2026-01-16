package store

import (
	"container/list"
	"sync"
	"time"
)

type Store interface {
	Get(key string) (Value, bool)
	Add(key string, value Value)
	AddWithExpiration(key string, value Value, expirationTime time.Time)
	Delete(key string) bool
}

type Value interface {
	Len() int
}

// lruCache 结构体
type lruCache struct {
	lock      sync.Mutex
	cacheMap  map[string]*list.Element
	expires   map[string]time.Time         
	ll        *list.List                
	OnEvicted func(key string, value Value)
	maxBytes  int64                      
	nbytes    int64                      
}

// 通过key可以在记录删除时，删除字典缓存中的映射
type entry struct {
	key   string
	value Value
}

func NewLruCache(maxSize int64) *lruCache {
	lc := lruCache{
		cacheMap: make(map[string]*list.Element),
		expires: make(map[string]time.Time),
		nbytes: 0,
		ll:		list.New(),
		maxBytes: maxSize,
	}
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			lc.periodicMemoryClean()	
		}
	}()
	return &lc
}

func (c *lruCache) Get(key string) (Value, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	// 惰性删除检查
	if epTime, ok := c.expires[key]; ok && epTime.Before(time.Now()) {
		c.rmElement(c.cacheMap[key])
		return nil, false
	}
	// 过期清除
	if v, ok := c.cacheMap[key]; ok {
		c.ll.MoveToBack(v)
		return v.Value.(*entry).value, true
	}
	return nil, false
}

// 添加或更新一个永不过期的键值对
func (c *lruCache) Add(key string, value Value) {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.expires, key)
	c.baseAdd(key, value)
	c.freeIfNeeded()
}

// 辅助函数
func (c *lruCache) baseAdd(key string, value Value) {
	if ele, ok := c.cacheMap[key]; ok {
		c.ll.MoveToBack(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.ll.PushBack(&entry{key, value})
		c.cacheMap[key] = ele
		c.nbytes += int64(len(key) + value.Len())
	}
}

// 添加或更新一个带过期时间的键值对
func (c *lruCache) AddWithExpiration(key string, value Value, expirationTime time.Time) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.expires[key] = expirationTime
	c.baseAdd(key, value)
	c.freeIfNeeded()
}

// 删除一个键值对
func (c *lruCache) Delete(key string) bool {
	c.lock.Lock()
	defer c.lock.Unlock()
	if ele, ok := c.cacheMap[key]; ok {
		c.rmElement(ele)
		return true
	}
	return true
}

// 辅助方法
func (c *lruCache) rmElement(e *list.Element) {
	if e == nil {
		return
	}
	c.ll.Remove(e)
	kv := e.Value.(*entry)
	delete(c.cacheMap, kv.key)
	delete(c.expires, kv.key)
	c.nbytes -= int64(len(kv.key) + kv.value.Len())

	if c.OnEvicted != nil {
		c.OnEvicted(kv.key, kv.value)
	}
}

// 内存淘汰策略 无锁
func (c *lruCache) freeIfNeeded() {
	for c.maxBytes > 0 && c.nbytes > c.maxBytes {
		ele := c.ll.Front()
		if ele != nil {
			c.rmElement(ele)
		}
	}
}

// 主动过期清理
func (c *lruCache) periodicMemoryClean() {
	c.lock.Lock()
	defer c.lock.Unlock()

	checked := 0
	limit := len(c.expires) / 10
	if limit < 10 {
		limit = 10
	}
	for key, t := range c.expires {
		if checked >= limit {
			break
		}
		checked++
		if t.Before(time.Now()) {
			if ele, ok := c.cacheMap[key]; ok {
				c.rmElement(ele)
			}
		}
	}
}
