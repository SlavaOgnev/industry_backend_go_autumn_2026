package main

import (
	"container/list"
	"sync"
)

type LRU[K comparable, V any] interface {
	Get(key K) (value V, ok bool)
	Set(key K, value V)
}
type entry[K comparable, V any] struct {
	key   K
	value V
}
type LRUCache[K comparable, V any] struct {
	capacity int
	mu       sync.Mutex
	ll       list.List
	items    map[K]*list.Element
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	return &LRUCache[K, V]{capacity: capacity, mu: sync.Mutex{}, ll: list.List{}, items: make(map[K]*list.Element)}
}
func (c *LRUCache[K, V]) Get(key K) (value V, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if v, ok := c.items[key]; ok {
		c.ll.MoveToFront(v)
		return v.Value.(*entry[K, V]).value, ok
	}
	return value, ok

}
func (c *LRUCache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.capacity <= 0 {
		return
	}

	if _, ok := c.items[key]; ok {
		c.items[key].Value.(*entry[K, V]).value = value
		return
	}
	if len(c.items) == c.capacity {
		eldestelem := c.ll.Back()
		delete(c.items, eldestelem.Value.(*entry[K, V]).key)
		c.ll.Remove(eldestelem)
	}
	newElem := c.ll.PushFront(&entry[K, V]{key: key, value: value})
	c.items[key] = newElem

}
