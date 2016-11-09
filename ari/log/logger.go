package log

import "github.com/Sirupsen/logrus"

type Logger struct {
	*logrus.Logger
	ID string
}
