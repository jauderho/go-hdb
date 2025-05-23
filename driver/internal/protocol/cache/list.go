// Package cache provides generic cache types.
package cache

import (
	"sync"
)

// Comparer is an interface defining a generic compare function.
type Comparer[E any] interface {
	Compare(e E) bool
}

// List is a generic cache list.
type List[K Comparer[K], V any] struct {
	valueFn func(k K) (V, error)
	mu      sync.RWMutex
	idx     int
	keys    []K
	values  []V
}

// NewList returns a new cache list.
func NewList[K Comparer[K], V any](maxEntry int, valueFn func(k K) (V, error)) *List[K, V] {
	return &List[K, V]{
		valueFn: valueFn,
		keys:    make([]K, 0, maxEntry),
		values:  make([]V, 0, maxEntry),
	}
}

func (l *List[K, V]) find(k K) (v V, ok bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	for i, k1 := range l.keys {
		if k1.Compare(k) {
			return l.values[i], true
		}
	}
	return
}

// Get returns the value for the given key.
func (l *List[K, V]) Get(k K) (V, error) {
	if v, ok := l.find(k); ok {
		return v, nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	v, err := l.valueFn(k)
	if err != nil {
		return v, err
	}
	if l.idx < len(l.keys) {
		l.keys[l.idx], l.values[l.idx] = k, v
	} else {
		l.keys = append(l.keys, k)
		l.values = append(l.values, v)
	}
	l.idx++
	if l.idx >= cap(l.keys) {
		l.idx = 0
	}
	return v, nil
}
