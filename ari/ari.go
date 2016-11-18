package ari

import (
	"sync/atomic"
	"sync"
	"github.com/argpass/go-ari/ari/log"
	"fmt"
	"os"
)

type Context struct {
	Ari *Ari
	Opts *Options
	Logger *log.Logger
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

	opts        atomic.Value

	closeChan   chan int
	status      int32

	// messageChan receives log messages from log producer
	// filter workers receives log messages from messageChan
	MessageChan chan *Message
}

// New creates instance of `*Ari`
func New(opts *Options) *Ari {
	p := &Ari{
		closeChan:make(chan int, 1),
		MessageChan:make(chan *Message, 1),
	}
	p.context = &Context{Ari:p,Opts:opts, Logger:log.GetLogger()}
	p.opts.Store(opts)
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
		p.Fatalf("%v", err)
	}
	// start workers
	err = p.startWorkers()
	if err != nil {
		p.Fatalf("%v", err)
	}
	// start the output endpoint
	err = p.startOutputGroups()
	if err != nil {
		p.Fatalf("%v", err)
	}
	// now i'm running
	atomic.StoreInt32(&p.status, STATUS.RUNNING)
}

// Dispatch a msg to all senders
func (p *Ari) Dispatch(msg *Message){
	// todo: dispatch the msg to all senders
	close(msg.DoneChan)
	p.context.Logger.Debugf("msg:%v", msg)
}

// NotifyStop stop all tasks
func (p *Ari) NotifyStop()  {
	close(p.closeChan)
	// wait all tasks finished
	p.waitGroup.Wait()
	os.Exit(0)
}

func (p *Ari) WrapMessage(body []byte, groupName string) *Message {
	msg := NewMessage(make(chan int, 1), 0, nil, body, nil, groupName)
	return msg
}

func (p *Ari) Fatalf(errMsg string, args ...interface{})  {
	p.context.Logger.Errorf(errMsg, args...)
	os.Exit(-1)
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
			rb := InputRunnerBuilders.get(pluginOpts.PluginName)
			if rb == nil {
				return fmt.Errorf("no such input plugin %s registered",
					pluginOpts.PluginName)
			}
			runner := rb.Build(p.context, pluginOpts.Conf, group.Name)
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

// startWorkers starts workers to process log messages
func (p *Ari) startWorkers() error  {
	p.context.Logger.Debugf("start workers num %d",
		p.Options().SysOpts.FilterWorkerN)
	for i:=1; i<=p.Options().SysOpts.FilterWorkerN; i++ {
		worker := NewWorker(p, i)
		p.waitGroup.Add(func(){
			worker.DoWork()
		})
	}
	return nil
}

func (p *Ari) startOutputGroups() error  {
	return nil
}

func (p *Ari) swapOption(opt *Options) {
	p.opts.Store(opt)
}

// Options return `*Option`
func (p *Ari) Options() *Options {
	return p.opts.Load().(*Options)
}
