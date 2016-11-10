package filebeater

import (
	"testing"
)

func showLogs(logs [][]byte, t *testing.T)  {
	if logs == nil {
		t.Log("logs nil")
		return
	}
	//t.Logf("logs len:%d", len(logs))
	for _, log := range logs {
		t.Logf("%s", string(log))
	}
}

func TestMultiLineCodec_NextLogs(t *testing.T) {
	codec, err:= NewMultiLineCodec(`\[`)
	if err != nil {
		t.Fatalf("unexpect err:%v", err)
	}
	// 8 tokens `[`
	source := []string {
		"[[I 2016-10-11 09:23:01] hello world, the first log item[[\n",
		"[I 2016-10-11 09:23:01] hello world, [the second log item\n",
		"[I 2016-10-11 09:23:01] hello world, [the third log item\n",
	}
	var logs = [][]byte{}
	for _, s:= range source {
		for _, log := range codec.NextLogs([]byte(s)) {
			logs = append(logs, log)
		}
	}
	logs = append(logs, codec.Done())
	if len(logs) != 8 {
		t.Fatalf("expect 8 logs, got %d", len(logs))
	}
	// show the logs
	showLogs(logs, t)
}
