package plugins

import (
	"github.com/argpass/go-ari/ari"
	"sync"
)

var _ ari.Beater = &FileBeater{}

// FileBeater produces log messages from log a file,
// and sends messages to `messageChan`
type FileBeater struct {
	sync.RWMutex

	messageChan chan <- *ari.Message
	conf ari.Configuration
}

func (f *FileBeater) Beating()  {
}

type fileBeaterCreator struct {}

func (c *fileBeaterCreator) Create(conf ari.Configuration,
	messageChan chan <- *ari.Message) (beater ari.Beater, err error)  {
	fb := &FileBeater{conf:conf, messageChan:messageChan}
	// todo: setup fb with conf
	beater = fb
	return beater, err
}

func init()  {
	ari.RegisterBeaterCreator("file", &fileBeaterCreator{})
}
