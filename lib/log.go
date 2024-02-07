package lib

import (
	"fmt"
	"net"
	"time"
)

var warn = 10

func init() {
	FlagParse()
}

type Logger struct{}

func (O Logger) Write(p []byte) (n int, err error) {
	fmt.Printf(string(p))

	go func() {
		dialer := net.Dialer{Timeout: 3 * time.Second}
		conn, err := dialer.Dial("tcp", *FlagLogSrv)
		if err != nil {
			if warn%10 == 0 {
				fmt.Printf("%v\n", err)
			}
			warn++
			return
		}

		if _, err := conn.Write(p); err != nil {
			fmt.Printf("%v\n", err)
		}

		if err := conn.Close(); err != nil {
			fmt.Printf("%v\n", err)
		}
	}()

	return len(p), nil
}
