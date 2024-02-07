package main

import (
	"fmt"
	"log"
	"net"
	"sfw/lib"
)

func init() {
	lib.FlagParse()
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("error %v", err)
	}
}

func run() error {
	listener, err := net.Listen("tcp", *lib.FlagLogSrv)
	if err != nil {
		return err
	}
	log.Printf("info listening on %s", *lib.FlagLogSrv)

	for {
		so, err := listener.Accept()
		if err != nil {
			return err
		}
		go func() {
			defer so.Close()
			re := make([]byte, 1024)
			len, err := so.Read(re)
			if err != nil {
				log.Printf("warning error reading socket %v", err)
				return
			}
			fmt.Printf("%s | %s\n", so.RemoteAddr(), string(re[:len]))
		}()
	}

	// return nil
}
