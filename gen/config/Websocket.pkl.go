// Code generated from Pkl module `sfw`. DO NOT EDIT.
package config

type Websocket interface {
	Server
}

var _ Websocket = (*WebsocketImpl)(nil)

type WebsocketImpl struct {
	Host string `pkl:"host"`
}

func (rcv *WebsocketImpl) GetHost() string {
	return rcv.Host
}
