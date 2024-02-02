package lib

import (
	"net"
)

type SockState struct {
	F0 string
	F1 GodSeed
}

type SockClient struct {
	Conn net.Conn
}

func (O *SockClient) Connect(server string) error {
	conn, err := net.Dial("tcp", server)
	if err != nil {
		return err
	}
	O.Conn = conn
	return nil
}
