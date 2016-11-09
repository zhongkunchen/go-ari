package log

import (
	"github.com/Sirupsen/logrus"
	"testing"
)

func TestLoggerFactory_GetLogger(t *testing.T) {
	logger := NewLoggerFactory().GetLogger("root")
	logger.WithFields(logrus.Fields{"[sn]": "sn-1998"}).Warn("go")
	logger.Warn("msg asewfwaefawef safwefawfawefawefawef    asdfasdfawefasfew asfaewfw-.....sfawefawfawefawefaewfaewfasefasdf.", "helloworld")
	logger.Warnf("%s bye", "akun")
}

func TestGetLogger(t *testing.T) {
	logger := GetLogger()
	logger.Warn("hello , id:", logger.ID)
	logger2 := GetLogger()
	if logger != logger2 {
		t.Logf("Expect got the same logger")
		t.Fail()
	}
	logger3 := GetLogger("root.me")
	if logger3 == logger2 {
		t.Logf("Expect got different loggers")
		t.Fail()
	}
}
