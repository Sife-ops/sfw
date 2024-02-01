package ws

import (
	"sfw/db"

	"nhooyr.io/websocket"
)

type NState struct {
	Foo     string     `json:"foo"`
	Idle    bool       `json:"idle"`
	GodSeed db.GodSeed `json:"god_seed"`
}

type ConnNState struct {
	Conn   *websocket.Conn
	NState NState
}
