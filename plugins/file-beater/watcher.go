package filebeater

import (
	"github.com/fsnotify/fsnotify"
	"fmt"
	"sync"
)

type CanNotifier interface {
	Notify(event fsnotify.Event)
}

type fileWatcher struct {
	sync.WaitGroup
	closeChan chan struct{}
	*fsnotify.Watcher
	subscribersMap map[string]map[CanNotifier]struct{}
}

func newFileWatcher() *fileWatcher {
	var err error
	w := &fileWatcher{
		subscribersMap:map[string]map[CanNotifier]struct{}{},
		closeChan:make(chan struct{}, 1),
	}
	w.Watcher, err = fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	w.WaitGroup.Add(1)
	go func(){
		defer w.WaitGroup.Done()
		w.pump()
	}()
	return w
}

func (w *fileWatcher) pump() {
	defer w.Watcher.Close()
	var event fsnotify.Event
	var err error
	for {
		select {
		case event = <- w.Events:
			// notify subscribers
			w.dispatch(event)
		case err = <- w.Errors:
			w.handleError(err)
		case <- w.closeChan:
			goto exit
		}
	}
	exit:
}

func (w * fileWatcher) handleError(err error)  {
	fmt.Printf("[E][fileWatcher]%+v\n", err)
}

func (w *fileWatcher) dispatch(event fsnotify.Event) {
	fmt.Printf("watcher:dispatch:%+v, len event:%d\n", event, len(w.Events))
	subscribers, exists := w.subscribersMap[event.Name]
	if !exists {
		return
	}
	for subscriber, _ := range subscribers {
		subscriber.Notify(event)
	}
}

func (w *fileWatcher) Stop()  {
	close(w.closeChan)
	w.WaitGroup.Wait()
	fmt.Println("[W][fileWatcher]stopped")
}

func (w *fileWatcher) EnsureWatch(path string, subscriber CanNotifier) error  {
	d, exists := w.subscribersMap[path]
	if !exists {
		w.subscribersMap[path] = map[CanNotifier]struct{}{subscriber: struct{}{}}
		err := w.Watcher.Add(path)
		if err != nil {
			return err
		}
	}else{
		d[subscriber] = struct{}{}
	}
	return nil
}

func (w *fileWatcher) EnsureUnWatch(path string, subscriber CanNotifier) error  {
	delete(w.subscribersMap, path)
	if _, exists := w.subscribersMap[path]; !exists {
		err := w.Watcher.Remove(path)
		if err != nil {
			return err
		}
	}
	return nil
}

var (
	Watcher = newFileWatcher()
)

