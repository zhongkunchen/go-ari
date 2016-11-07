package ari

import (
	"sync/atomic"
	"sync"
)


type Ari struct {
	sync.RWMutex
	waitGroup WaitGroupWrapper

	option atomic.Value
	conf Config

	runningChan chan int
	isRunning int32

	messageChan chan *Message
}

// New creates instance of `*Ari`
func New(opt *Option) *Ari {
	p := &Ari{}
	return p
}

// Main is the entry to bootstrap Ari
func (p *Ari) Main() {
	p.Lock()
	defer p.Unlock()

	// start the input endpoint
	p.waitGroup.Add(func(){
		p.inputPump()
	})
	// start the filter endpoint
	p.waitGroup.Add(func(){
		p.filterPump()
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

func (p *Ari) inputPump()  {

}

// filterPump starts some filter workers to process log messages
func (p *Ari) filterPump()  {
	for i:=0; i<p.Option().filterWorkerNum; i++ {
		worker := newFilterWorker(p)
		p.waitGroup.Add(func(){
			worker.loop()
		})
	}
}

func (p *Ari) outputPump()  {
}

func (p *Ari) swapOption(opt *Option) {
	p.option.Store(opt)
}

// Option return `*Option`
func (p *Ari) Option() *Option {
	return p.option.Load().(*Option)
}
