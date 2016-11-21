package ari

import (
	"sync"
	"fmt"
)

type Filter interface {
	DoFilter(msg *Message) bool
}

type FilterBuilder interface {
	Build(ctx *Context, cfg map[string]interface{}) (Filter, error)
}

type FilterRegistry struct {
	lock sync.RWMutex
	registry map[string] FilterBuilder
}

func newFilterRegistry() *FilterRegistry {
	return &FilterRegistry{
		registry:map[string]FilterBuilder{},
	}
}

func (r *FilterRegistry) Register(name string, builder FilterBuilder) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	if _, exists := r.registry[name]; exists {
		return fmt.Errorf("filter %s exists", name)
	}
	r.registry[name] = builder
	return nil
}

func (r *FilterRegistry) get(name string) FilterBuilder {
	r.lock.RLock()
	defer r.lock.RUnlock()
	if fi, exists := r.registry[name]; exists {
		return fi
	}
	return nil
}

var (
	FilterBuilders = newFilterRegistry()
)

