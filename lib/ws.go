package lib

import (
	"nhooyr.io/websocket"
)

type NState struct {
	Foo     string  `json:"foo"`
	Idle    bool    `json:"idle"`
	GodSeed GodSeed `json:"god_seed"`
}

type ConnNState struct {
	Conn   *websocket.Conn
	NState NState
}
