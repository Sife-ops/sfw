package lib

import (
	"fmt"
	"net"
	"os"
	"time"
)

// todo throttle error output

type FileLogger struct{}

func (O FileLogger) Write(p []byte) (n int, err error) {
	f, err := os.OpenFile(fmt.Sprintf("/tmp/sfw_nfs/log/%s", Cfg.Wgip), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	n, err = f.Write(p)
	if err != nil {
		fmt.Println(err)
	}

	return
}

///////////////////////////////////////////////////////////////////////////////

var sockDialer = net.Dialer{Timeout: 3 * time.Second}

type SockLogger struct{}

func (O SockLogger) Write(p []byte) (n int, err error) {
	conn, err := sockDialer.Dial("tcp", Cfg.Log.GetHost())
	if err != nil {
		fmt.Println(err)
		return
	}

	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	n, err = conn.Write(p)
	if err != nil {
		fmt.Println(err)
	}

	return
}
