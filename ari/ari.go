package ari

import (
	"sync/atomic"
	"sync"
)


type Ari struct {
	sync.RWMutex
	waitGroup   WaitWrapper

	option      atomic.Value

	runningChan chan int
	isRunning   int32

	// messageChan receives log messages from log producer
	// filter workers receives log messages from messageChan
	messageChan chan *Message
}

// New creates instance of `*Ari`
func New(opt *Options) *Ari {
	p := &Ari{}
	return p
}

// Main is the entry to bootstrap Ari
func (p *Ari) Main() {
	p.Lock()
	defer p.Unlock()

	// start the input endpoint
	p.waitGroup.Add(func(){
		p.startProducers()
	})
	// start the filter endpoint
	p.waitGroup.Add(func(){
		p.startFilters()
	})
	// start the output endpoint
	p.waitGroup.Add(func(){
		p.outputPump()
	})

	// now i'm running
	atomic.StoreInt32(&p.isRunning, 1)
}

// NotifyStop stop all tasks
func (p *Ari) NotifyStop()  {
	close(p.runningChan)
	// wait all tasks finished
	p.waitGroup.Wait()
}

// startProducers bootstraps all registered message producers
func (p *Ari) startProducers() {
	//for name, conf := range p.conf.InputConf {
	//
	//}
}

// startFilters starts some filter workers to process log messages
func (p *Ari) startFilters()  {
	//for i:=uint32(0); i<p.Option().filterWorkerNum; i++ {
	//	worker := newFilterWorker(p)
	//	p.waitGroup.Add(func(){
	//		worker.loop()
	//	})
	//}
}

func (p *Ari) outputPump()  {
}

func (p *Ari) swapOption(opt *Options) {
	p.option.Store(opt)
}

// Option return `*Option`
func (p *Ari) Option() *Options {
	return p.option.Load().(*Options)
}
