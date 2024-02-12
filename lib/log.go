package lib

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"
)

type Logger struct {
	Conn net.Conn
}

var LogErrC = make(chan error, 10)
var reconnectC = make(chan struct{}, 1)
var sigC = make(chan os.Signal, 1)

func init() {
	signal.Notify(sigC, os.Interrupt)
	go dialateLogger()
}

func NewLogger() Logger {
	dialer := net.Dialer{Timeout: 3 * time.Second}
	conn, err := dialer.Dial("tcp", Cfg.Log.GetHost())
	if err != nil {
		LogErrC <- err
	}
	return Logger{Conn: conn}
}

// todo split log files
func (O Logger) Write(p []byte) (n int, err error) {
	fmt.Printf(string(p))

	go func() {
		if O.Conn == nil {
			LogErrC <- errors.New("conn nil")
			return
		}
		if _, err := O.Conn.Write(p); err != nil {
			LogErrC <- err
			return
		}
	}()

	return len(p), nil
}

func dialateLogger() {
	for {
		select {
		case logErr := <-LogErrC:
			cC := make(chan net.Conn)

			go func() {
				if len(reconnectC) > 0 {
					return
				}
				reconnectC <- struct{}{}
				fmt.Printf("%v\n", logErr)

				dialer := net.Dialer{Timeout: 3 * time.Second}
				conn, err := dialer.Dial("tcp", Cfg.Log.GetHost())
				if err != nil {
					for len(LogErrC) > 0 {
						<-LogErrC
					}
					LogErrC <- err
					return
				}
				cC <- conn
			}()

			select {
			case <-sigC:
				return
			case c := <-cC:
				log.SetOutput(Logger{Conn: c})
				// SfwLogger = Logger{Conn: c}
			case <-time.After(3 * time.Second):
			}

			for len(reconnectC) > 0 {
				<-reconnectC
			}

		case <-sigC:
			return
		}
	}
}
