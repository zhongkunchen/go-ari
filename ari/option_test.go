package ari

import (
	"testing"
	"encoding/json"
)

var demoJson = `
{
    "system": {
        "worker_num":20
    },

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
            }
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
        "g*": [
            {
                "plugin": "grok",
                "options": {
                    "message": ["pattern1"],
                    "a":99
                }
            },
            {
                "plugin": "date",
                "options": {
                    "pick": "request_time"
                }
            }
       ]
    },

    "output": {
        "g*": [{"plugin":"elastisearch", "options": {"a":99}}],
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
`

func TestOptions_InputGroups(t *testing.T) {
	var conf = map[string]interface{}{}
	err := json.Unmarshal([]byte(demoJson), &conf)
	if err != nil {
		t.Fatalf("err:%v", err)
	}
	opts, _:= NewOptions(conf)
	groups, err := opts.InputGroups()
	if err != nil {
		t.Fatalf("input groups err:%v", err)
	}
	p1 := groups["gw"].Plugins[0]
	if p1.Conf["tags"].([]interface{})[0].(string) != "server#1" {
		t.Fatal("fail to parse input plugin#1")
	}
	if p1.PluginName != "file" {
		t.Fatal("fail to parse input plugin#1#pluginname")
	}
}

func TestOptions_FilterOptions(t *testing.T) {
	var conf = map[string]interface{}{}
	err := json.Unmarshal([]byte(demoJson), &conf)
	if err != nil {
		t.Fatalf("err:%v", err)
	}
	opts, _:= NewOptions(conf)
	filters, err := opts.FilterGroups()
	if err != nil {
		t.Fatalf("filter groups err:%v", err)
	}
	p1 := filters["g*"].Plugins[0]
	if p1.PluginName != "grok" {
		t.Fatal("fail to parse filter plugin name")
	}
	if p1.Conf["a"].(float64) != 99{
		t.Fatal("fail to parse filter conf")
	}
}

func TestOptions_OutputGroups(t *testing.T) {
	var conf = map[string]interface{}{}
	err := json.Unmarshal([]byte(demoJson), &conf)
	if err != nil {
		t.Fatalf("err:%v", err)
	}
	opts, _:= NewOptions(conf)
	groups, err := opts.OutputGroups()
	if err != nil {
		t.Fatalf("input groups err:%v", err)
	}
	p1 := groups["g*"].Plugins[0]
	if p1.PluginName != "elastisearch" {
		t.Fatal("fail to parse output plugin name")
	}
	if p1.Conf["a"].(float64) != 99{
		t.Fatal("fail to parse output conf")
	}
}
