package main

import (
	"github.com/fsnotify/fsnotify"
	"fmt"
)

func main()  {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	var event fsnotify.Event
	var e error
	go func(){
		for {
			select {
			case event = <- w.Events:
				fmt.Println("event:", event)
			case e = <- w.Errors:
				fmt.Println("err:", e)
			}
		}

	}()
	w.Add("/Users/zkchen/tmp/t.log")
	w.Add("/Users/zkchen/tmp/t.log")
	w.Add("/Users/zkchen/tmp/t.log")
	done := make(chan int, 1)
	<- done
}
