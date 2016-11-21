package ari

import (
	"fmt"
	"sync"
)

// SysOpts is system options
// example:
// {"worker_num":30}
type SysOpts struct {
	FilterWorkerN int
}

func NewSysOpts(cf map[string]interface{}) (*SysOpts, error)  {
	// defaults
	s := &SysOpts{
		FilterWorkerN:10,
	}
	// cf is nil ,return default opts
	if cf == nil {
		return s, nil
	}
	// worker_num
	if num, ok := cf["worker_num"]; ok {
		s.FilterWorkerN = int(num.(float64))
	}
	return s, nil
}

type PluginOptions struct {
	// Plugin name
	PluginName string
	Conf       map[string]interface{}
}

type PluginGroup struct {
	// Group name
	Name string
	// Plugins options in the group
	Plugins []*PluginOptions
}

type Options struct {
	lock sync.RWMutex
	cfg map[string]interface{}
	*SysOpts

	inputGroups map[string]*PluginGroup
	filterGroups map[string]*PluginGroup
	outputGroups map[string]*PluginGroup
}

func NewOptions(cfg map[string]interface{}) (*Options, error) {
	var sysCfg map[string]interface{}
	if sys, ok := cfg["system"]; ok {
		sysCfg = sys.(map[string]interface{})
	}
	sysOpts, err := NewSysOpts(sysCfg)
	if err != nil {
		return nil, err
	}
	opts := &Options{
		cfg:cfg,
		SysOpts:sysOpts,
	}
	_, err = opts.InputGroups()
	if err != nil {
		return nil, err
	}
	_, err = opts.FilterGroups()
	if err != nil {
		return nil, err
	}
	_, err = opts.OutputGroups()
	if err != nil {
		return nil, err
	}
	return opts, nil
}

func (opts *Options) InputGroups()(map[string]*PluginGroup, error){
	if opts.inputGroups != nil {
		return opts.inputGroups, nil
	}
	opts.lock.Lock()
	defer opts.lock.Unlock()
	inputConf, ok := opts.cfg["input"]
	if !ok {
		return nil, nil
	}
	var inputGroups map[string]*PluginGroup
	for source, plugins := range inputConf.(map[string]interface{}) {
		if inputGroups == nil {
			inputGroups = make(map[string]*PluginGroup)
		}
		// plugins is a slice of map like [{"options": {...}, "plugin": "file"},...]
		pos := make([]*PluginOptions, len(plugins.([]interface{})))
		for i, plu := range plugins.([]interface{}) {
			plugin := plu.(map[string]interface{})
			if pluginName, nameOk := plugin["plugin"]; nameOk {
				pos[i] = &PluginOptions{
					PluginName:pluginName.(string),
					Conf:plugin["options"].(map[string]interface{}),
				}
			}else{
				return nil, fmt.Errorf("invalid input plugin (%s) conf", pluginName)
			}
		}
		inputGroups[source] = &PluginGroup{
			Name:source,
			Plugins:pos,
		}
	}
	opts.inputGroups = inputGroups
	return inputGroups, nil
}

func (opts *Options) FilterGroups() (map[string] *PluginGroup, error) {
	if opts.filterGroups != nil {
		return opts.filterGroups, nil
	}
	opts.lock.Lock()
	defer opts.lock.Unlock()
	inputConf, ok := opts.cfg["filter"]
	if !ok {
		return nil, nil
	}
	var filterGroups map[string]*PluginGroup
	for source, plugins := range inputConf.(map[string]interface{}) {
		if filterGroups == nil {
			filterGroups = make(map[string]*PluginGroup)
		}
		// plugins is a slice of map like [{"options": {...}, "plugin": "file"},...]
		pos := make([]*PluginOptions, len(plugins.([]interface{})))
		for i, plu := range plugins.([]interface{}) {
			plugin := plu.(map[string]interface{})
			if pluginName, nameOk := plugin["plugin"]; nameOk {
				pos[i] = &PluginOptions{
					PluginName:pluginName.(string),
					Conf:plugin["options"].(map[string]interface{}),
				}
			}else{
				return nil, fmt.Errorf("invalid filter plugin (%s) conf", pluginName)
			}
		}
		filterGroups[source] = &PluginGroup{
			Name:source,
			Plugins:pos,
		}
	}
	opts.filterGroups = filterGroups
	return filterGroups, nil
}

func (opts *Options) OutputGroups() (map[string] *PluginGroup, error)  {
	if opts.outputGroups != nil {
		return opts.outputGroups, nil
	}
	opts.lock.Lock()
	defer opts.lock.Unlock()
	inputConf, ok := opts.cfg["output"]
	if !ok {
		return nil, nil
	}
	var outputGroups map[string]*PluginGroup
	for source, plugins := range inputConf.(map[string]interface{}) {
		if outputGroups == nil {
			outputGroups = make(map[string]*PluginGroup)
		}
		// plugins is a slice of map like [{"options": {...}, "plugin": "file"},...]
		pos := make([]*PluginOptions, len(plugins.([]interface{})))
		for i, plu := range plugins.([]interface{}) {
			plugin := plu.(map[string]interface{})
			if pluginName, nameOk := plugin["plugin"]; nameOk {
				pos[i] = &PluginOptions{
					PluginName:pluginName.(string),
					Conf:plugin["options"].(map[string]interface{}),
				}
			}else{
				return nil, fmt.Errorf("invalid output plugin (%s) conf", pluginName)
			}
		}
		outputGroups[source] = &PluginGroup{
			Name:source,
			Plugins:pos,
		}
	}
	opts.outputGroups = outputGroups
	return outputGroups, nil
}
