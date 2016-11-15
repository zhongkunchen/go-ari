package ari

import (
	"sync"
	"fmt"
)

type Filter interface {
	// Handle method process the msg and return flag (true:continue, false: break)
	Handle(msg *Message)(bool)
}

type FilterCreator interface {
	Create(filterConf map[string]interface{}) Filter
}

var (
	registry = map[string] FilterCreator{}
	registryLock = &sync.RWMutex{}
	deniedNames = map[string]interface{}{"time": nil}
)

// RegisterCreator registers filter creator with name
func RegisterCreator(name string, creator FilterCreator) (err error) {
	registryLock.Lock()
	defer registryLock.Unlock()
	if _, ok := deniedNames[name]; ok {
		return fmt.Errorf("filter name %s denied", name)
	}

	if _, ok := registry[name]; ok {
		return fmt.Errorf("filter %s exists", name)
	}
	registry[name] = creator
	return err
}

// GetCreator returns filter creator with name
func GetCreator(name string) FilterCreator  {
	registryLock.RLock()
	defer registryLock.RUnlock()
	if creator, ok := registry[name]; ok {
		return creator
	}
	return nil
}

// filterRunner process messages with all filters
// appointed with filter configuration
type filterWorker struct {
	pAri *Ari
	filters []Filter
	// todo: inner filters, such as `time`
}

func newFilterWorker(pAri *Ari) *filterWorker {
	p := &filterWorker{pAri:pAri}
	// config outer filters
	//for filterName, pluginConf := range pAri.conf.FilterConf {
	//	if creator := GetCreator(filterName); creator != nil {
	//		filter := creator.Create(pluginConf)
	//		if filter != nil {
	//			p.filters = append(p.filters, filter)
	//		}
	//	}
	//}
	return p
}

func (f *filterWorker) loop(){
	var msg *Message
	for {
		select {
		case <- f.pAri.runningChan:
			goto exit
		case msg = <- f.pAri.MessageChan:
			f.handle(msg)
		}

	}
  exit:
}

func (f *filterWorker) handle(msg *Message) bool {
	for _, filter := range f.filters {
		if next := filter.Handle(msg); !next {
			// never to continue, send response to the sender
			close(msg.DoneChan)
			return false
		}
	}
	return true
}
