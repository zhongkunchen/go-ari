package ari

import (
	"sync"
	"fmt"
)

type Sender interface {
	Send(msg *Message)
}

type SenderBuilder interface {
	Build(ctx *Context, cfg map[string]interface{}) (Sender, error)
}

type senderRegistry struct {
	lock sync.RWMutex
	registry map[string] SenderBuilder
}

func newSenderRegistry() *senderRegistry {
	s := &senderRegistry{
		registry:map[string] SenderBuilder{},
	}
	return s
}

func (r *senderRegistry) Register(name string, builder SenderBuilder) error  {
	r.lock.Lock()
	defer r.lock.Unlock()
	if _, exists := r.registry[name]; exists {
		return fmt.Errorf("sender %s exists", name)
	}
	r.registry[name] = builder
	return nil
}

func (r *senderRegistry) get(name string) (SenderBuilder) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	if builder, exists := r.registry[name]; exists {
		return builder
	}
	return nil
}

var (
	SenderBulders = newSenderRegistry()
)
