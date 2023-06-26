package cache

import (
	"container/list"
	"sync"
	"time"
)

// Cache 根据key索引的cache
type Cache struct {
	timeout int64
	vars    map[string]*cacheEntry
	ll      *list.List
	mu      sync.RWMutex
}

type cacheEntry struct {
	key  string
	last int64
	val  interface{}
	le   *list.Element
}

// NewCache 生成一个lruCache，超时单位是秒
func NewCache(timeout int64) *Cache {
	return &Cache{timeout: timeout, vars: make(map[string]*cacheEntry), ll: list.New()}
}

// Get get from cache and remove expired key.
func (c *Cache) Get(key string) interface{} {
	c.evict()

	c.mu.RLock()
	if v, ok := c.vars[key]; ok {
		c.mu.RUnlock()
		return v.val
	}
	c.mu.RUnlock()
	return nil
}

// Add add to cache and remove expired key.
func (c *Cache) Add(key string, val interface{}) {
	c.mu.Lock()
	if v, ok := c.vars[key]; ok {
		v.val = val
		c.mu.Unlock()
		return
	}

	v := &cacheEntry{key: key, val: val, last: time.Now().Unix()}
	v.le = c.ll.PushFront(v)
	c.vars[key] = v
	c.mu.Unlock()

}

func (c *Cache) evict() {
	last := time.Now().Unix() - c.timeout
	c.mu.Lock()
	for b := c.ll.Back(); b != nil; b = c.ll.Back() {
		e := b.Value.(*cacheEntry)
		if last < e.last {
			break
		}
		c.ll.Remove(b)
		delete(c.vars, e.key)
	}
	c.mu.Unlock()
}
