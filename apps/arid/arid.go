package main

import (
	"github.com/argpass/go-ari/ari"
	"syscall"
	_ "github.com/argpass/go-ari/plugins"
)

type program struct {
}

func (p *program) Init() (error) {
	return nil
}

func (p *program) Start() (error) {
	return nil
}

func (p *program) Stop() (error) {
	return nil
}


func main()  {
	p := &program{}
	ari.Run(p, syscall.SIGINT, syscall.SIGTERM)
}
