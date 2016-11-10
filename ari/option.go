package ari
/**
{
    "input": {
        "gw": [
            {
                "plugin": "file",
                "options": {
                    "paths": ["/var/log/gateway/*.log"],
                    "tags":["server#1"]
                }
            },
            {
                "plugin": "file",
                "options": {
                    "paths": ["/var/log/gateway2/*.log"],
                    "tags":["server#2"]
                }
            },
        ],
        "test":[
            {
                "plugin": "generate",
                "options": {
                    "content": "[hello world]xxxxxxxx",
                    "times": 1000
                }
            }
        ]
    },
    "filter": {
        "g*": {
            "grok":{
                "message":["pattern1", "pattern2"]
            },
            "date": {
                "match": ["request_time", "yyMMdd HH:mm:ss"]
            }
        }
    },

    "output": {
        "g*": [{"plugin":"elastisearch", "options": {}}],
        "test": [
            {
                "plugin":"elastisearch",
                "options": {
                     "hosts":["localhost:9010"],
                     "document_type": "test",
                     "terms":{
                         "source": "test-server"
                     }
                }
            }
        ]
    }
}
 */

type FilterOptions struct {
	SourcePat string
	Conf map[string]interface{}
}

type PluginOptions struct {
	Plugin string
	Conf   map[string]interface{}
}

type PluginGroup struct {
	Name string
	Plugins []*PluginOptions
}

type Options struct {
	cfg map[string]interface{}
	filterWorkerNum int
}

func NewOptions(cfg map[string]interface{}) (*Options, error) {
	opts := &Options{
		filterWorkerNum:20,
		cfg:cfg,
	}
	return opts
}

func (opts *Options) InputGroups()(map[string]*PluginGroup){
}

func (opts *Options) FilterOptions()(map[string]*FilterOptions) {
}

func (opts *Options) OutputGroups() (map[string] *PluginGroup)  {
}
