package lib

import (
	"context"
	"fmt"
	"os"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type Logger struct{}

var warn = 10

func init() {
	FlagParse()
}

// todo split log files
func (O Logger) Write(p []byte) (n int, err error) {
	fmt.Printf(string(p))

	go func() {
		conn, _, err := websocket.Dial(context.TODO(), fmt.Sprintf("ws://%s", *FlagWsSrv), nil)
		if err != nil {
			if warn%20 == 0 {
				fmt.Printf("%v\n", err)
			}
			warn++
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		m := struct {
			Hostname string
			Message  string
		}{
			Hostname: func() string {
				name, err := os.Hostname()
				if err != nil {
					return "unknown"
				}
				return name
			}(),
			Message: string(p),
		}
		if err := wsjson.Write(ctx, conn, m); err != nil {
			fmt.Printf("%v\n", err)
		}

		if err := conn.CloseNow(); err != nil {
			fmt.Printf("%v\n", err)
		}
	}()

	// go func() {
	// 	dialer := net.Dialer{Timeout: 3 * time.Second}
	// 	conn, err := dialer.Dial("tcp", *FlagLogSrv)
	// 	if err != nil {
	// 		if warn%20 == 0 {
	// 			fmt.Printf("%v\n", err)
	// 		}
	// 		warn++
	// 		return
	// 	}

	// 	if _, err := conn.Write(p); err != nil {
	// 		fmt.Printf("%v\n", err)
	// 	}

	// 	if err := conn.Close(); err != nil {
	// 		fmt.Printf("%v\n", err)
	// 	}
	// }()

	return len(p), nil
}
