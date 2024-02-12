// Code generated from Pkl module `sfw`. DO NOT EDIT.
package config

type Log interface {
	Server
}

var _ Log = (*LogImpl)(nil)

type LogImpl struct {
	Host string `pkl:"host"`
}

func (rcv *LogImpl) GetHost() string {
	return rcv.Host
}
