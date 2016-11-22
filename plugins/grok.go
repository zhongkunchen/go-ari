package plugins

import (
	"github.com/argpass/go-ari/ari"
)



var _ ari.Filter = &grokFilter{}

type grokFilter struct {
	ctx *ari.Context
}

func newGrokFilter(ctx *ari.Context,
	cfg map[string]interface{}) (*grokFilter, error) {
	g := &grokFilter{
		ctx:ctx,
	}
	return g, nil
}

// todo: filter the msg
func (g *grokFilter) DoFilter(msg *ari.Message)bool {
	msg.SetTerm("attach", "grok")
	msg.SetTerm("name", "fromg")
	return true
}

var _ ari.FilterBuilder = &grokBuilder{}

type grokBuilder struct {
}

func(gb *grokBuilder) Build(ctx *ari.Context,
	cfg map[string]interface{}) (ari.Filter, error){
	return newGrokFilter(ctx, cfg)
}

func init()  {
	ari.FilterBuilders.Register("grok", &grokBuilder{})
}
