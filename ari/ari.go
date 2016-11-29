package ari

import (
	"sync/atomic"
	"sync"
	"github.com/argpass/go-ari/ari/log"
	"fmt"
	"os"
	"regexp"
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

	// sendersMap: input group name => []Sender
	sendersMap  map[string] []Sender
	// beatersMap: input group name => []Beater
	beatersMap  map[string] []Beater
	// filtersMap: input group name => []Filter
	filtersMap map[string] []Filter
}

// New creates instance of `*Ari`
func New(opts *Options) *Ari {
	p := &Ari{
		closeChan:make(chan int, 1),
		MessageChan:make(chan *Message, 1),
		sendersMap:map[string][]Sender{},
		filtersMap:map[string][]Filter{},
		beatersMap:map[string][]Beater{},
	}
	p.context = &Context{Ari:p,Opts:opts, Logger:log.GetLogger()}
	p.opts.Store(opts)
	atomic.StoreInt32(&p.status, STATUS.UNKNOWN)

	// setup beaters filters and senders with options
	p.mustSetupBeaters()
	p.mustSetupFilters()
	p.mustSetupSenders()
	return p
}

func (p *Ari) mustSetupFilters()  {
	// build filtersMap
	cfg, err := p.Options().FilterGroups()
	if err != nil {
		panic(err)
	}
	for pattern, plgs := range cfg {
		if len(plgs.Plugins) == 0 {
			continue
		}
		var filters []Filter
		for i, plg := range plgs.Plugins {
			if filters == nil {
				filters = make([]Filter, len(plgs.Plugins))
			}
			builder := FilterBuilders.get(plg.PluginName)
			if builder == nil {
				panic(fmt.Errorf("expect filter %s registered", plg.PluginName))
			}
			filter, err := builder.Build(p.context, plg.Conf)
			if err != nil {
				panic(err)
			}
			filters[i] = filter
		}
		if len(filters) != len(plgs.Plugins) {
			panic(fmt.Errorf("build senders fail, cfg:%v", plgs.Plugins))
		}
		reg, err := regexp.Compile(pattern)
		if err != nil {
			panic(fmt.Errorf("invalid pattern %s", pattern))
		}
		// attach to the `beatersMap`
		for gpName, _ := range p.beatersMap {
			if reg.Match([]byte(gpName)) {
				if attached, exists := p.filtersMap[gpName]; !exists {
					merge := make([]Filter, len(attached) + len(filters))
					i := copy(merge, attached)
					copy(merge[i:], filters)
					p.filtersMap[gpName] = merge
				}
			}
		}
	}
}

func (p *Ari) mustSetupBeaters()  {
	// build beaters
	cfg, err := p.Options().InputGroups()
	if err != nil {
		panic(err)
	}
	for gpName, plgs := range cfg {
		if len(plgs.Plugins) == 0 {
			continue
		}
		var beaters []Beater
		for i, plg := range plgs.Plugins {
			if beaters == nil {
				beaters = make([]Beater, len(plgs.Plugins))
			}
			builder := BeaterBuilders.get(plg.PluginName)
			if builder == nil {
				panic(fmt.Errorf("expect input plugin %s registered",
					plg.PluginName))
			}
			beater := builder.Build(p.context, plg.Conf, gpName)
			beaters[i] = beater
		}
		p.beatersMap[gpName] = beaters
	}
}

func (p *Ari) mustSetupSenders() {
	// build sendersMap
	outputCfg, err := p.Options().OutputGroups()
	if err != nil {
		panic(err)
	}
	for pattern, plgs := range outputCfg {
		if len(plgs.Plugins) == 0 {
			continue
		}
		var senders []Sender
		for i, plg := range plgs.Plugins {
			if senders == nil {
				senders = make([]Sender, len(plgs.Plugins))
			}
			builder := SenderBuilders.get(plg.PluginName)
			if builder == nil {
				panic(fmt.Errorf("expect sender %s registered", plg.PluginName))
			}
			sender, err := builder.Build(p.context, plg.Conf)
			if err != nil {
				panic(err)
			}
			senders[i] = sender
		}
		if len(senders) != len(plgs.Plugins) {
			panic(fmt.Errorf("build senders fail, cfg:%v", plgs.Plugins))
		}
		reg, err := regexp.Compile(pattern)
		if err != nil {
			panic(fmt.Errorf("invalid pattern %s", pattern))
		}
		// attach to the `beatersMap`
		for gpName, _ := range p.beatersMap {
			if reg.Match([]byte(gpName)) {
				if attached, exists := p.sendersMap[gpName]; !exists {
					merge := make([]Sender, len(attached) + len(senders))
					i := copy(merge, attached)
					copy(merge[i:], senders)
					p.sendersMap[gpName] = merge
				}
			}
		}
	}
}

// Main is the entry to bootstrap Ari
func (p *Ari) Main() {
	p.Lock()
	defer p.Unlock()
	atomic.StoreInt32(&p.status, STATUS.STARTING)

	// start the input endpoint
	err := p.startBeaters()
	if err != nil {
		p.Fatalf("%v", err)
	}
	// start workers
	err = p.startFiltering()
	if err != nil {
		p.Fatalf("%v", err)
	}
	// now i'm running
	atomic.StoreInt32(&p.status, STATUS.RUNNING)
}

func (p *Ari) getSenders(groupName string) []Sender {
	senders, exists := p.sendersMap[groupName]
	if exists {
		return senders
	}
	return nil
}

// Dispatch a msg to all senders
func (p *Ari) Dispatch(msg *Message){
	close(msg.DoneChan)
	for _, sender := range p.getSenders(msg.GroupName) {
		sender.Send(msg)
	}
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

// startBeaters bootstraps all registered message producers
func (p *Ari) startBeaters() error {
	for _, beaters := range p.beatersMap {
		for _, beater := range beaters {

			p.waitGroup.WaitGroup.Add(1)
			go func(bt Beater){
				defer p.waitGroup.WaitGroup.Done()
				err := bt.Run()
				if err != nil {
					p.Fatalf("[%s]%v", err)
				}
			}(beater)
		}
	}
	return nil
}

func (p *Ari) doFilters() {
	var msg *Message
	var msgChan chan *Message = p.MessageChan
	for {
		select {
		case msg = <- msgChan:
			filters, _ := p.filtersMap[msg.GroupName]
			for _, filter := range filters {
				if !filter.DoFilter(msg) {
					break
				}
			}
			p.Dispatch(msg)
		case  <- p.closeChan:
			goto exit
		}
	}
	exit:
}

// startFiltering starts workers to process log messages
func (p *Ari) startFiltering() error  {
	p.context.Logger.Debugf("start filter worker num %d",
		p.Options().SysOpts.FilterWorkerN)
	for i:=1; i<=p.Options().SysOpts.FilterWorkerN; i++ {
		p.waitGroup.Add(func(){
			p.doFilters()
		})
	}
	return nil
}

func (p *Ari) swapOption(opt *Options) {
	p.opts.Store(opt)
}

// Options return `*Option`
func (p *Ari) Options() *Options {
	return p.opts.Load().(*Options)
}
