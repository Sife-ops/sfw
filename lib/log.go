package lib

import (
	"fmt"
	"net"
)

var warn = 10

func init() {
	FlagParse()
}

type Logger struct{}

func (O Logger) Write(p []byte) (n int, err error) {
	fmt.Printf(string(p))

	conn, err := net.Dial("tcp", *FlagLogSrv)
	if err != nil {
		if warn%10 == 0 {
			fmt.Printf("%v\n", err)
		}
		warn++
		return len(p), nil
	}

	return conn.Write(p)
}
