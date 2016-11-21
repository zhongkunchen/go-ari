package ari

import (
	"github.com/argpass/go-ari/ari/log"
	"regexp"
	"sync"
)

type GroupFilter struct {
	filters []Filter
}

func newGroupFilter(ctx *Context, optsList []*PluginOptions) (*GroupFilter,error) {
	filters := make([]Filter, len(optsList))
	for i, opts := range optsList {
		fb := FilterBuilders.get(opts.PluginName)
		if fb != nil {
			fi, err := fb.Build(ctx, opts.Conf)
			if err != nil {
				return nil, err
			}
			filters[i] = fi
		}else{
			filters[i] = nil
		}
	}
	g := &GroupFilter{
		filters:filters,
	}
	return g, nil
}

func (g *GroupFilter) DoFilter(msg *Message) bool {
	for _, fi := range g.filters {
		if fi != nil {
			if !fi.DoFilter(msg) {
				break
			}
		}
	}
	return true
}

type Worker struct {
	lock sync.RWMutex
	ari *Ari
	id int
	context *Context
	Logger *log.Logger
	groupFilters map[string] Filter
}

func NewWorker(ari *Ari, id int) *Worker  {
	w := &Worker{
		ari:ari,
		id:id,
		context:ari.context,
		Logger:ari.context.Logger,
		groupFilters:map[string]Filter{},
	}
	return w
}

func (w *Worker) exit()  {
	w.Logger.Debugf("[worker#%d] bye", w.id)
}

// GetGroupFilter gets or creates a group filter for the name `groupName`
// got nil, if no filters configured for the source group
func (w *Worker) GetGroupFilter(groupName string) (Filter, error) {
	fi, exists := w.groupFilters[groupName]
	if exists {
		return fi, nil
	}
	groups, err := w.ari.Options().FilterGroups()
	if err != nil {
		return nil, err
	}
	for _, group := range groups {
		reg, err := regexp.Compile(group.Name)
		if err != nil {
			return nil, err
		}
		if reg.Match([]byte(groupName)) {
			fi, err := newGroupFilter(w.context, group.Plugins)
			if err != nil {
				return nil, err
			}
			w.groupFilters[groupName] = Filter(fi)
		}
	}
	if _, exists = w.groupFilters[groupName]; !exists {
		w.groupFilters[groupName] = nil
	}
	return w.GetGroupFilter(groupName)
}

// Handle method filters the msg
func (w *Worker) Handle(msg *Message) (*Message, error) {
	fi, err := w.GetGroupFilter(msg.GroupName)
	if err != nil {
		return msg, err
	}
	if fi != nil {
		fi.DoFilter(msg)
	}
	return msg, nil
}

func(w *Worker) DoWork() {
	w.Logger.Debugf("[worker#%d] working...", w.id)
	var msg *Message
	var inputChan = w.ari.MessageChan
	for {
		select {
		case msg = <-inputChan:
			filteredMsg, err := w.Handle(msg)
			if err == nil {
				w.ari.Dispatch(filteredMsg)
			}else {
				w.Logger.Errorf("[workder#%d]%v",
					w.id, err)
			}
		case <-w.ari.closeChan:
			goto exit
		}
	}
exit:
	w.exit()
}
