package ari

import "github.com/argpass/go-ari/ari/log"

type Worker struct {
	ari *Ari
	id int
	context *Context
	Logger *log.Logger
}

func NewWorker(ari *Ari, id int) *Worker  {
	w := &Worker{
		ari:ari,
		id:id,
		context:ari.context,
		Logger:ari.context.Logger,
	}
	return w
}

func (w *Worker) exit()  {
	w.Logger.Debugf("[worker#%d] bye", w.id)
}

// Handle method filters the msg
func (w *Worker) Handle(msg *Message) (*Message, error) {
	// todo: filter the msg
	return msg, nil
}

func(w *Worker) DoWork() {
	w.Logger.Debugf("[worker#%d] working...", w.id)
	var msg *Message
	var inputChan = w.ari.MessageChan
	for {
		select {
		case msg = <-inputChan:
			filteredMsg, err := w.Handle(msg)
		        if err == nil {
				w.ari.Dispatch(filteredMsg)
			}else {
				w.Logger.Errorf("[workder#%d]%v",
					w.id, err)
			}
		case <-w.ari.closeChan:
			goto exit
		}
	}
exit:
	w.exit()
}
