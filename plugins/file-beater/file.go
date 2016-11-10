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
	"strings"
	"github.com/argpass/go-ari/ari/log"
	"fmt"
)

// Options
// Example:
//
//{
//  "paths": ["/var/log/*.log", "/var/log/*.log1"],
//  "start_position": "beginning",
//  "tags": ["gateway", "server#1"],
//  "codec": {
//    "multiline": {
//      "token": "\\["
//    }
//  }
//}
type Options struct {
	Tags 		[]string
	// Paths are target files paths(blob format)
	Paths        	[]string
	MaxBlockSize 	uint

	// StartPos is the file offset to start
	// -1 means start at the end
	// now it only can be set as 0 or -1
	StartPos     int
	CodecOption map[string]interface{}
}

func NewOptions(conf ari.Configuration) (opts *Options, err error) {
	defer func(){
		if e := recover(); e != nil {
			opts, err = nil, fmt.Errorf("invalid config, err:%v", e)
		}
	}()
	opts = &Options{
		MaxBlockSize:1024 * 32,
		StartPos:-1,
	}
	var ok bool
	if opts.CodecOption, ok = conf["codec"]; !ok {
		return nil, errors.New("expect `codec` conf")
	}
	// resolve start position
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
	// paths
	if paths, exists := conf["paths"]; exists {
		if ps, ok := paths.([]interface{}); ok {
			if opts.Paths == nil {
				opts.Paths = make([]string, len(ps))
			}
			for _, p := range ps {
				opts.Paths = append(opts.Paths, p.(string))
			}
		}else{
			return nil, errors.New("expect slice paths type")
		}
	}else{
		return nil, errors.New("need paths config")
	}
	err = nil
	return opts,  err
}

// FileBeater produces log messages from a log file,
// and sends messages to `messageChan`.
// a `FileBeater` only watches one log file.
type FileBeater struct {
	runner       *BeaterRunner

	SendChan     chan <- *ari.Message
	Logger       *log.Logger
	Codec        Codec

	readFile     *os.File
	reader       *bufio.Reader
	readFileName string
	readPos      uint64
	sentCnt      uint64
	closeChan    chan int
	blockBuf     []byte
	// StartPos is the file offset to start
	// -1 means start at the end
	// now it only can be set as 0 or -1
	StartPos     int
	MaxBlockSize uint
	FilePath     string
	undelivered  []*ari.Message
}

func NewFileBeater(runner *BeaterRunner, path string, startPos int, maxBlockSize uint,
	sendChan chan <- *ari.Message, logger *log.Logger, codec Codec) *FileBeater {
	f := &FileBeater{
		runner:runner,
		Codec:codec,
		StartPos:startPos,
		MaxBlockSize:maxBlockSize,
		FilePath:path,
		SendChan:sendChan,
		Logger:logger,
	}
	f.blockBuf = make([]byte, f.MaxBlockSize)
	return f
}

func (f *FileBeater) String() string {
	return strings.Join([]string{"FILE-BEATER", f.FilePath}, ":")
}

// Stop the beater
// notify the beater to exit
func (f *FileBeater) Stop()  {
	close(f.closeChan)
}

func (f *FileBeater) exit()  {
	f.Logger.Debugf("%s bye", f.String())
}

