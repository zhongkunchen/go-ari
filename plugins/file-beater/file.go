package filebeater

import (
	"github.com/argpass/go-ari/ari"
	"sync"
	"os"
	"runtime"
	"io"
	"bufio"
	"errors"
	"path/filepath"
)

type FileBeaterOptions struct {
	conf         ari.Configuration

	// Paths are target files paths(blob format)
	Paths        []string
	MaxBlockSize uint

	// StartPos is the file offset to start
	// -1 means start at the end
	// now it only can be set as 0 or -1
	StartPos     int
}

func NewFileBeaterOptions(conf ari.Configuration) (*FileBeaterOptions, error) {
	opts := &FileBeaterOptions{
		conf:conf,
		MaxBlockSize:1024 * 32,
		StartPos:-1,
	}
	start, err := conf.GetString("start_position")
	if err != nil {
		return nil, err
	}
	if start == "beginning" {
		opts.StartPos = 0
	}
	if start == "end" {
		opts.StartPos = -1
	}
	return opts
}

var _ ari.Beater = &SingleFileBeater{}

// SingleFileBeater produces log messages from a log file,
// and sends messages to `messageChan`.
// a `FileBeater` only watches one log file.
type SingleFileBeater struct {
	sync.RWMutex

	messageChan chan <- *ari.Message
	options *FileBeaterOptions

	readFile *os.File
	reader *io.Reader
	readFileName string
	readPos uint64
	sentCnt uint64
	closeChan chan int
	blockBuf []byte
}

func NewSingleFileBeater(options *FileBeaterOptions) {
	f := &SingleFileBeater{
		options:options,
	}
	f.blockBuf = make([]byte, f.options.MaxBlockSize)
}

// Stop the beater
// notify the beater to exit
func (f *SingleFileBeater) Stop()  {
	close(f.closeChan)
}

func (f *SingleFileBeater) exit()  {
}

func (f *SingleFileBeater) readOneBlock() ([]byte, error) {
	var readN int
	var err error
	var data []byte
	// open file if necessary
	if f.readFile == nil {
		f.readFile, err = os.OpenFile(f.readFileName, os.O_RDONLY, 0666)
		if err != nil {
			return data, err
		}
		f.readPos = 0
		f.readFile.Seek(f.options, 0)
		f.reader = bufio.NewReader(f.readFile)
	}
	// read data to buf
	readN, err = io.ReadFull(f.reader, f.blockBuf)
	// touch eof, pause the beater to wait file changed
	if err == io.ErrUnexpectedEOF || readN <= 0{
	}
	data = make([]byte, readN)
	if copy(data, f.blockBuf) != readN {
		return data, errors.New("copy fail")
	}
	return data, nil
}

// WaitOneMessage reads one log message
func (f *SingleFileBeater) WaitOneMessage() (msg *ari.Message, err error) {
	var block []byte
	var pending [][]byte
	for {
		block, err = f.readOneBlock()
		if err != nil {
			return msg, err
		}

		pending = append(pending, block)
	}
}

func (f *SingleFileBeater) Beating()  {
	var err error
	var message *ari.Message
	var sendChan chan <- *ari.Message
	var doneChan <- chan *ari.Message
	for{
		if message == nil {
			message, err = f.WaitOneMessage()
			if err != nil {
				// continue to read one log block
				runtime.Gosched()
				continue
			}
			sendChan = f.messageChan
			doneChan = message.DoneChan
		}

		// now one log message is prepared
		select {
		case sendChan <- message:
			f.sentCnt++
			// wait response, close sendChan
			sendChan = nil
		case <- doneChan:
			// message send successfully
			doneChan = nil
			message = nil
		case <- f.closeChan:
			goto exit
		default:
			runtime.Gosched()
		}
	}
	exit:
	f.exit()
}

// FileBeater organize `SingleFileBeater` to produce log messages
type FileBeater struct {
	options *FileBeaterOptions
	sendChan chan <- *ari.Message
	closeChan chan int

	group ari.WaitGroupWrapper
}

func NewFileBeater(conf ari.Configuration,
	sendChan chan <- *ari.Message) (*FileBeater, error) {
	var opts *FileBeaterOptions
	var err error
	opts, err= NewFileBeaterOptions(conf)
	if err != nil {
		return nil, err
	}
	fb := &FileBeater{
		sendChan:sendChan,
		closeChan:make(chan int, 1),
		options:opts,
	}
	return fb, nil
}

// Beating method bootstraps all beaters
func (f *FileBeater) Beating()  {
	// paths set
	paths := map[string]interface{}{}
	// resolve paths with glob patterns
	for _, pathPattern := range f.options.Paths{
		matches, err := filepath.Glob(pathPattern)
		if err != nil {
			for p:=range matches {
				paths[p] = nil
			}
		}
		// invalid path patterns will be ignored
	}
	// todo: start SingleFileBeaters on every file path; add all to group
}

// Stop method notifies all `SingleFileBeater` to stop
func (f *FileBeater) Stop()  {
	close(f.closeChan)
	// waiting beaters to finish
	f.group.Wait()
}

type fileBeaterCreator struct {}

func (c *fileBeaterCreator) Create(conf ari.Configuration,
	messageChan chan <- *ari.Message) (beater ari.Beater, err error)  {
	beater, err = NewFileBeater(conf, messageChan)
	return beater, err
}

func init()  {
	ari.RegisterBeaterCreator("file", &fileBeaterCreator{})
}
