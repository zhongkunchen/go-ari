package main

import (
	"github.com/argpass/go-runner/runner"
	"syscall"
	_ "github.com/argpass/ari/plugins"
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
	runner.Run(p, syscall.SIGINT, syscall.SIGTERM)
}
