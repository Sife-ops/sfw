package main

import (
	"bytes"
	"fmt"
	"io"
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
			var re bytes.Buffer
			io.Copy(&re, so)
			fmt.Printf("%s | %s", so.RemoteAddr(), re.String())
			if err := so.Close(); err != nil {
				log.Printf("%v", err)
			}
		}()
	}

	// return nil
}
