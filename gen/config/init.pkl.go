// Code generated from Pkl module `sfw`. DO NOT EDIT.
package config

import "github.com/apple/pkl-go/pkl"

func init() {
	pkl.RegisterMapping("sfw", Sfw{})
	pkl.RegisterMapping("sfw#Postgres", PostgresImpl{})
	pkl.RegisterMapping("sfw#Log", LogImpl{})
	pkl.RegisterMapping("sfw#Web", WebImpl{})
	pkl.RegisterMapping("sfw#Websocket", WebsocketImpl{})
	pkl.RegisterMapping("sfw#Worldgen", Worldgen{})
}
