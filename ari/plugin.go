package ari

import (
	"sync"
	"fmt"
)

type Runner interface {
	Run() error
}

type RunnerBuilder interface {
	Build(ctx *Context, cfg map[string]interface{}) Runner
}

type PluginRegistry struct {
	lock sync.RWMutex
	registry map[string]RunnerBuilder
	group string
}

func newPluginRegistry(group string) *PluginRegistry {
	p := &PluginRegistry{
		registry:map[string]RunnerBuilder{},
		group:group,
	}
	return p
}

// Register a plugin runner with name
func (r *PluginRegistry) Register(name string, runner RunnerBuilder) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	if _, exists := r.registry[name]; exists {
		return fmt.Errorf("%s plugin %s exists", r.group, name)
	}
	r.registry[name] = runner
	return nil
}

func (r *PluginRegistry) get(name string) (builder RunnerBuilder)  {
	r.lock.RLock()
	defer r.lock.RUnlock()
	builder, ok := r.registry[name]
	if !ok {
		builder = nil
	}
	return builder
}

var (
	InputPlugins = newPluginRegistry("input")
	OutputPlugins = newPluginRegistry("output")
	FilterPlugins = newPluginRegistry("filter")
)
