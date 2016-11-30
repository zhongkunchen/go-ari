package core

import (
	"github.com/argpass/go-ari/ari"
	"github.com/argpass/go-ari/ari/log"
)

type std struct {
	logger *log.Logger
	ctx *ari.Context
}

func newStd(logger *log.Logger, ctx *ari.Context) *std  {
	return &std{
		logger:logger,
		ctx:ctx,
	}
}

func (s *std) Send(msg *ari.Message)  {
	s.logger.Debugf("[stdout]%v", msg)
}

func (s *std) Run() error {
	// todo:
	return nil
}

type stdoutBuilder struct {
}

func (b *stdoutBuilder) Build(ctx *ari.Context,
	cfg map[string]interface{}) (ari.Sender, error) {
	return newStd(ctx.Logger, ctx), nil
}

type stdinBuilder struct {
}

func (b *stdinBuilder) Build(ctx *ari.Context,
cfg map[string]interface{}, groupName string) (ari.Beater) {
	return newStd(ctx.Logger, ctx)
}

func init() {
	var stdout = &stdoutBuilder{}
	var stdin = &stdinBuilder{}
	ari.SenderBuilders.Register("stdout", stdout)
	ari.BeaterBuilders.Register("stdin", stdin)
}