func (f *FileBeater) readOne() ([]byte, error) {
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
		if f.StartPos == 0 {
			// seek to the begin
			f.readFile.Seek(0, 0)
		}else{
			// seek to the end
			f.readFile.Seek(0, 2)
		}
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

// WaitOneMSG todo: WaitOneMessage reads one log message
func (f *FileBeater) WaitOneMSG() (msg *ari.Message, err error) {
	var block []byte
	for {
		block, err = f.readOne()
		if err != nil {
			return msg, err
		}
		for _, msg f.Codec.NextLogs(block)
	}
}

func (f *FileBeater) Beating()  {
	f.Logger.Debugf("%s start beating", f.String())
	var err error
	var message *ari.Message
	var sendChan chan <- *ari.Message
	var doneChan <- chan int
	for{
		if message == nil {
			message, err = f.WaitOneMSG()
			if err != nil {
				// got err, stop the beater
				// todo: maybe i can wait, err may disappear
				// 可能错误是文件中途变化引起的（比如日志文件被重建）
				f.Logger.Debugf("%s, fail to wait one msg, err:%v",
					f.String(), err)
				f.Stop()
			}else{
				sendChan = f.SendChan
				doneChan = message.DoneChan
			}
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

var _ ari.Beater = &BeaterRunner{}

// BeaterRunner organize `FileBeater` to produce log messages
type BeaterRunner struct {
	sync.RWMutex
	Logger    *log.Logger

	beaters   []*FileBeater
	options   *Options
	sendChan  chan <- *ari.Message
	closeChan chan int

	group     ari.WaitWrapper
}

// NewBeaterRunner creates a new BeaterRunner instance
func NewBeaterRunner(conf ari.Configuration,
	sendChan chan <- *ari.Message, logger *log.Logger) (*BeaterRunner, error) {
	var opts *Options
	var err error
	opts, err= NewOptions(conf)
	if err != nil {
		return nil, err
	}
	fb := &BeaterRunner{
		sendChan:sendChan,
		Logger:logger,
		closeChan:make(chan int, 1),
		options:opts,
	}
	return fb, nil
}

func (r *BeaterRunner) EncodeMSG(body []byte) *ari.Message {
	m := &ari.Message{
		r.options
	}
	return m
}

func (r *BeaterRunner) Fatal(err error)  {
	r.Logger.Errorf("got err:%v", err)
	runtime.Goexit()
}

// Beating method bootstraps all beaters
func (r *BeaterRunner) Beating()  {
	var err error
	// paths set
	paths := map[string]interface{}{}
	// resolve paths with glob patterns
	for _, pathPattern := range r.options.Paths {
		matches, err := filepath.Glob(pathPattern)
		if err != nil {
			for _, p:=range matches {
				paths[p] = nil
			}
		}
		// invalid path patterns will be ignored
	}
	// start FileBeater on every file path; add all to group
	if r.beaters == nil {
		r.beaters = make([]*FileBeater, len(paths))
	}
	r.Logger.Debugf("[FBR]is to start %d beaters", len(paths))
	// pickup one codec config
	var codecName string
	var codecOpts map[string]interface{}
	for name, codecOpts := range r.options.CodecOption {
		codecName, codecOpts = name, codecOpts
		break
	}
	// create beaters
	for path, _ := range paths {
		// every beater has its own codec
		var codec Codec;
		codec, err = resolveCodec(codecName, codecOpts)
		if err != nil {
			r.Fatal(err)
		}
		break
		beater := NewFileBeater(r, path, r.options.StartPos,
			r.options.MaxBlockSize, r.sendChan, r.Logger, codec)
		r.beaters = append(r.beaters, beater)
	}

	// bootstrap all beaters
	for _, beater := range r.beaters {
		r.group.Add(func(){
			// bootstrap the beater
			beater.Beating()
		})
	}
}

func resolveCodec(name string, codecOpts map[string]interface{}) (Codec, error) {
	if name == "multiline" {
		pattern, ok := codecOpts["token"];
		if !ok {
			return nil, errors.New("multiline codec expects `token` conf")
		}
		return NewMultiLineCodec(pattern), nil
	}
	return nil, fmt.Errorf("unsupportted codec %s", name)
}

// Stop method notifies all `SingleFileBeater` to stop
// the method will be blocked to wait all beaters to be finished
func (r *BeaterRunner) Stop()  {
	close(r.closeChan)
	// waiting beaters to finish
	r.group.Wait()
}

type runnerCreator struct {}

func (c *runnerCreator) Create(conf ari.Configuration,
	messageChan chan <- *ari.Message) (beater ari.Beater, err error)  {
	beater, err = NewBeaterRunner(conf, messageChan, log.GetLogger())
	return beater, err
}

func init()  {
	// register the plugin
	ari.RegisterBeaterCreator("file", &runnerCreator{})
}
