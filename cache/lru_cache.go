package cache

import (
	"container/list"
	"sync"
	"time"
)

// consts
const (
	DefaultCacheTime = time.Minute
)

// OnEvicted callback func will be called when the cached key expired
type OnEvicted func(key Key, value Value)

type lruCache struct {
	maxLen    int
	onEvicted OnEvicted
	lst       *list.List
	hash      map[Key]*list.Element
	cacheTime time.Duration
	sync.Mutex
}

type listEntry struct {
	key      Key
	value    Value
	deadTime time.Time
}

// Config of the cache
type Config struct {
	MaxLen    int
	Callback  OnEvicted
	CacheTime time.Duration
}

// NewCache will create a default configured cache
func NewCache() Cache {
	return NewCacheWithConfig(Config{MaxLen: DefaultMaxLen, CacheTime: DefaultCacheTime})
}

// NewCacheWithConfig will create a cache with the configs
func NewCacheWithConfig(config Config) Cache {
	if config.CacheTime < time.Millisecond {
		config.CacheTime = DefaultCacheTime
	}
	return &lruCache{
		maxLen:    config.MaxLen,
		onEvicted: config.Callback,
		lst:       &list.List{},
		hash:      map[Key]*list.Element{},
		cacheTime: config.CacheTime,
	}
}

func (lru *lruCache) removeElem(elem *list.Element) {
	if elem == nil {
		return
	}
	lru.lst.Remove(elem)

	entry := elem.Value.(*listEntry)
	delete(lru.hash, entry.key)
	if lru.onEvicted != nil {
		lru.onEvicted(entry.key, entry.value)
	}
}

func (lru *lruCache) lazyRemoveOldest() {
	if len(lru.hash) > lru.maxLen {
		lru.removeElem(lru.lst.Back())
	}
}

func (lru *lruCache) Put(key Key, value Value) {
	lru.PutWithTimeout(key, value, lru.cacheTime)
}

func (lru *lruCache) PutWithTimeout(key Key, value Value, t time.Duration) {
	if t < time.Second {
		t = time.Second
	}
	lru.Lock()
	defer lru.Unlock()
	if elem, exists := lru.hash[key]; exists {
		lru.lst.MoveToFront(elem)
		elem.Value.(*listEntry).value = value
		elem.Value.(*listEntry).deadTime = time.Now().Add(t)
	} else {
		lru.hash[key] = lru.lst.PushFront(&listEntry{key: key, value: value, deadTime: time.Now().Add(t)})
		lru.lazyRemoveOldest()
	}
}

func (lru *lruCache) Get(key Key) (Value, bool) {
	lru.Lock()
	defer lru.Unlock()
	if elem, exists := lru.hash[key]; exists {
		entry := elem.Value.(*listEntry)
		// delete the cached value if it has already timeouted
		if entry.deadTime.Before(time.Now()) {
			lru.removeElem(elem)
			return nil, false
		}
		lru.lst.MoveToFront(elem)
		return elem.Value.(*listEntry).value, true
	}
	return nil, false

}
func (lru *lruCache) Del(key Key) Value {
	lru.Lock()
	defer lru.Unlock()
	if elem, exists := lru.hash[key]; exists {
		value := elem.Value.(*listEntry).value
		lru.removeElem(elem)
		return value
	}
	return nil
}
func (lru *lruCache) Len() int {
	lru.Lock()
	defer lru.Unlock()
	if lru.hash == nil {
		return 0
	}
	return len(lru.hash)
}
func (lru *lruCache) Close() {
	lru.Lock()
	defer lru.Unlock()
	lru.hash = map[Key]*list.Element{}
	lru.lst.Init()
}
