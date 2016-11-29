package main

import (
	"github.com/argpass/go-ari/ari"
	"syscall"
	_ "github.com/argpass/go-ari/plugins"
	_ "github.com/argpass/go-ari/plugins/file-beater"
	_ "github.com/argpass/go-ari/plugins/core"
	"github.com/argpass/go-ari/ari/log"
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
                    "paths": ["/Users/zkchen/tmp/log/ch_gateway/error.log"],
                    "tags":["server#1"],
                    "start_position": "beginning",
                    "codec":{
                        "multiline":{"token":"\\["}
                    }
                }
            },
            {
                "plugin": "file",
                "options": {
                    "paths": ["/var/log/gateway/*.log"],
                    "tags":["server#2"],
                    "start_position": "beginning",
                    "codec":{
                        "multiline":{"token":"\\["}
                    }
                }
            }
        ]
    },
    "filter": {
        "g*": [
            {
		"plugin": "grok",
	        "options": {}
	    }
        ]
    },

    "output": {
        "g*": [{"plugin":"stdout", "options": {"a":99}}]
    }
}
`

type program struct {
	Logger *log.Logger
	ari *ari.Ari
}

func (p *program) Init() (error) {
	return nil
}

func (p *program) Start() (error) {
	p.Logger.Warnf("program start...")
	p.ari.Main()
	return nil
}

func (p *program) Stop() (error) {
	p.ari.NotifyStop()
	p.Logger.Warnf("program stopped")
	return nil
}


func main()  {
	var cfg = map[string]interface{}{}
	err := json.Unmarshal([]byte(demoJson), &cfg)
	if err != nil {
		panic(err)
	}
	opts, e := ari.NewOptions(cfg)
	if e != nil {
		panic(e)
	}
	p := &program{
		Logger:log.GetLogger(),
		ari:ari.New(opts),
	}
	ari.Run(p, syscall.SIGINT, syscall.SIGTERM)
}
