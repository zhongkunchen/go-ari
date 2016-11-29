package ari

import (
	"sync"
	"fmt"
)

type Beater interface {
	Run() error
}

type BeaterBuilder interface {
	Build(ctx *Context, cfg map[string]interface{}, group string) Beater
}

type BeaterRegistry struct {
	lock sync.RWMutex
	registry map[string]BeaterBuilder
	group string
}

func newBeaterRegistry(group string) *BeaterRegistry {
	p := &BeaterRegistry{
		registry:map[string]BeaterBuilder{},
		group:group,
	}
	return p
}

// Register a plugin runner with name
func (r *BeaterRegistry) Register(name string, beater BeaterBuilder) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	if _, exists := r.registry[name]; exists {
		return fmt.Errorf("%s plugin %s exists", r.group, name)
	}
	r.registry[name] = beater
	return nil
}

func (r *BeaterRegistry) get(name string) (builder BeaterBuilder)  {
	r.lock.RLock()
	defer r.lock.RUnlock()
	builder, ok := r.registry[name]
	if !ok {
		builder = nil
	}
	return builder
}

var (
	BeaterBuilders = newBeaterRegistry("input")
)
