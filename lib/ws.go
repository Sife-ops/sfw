package lib

import (
	"context"
	"fmt"

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

var Ws *websocket.Conn

func Dial(s *string) error {
	conn, _, err_ := websocket.Dial(
		context.TODO(),
		fmt.Sprintf("ws://%s", *s),
		nil,
	)
	if err_ != nil {
		return err_
	}
	Ws = conn
	return nil
}
