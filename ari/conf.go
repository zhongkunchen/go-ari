package ari

/**
{
    "input": {
        "file": {
            "path": ["abc.log.*"],
            "type": "gw",
            "tags": [206, "test"]
        },
        "kafka":{
        },
        "generator": {
            "lines": ["sefewdfsd", "fsdfagwef"],
            "times": 1024
        }
    },

    "filter": {
        "grok":{
            "message":["pattern1", "pattern2"]
        },
        "date": {
            "match": ["request_time", "yyMMdd HH:mm:ss"]
        }
    },

    "output": {
        "stdout": {
            "codec": "rubydebug"
        },
        "kafka": {

        },
        "logstash": {

        }
        "elastisearch": {
             "host": ["127.0.0.1", "host2"],
             "document_type": "gateway",
             "override_template":true,
             "routing": "iccid",
             "template":{...}
        }
    }
}
 */

type Config struct {
	InputConf map[string]map[string]interface{} 	`json:"input"`
	FilterConf map[string]map[string]interface{} 	`json:"filter"`
	outputConf map[string]map[string]interface{} 	`json:"output"`
}


