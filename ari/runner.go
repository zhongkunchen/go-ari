package ari

import (
	"sync"
	"fmt"
)

type Runner interface {
	Run() error
}

type RunnerBuilder interface {
	Build(ctx *Context, cfg map[string]interface{}, group string) Runner
}

type RunnerRegistry struct {
	lock sync.RWMutex
	registry map[string]RunnerBuilder
	group string
}

func newRunnerRegistry(group string) *RunnerRegistry {
	p := &RunnerRegistry{
		registry:map[string]RunnerBuilder{},
		group:group,
	}
	return p
}

// Register a plugin runner with name
func (r *RunnerRegistry) Register(name string, runner RunnerBuilder) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	if _, exists := r.registry[name]; exists {
		return fmt.Errorf("%s plugin %s exists", r.group, name)
	}
	r.registry[name] = runner
	return nil
}

func (r *RunnerRegistry) get(name string) (builder RunnerBuilder)  {
	r.lock.RLock()
	defer r.lock.RUnlock()
	builder, ok := r.registry[name]
	if !ok {
		builder = nil
	}
	return builder
}

var (
	InputRunnerBuilders = newRunnerRegistry("input")
)
