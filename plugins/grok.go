package plugins

import "github.com/argpass/go-ari/ari"

type grok struct {}

func (g *grok) config(conf map[string]interface{})  {

}

func new() *grok {
	return &grok{}
}

func (g *grok) Create(conf map[string]interface{}) ari.Filter {
	g.config(conf)
	return g
}

func (g *grok) Handle(msg *ari.Message) bool {
	msg.SetTerm("attach", "grok")
	return true
}

func init()  {
	//ari.RegisterCreator("grok", new())
}
