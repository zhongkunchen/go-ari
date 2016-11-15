package ari

import (
	"sync/atomic"
	"sync"
	"github.com/argpass/go-ari/ari/log"
	"fmt"
	"runtime"
)

type Context struct {
	Ari *Ari
	Opts *Options
	Logger log.Logger
}

var STATUS = struct{
	UNKNOWN int32
	STARTING int32
	RUNNING int32
	CLOSING int32
	DEAD int32
}{
	0,1,2,3,4,
}

type Ari struct {
	sync.RWMutex
	waitGroup   WaitWrapper
	context     *Context

	option      atomic.Value

	runningChan chan int
	status      int32

	// messageChan receives log messages from log producer
	// filter workers receives log messages from messageChan
	MessageChan chan *Message
}

// New creates instance of `*Ari`
func New(opts *Options) *Ari {
	p := &Ari{}
	p.context = &Context{Ari:p,Opts:opts, Logger:log.GetLogger()}
	atomic.StoreInt32(&p.status, STATUS.UNKNOWN)
	return p
}

// Main is the entry to bootstrap Ari
func (p *Ari) Main() {
	p.Lock()
	defer p.Unlock()
	atomic.StoreInt32(&p.status, STATUS.STARTING)

	// start the input endpoint
	err := p.startInputGroups()
	if err != nil {
		p.NotifyStop()
	}
	// start the filter endpoint
	err = p.startFilters()
	if err != nil {
		p.NotifyStop()
	}
	// start the output endpoint
	err = p.startOutputGroups()
	if err != nil {
		p.NotifyStop()
	}
	// now i'm running
	atomic.StoreInt32(&p.status, STATUS.RUNNING)
}

// NotifyStop stop all tasks
func (p *Ari) NotifyStop()  {
	close(p.runningChan)
	// wait all tasks finished
	p.waitGroup.Wait()
}

func (p *Ari) Fatalf(errMsg string, args ...interface{})  {
	p.context.Logger.Errorf(errMsg, args...)
	runtime.Goexit()
}

// startInputGroups bootstraps all registered message producers
func (p *Ari) startInputGroups() error {
	groupsMap , err := p.Options().InputGroups()
	if err != nil {
		return err
	}
	for _, group := range groupsMap {
		// start input group
		p.context.Logger.Debugf("start input group %s", group.Name)
		for _, pluginOpts := range group.Plugins {
			// get runner builder with plugin name
			rb := InputPlugins.get(pluginOpts.PluginName)
			if rb == nil {
				return fmt.Errorf("no such input plugin %s registered",
					pluginOpts.PluginName)
			}
			runner := rb.Build(p.context, pluginOpts.Conf)
			// start plugin in a goroutine
			p.waitGroup.Add(func(){
				err := runner.Run()
				if err != nil {
					p.Fatalf("[%s]%v", pluginOpts.PluginName, err)
				}
			})
		}
	}
	return nil
}

// startFilters starts some filter workers to process log messages
func (p *Ari) startFilters() error  {
	//for i:=uint32(0); i<p.Option().filterWorkerNum; i++ {
	//	worker := newFilterWorker(p)
	//	p.waitGroup.Add(func(){
	//		worker.loop()
	//	})
	//}
	return nil
}

func (p *Ari) startOutputGroups() error  {
	return nil
}

func (p *Ari) swapOption(opt *Options) {
	p.option.Store(opt)
}

// Options return `*Option`
func (p *Ari) Options() *Options {
	return p.option.Load().(*Options)
}
