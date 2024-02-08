package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sfw/lib"
)

var sigC = make(chan os.Signal, 1)
var asyncErrC = make(chan error)

func init() {
	lib.FlagParse()
	signal.Notify(sigC, os.Interrupt)
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("error %v", err)
	}
}

// todo handle sigint
func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go acceptSockets(ctx)

	select {
	case <-ctx.Done():
	case <-sigC:
	case err := <-asyncErrC:
		return err
	}

	return nil
}

func acceptSockets(ctx context.Context) {
	listener, err := net.Listen("tcp", *lib.FlagLogSrv)
	if err != nil {
		<-asyncErrC
		return
	}
	log.Printf("info listening on %s", *lib.FlagLogSrv)

	for {
		soC := make(chan net.Conn)
		go func() {
			so, err := listener.Accept()
			if err != nil {
				asyncErrC <- err
				return
			}
			soC <- so

			var re bytes.Buffer
			io.Copy(&re, so)
			fmt.Printf("%s | %s", so.RemoteAddr(), re.String())
			if err := so.Close(); err != nil {
				log.Printf("%v", err)
			}
		}()

		select {
		case <-ctx.Done():
			return
		// case so := <-soC:
		case <-soC:
		}
	}
}
