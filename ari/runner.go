package ari

import (
	"os"
	"syscall"
	"os/signal"
)

var signalNotify = signal.Notify

type Service interface {
	Init() (error)
	Start() (error)
	Stop() (error)
}

func Run(service Service, sigs ...os.Signal)(err error){
	if len(sigs) == 0 {
		sigs = []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	}
	// bootstrap the service
	if err = service.Init(); err != nil {
		return err
	}
	if err = service.Start(); err != nil {
		return err
	}
	ch := make(chan <-os.Signal, 1)
	signalNotify(ch, sigs...)
	<-ch
	return service.Stop()
}
