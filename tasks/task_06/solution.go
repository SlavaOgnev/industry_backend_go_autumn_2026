package main

import "container/list"

type entry[K comparable, V any] struct {
	key   K
	value V
}
type LRUCache[K comparable, V any] struct {
	capacity int
	ll       list.List
	items    map[K]*list.Element
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	return &LRUCache[K, V]{capacity: capacity, ll: list.List{}, items: make(map[K]*list.Element)}
}
func (c *LRUCache[K, V]) Get(key K) (value V, ok bool) {

	elem, ok := c.items[key]
	if ok {
		c.ll.MoveToFront(elem)
		return elem.Value.(*entry[K, V]).value, true
	}
	return value, false
}
func (c *LRUCache[K, V]) Set(key K, value V) {
	if c.capacity <= 0 {
		return
	}
	if elem, ok := c.items[key]; ok {
		elem.Value.(*entry[K, V]).value = value
		return
	} else if c.capacity == len(c.items) {
		delete(c.items, c.ll.Back().Value.(*entry[K, V]).key)
		c.ll.Remove(c.ll.Back())
	}
	c.items[key] = c.ll.PushFront(&entry[K, V]{key: key, value: value})

}

type LRU[K comparable, V any] interface {
	Get(K) (V, bool)
	Set(K, V)
}
