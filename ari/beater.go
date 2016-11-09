package ari

import (
	"sync"
	"fmt"
)

var (
	beaterCreators = map[string]BeaterCreator{}
	beaterLock = &sync.RWMutex{}
	// reserved names
	reservedNames = map[string]interface{}{"":nil,}
)

// RegisterBeaterCreator registers a beater creator with `name`
func RegisterBeaterCreator(name string, creator BeaterCreator) (error) {
	beaterLock.Lock()
	defer beaterLock.Unlock()
	if _, denied :=reservedNames[name]; denied {
		return fmt.Errorf("%s is reserved", name)
	}
	if _, ok := beaterCreators[name]; ok{
		return fmt.Errorf("beater %s exists", name)
	}
	beaterCreators[name] = creator
	return nil
}

// GetBeaterCreator returns creator named `name`,
// returns nil if not exist
func GetBeaterCreator(name string) BeaterCreator {
	beaterLock.RLock()
	defer beaterLock.RUnlock()
	if c, ok := beaterCreators[name]; ok {
		return c
	}
	return nil
}

// BeaterCreator serves methods to configure a `Beater`
type BeaterCreator interface {
	// Create method returns a beater instance
	// the arg `messageChan` is the chan to put log messages
	Create(conf Configuration, messageChan chan <- *Message) (Beater, error)
}

// Beater produces log messages, and sends them to the `Ari`
type Beater interface {
	Beating()
	Stop()
}
