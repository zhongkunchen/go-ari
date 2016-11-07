package ari

import (
	"sync"
	"bytes"
)

type WaitGroupWrapper struct {
	sync.WaitGroup
}

func (p *WaitGroupWrapper) Add (fn func()) {
	p.WaitGroup.Add(1)
	go func(){
		fn()
		p.WaitGroup.Done()
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
	bp.New = func()(*bytes.Buffer) {
		return &bytes.Buffer{}
	}
}

