// Code generated from Pkl module `sfw`. DO NOT EDIT.
package config

type Web interface {
	Server
}

var _ Web = (*WebImpl)(nil)

type WebImpl struct {
	Host string `pkl:"host"`
}

func (rcv *WebImpl) GetHost() string {
	return rcv.Host
}
