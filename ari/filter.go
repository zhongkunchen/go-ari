package ari

import (
	"sync"
	"fmt"
)

type Filter interface {
	DoFilter(msg *Message) bool
}

type FilterRegistry struct {
	lock sync.RWMutex
	registry map[string]Filter
}

func newFilterRegistry() *FilterRegistry {
	return &FilterRegistry{
		registry:map[string]Filter{},
	}
}

func (r *FilterRegistry) Register(name string, filter Filter) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	if _, exists := r.registry[name]; exists {
		return fmt.Errorf("filter %s exists", name)
	}
	r.registry[name] = filter
	return nil
}

func (r *FilterRegistry) get(name string) Filter {
	r.lock.RLock()
	defer r.lock.RUnlock()
	if fi, exists := r.registry[name]; exists {
		return fi
	}
	return nil
}

var (
	Filters = newFilterRegistry()
)

