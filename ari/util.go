package ari

import (
	"sync"
	"bytes"
)

type WaitWrapper struct {
	sync.WaitGroup
}

func (p *WaitWrapper) Add (fn func()) {
	p.WaitGroup.Add(1)
	go func(){
		defer p.WaitGroup.Done()
		fn()
	}()
}

var bp sync.Pool

// GetBuffer returns a `Buffer` of bytes
func GetBuffer() (*bytes.Buffer) {
	return bp.Get().(*bytes.Buffer)
}

// PutBuffer gives `Buffer` of bytes back to the pool
func PutBuffer(bf *bytes.Buffer){
	bp.Put(bf)
}

func init()  {
	// init the Buffer Pool
	bp.New = func()(interface{}) {
		return &bytes.Buffer{}
	}
}

