package log

import (
	"fmt"
	"github.com/Sirupsen/logrus"
)

var defaultFactory = NewLoggerFactory()

type defaults struct {
	level     logrus.Level
	formatter logrus.Formatter
}

type LoggerFactory struct {
	defaults
	cache map[string]*Logger
}

// NewLoggerFactory constructor to build a `LoggerFactory`
func NewLoggerFactory() (factory *LoggerFactory) {
	p := &LoggerFactory{}
	p.level = logrus.DebugLevel
	p.cache = make(map[string]*Logger)

	formatter := &logrus.TextFormatter{}
	formatter.ForceColors = true
	formatter.FullTimestamp = true

	p.formatter = formatter

	return p
}

// GetLogger creates a `Logger` and returns,
// simplify the process to create a logger by
// setup logger with default text formatter and log level
// loggers will be cached with id
func (p *LoggerFactory) GetLogger(id string) (logger *Logger) {
	if logger, ok := p.cache[id]; ok {
		return logger
	}

	// create logger instance of `id`
	logger = &Logger{logrus.New(), id}
	logger.Level = p.defaults.level
	logger.Formatter = p.defaults.formatter
	p.cache[id] = logger

	return p.GetLogger(id)
}

func GetLogger(v ...interface{}) (logger *Logger) {
	id := "root"
	if len(v) > 0 {
		id = fmt.Sprintf("%v", v[0])
	}
	return defaultFactory.GetLogger(id)
}
