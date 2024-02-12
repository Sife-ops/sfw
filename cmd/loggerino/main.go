package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sfw/lib"
)

var asyncErrC = make(chan error)
var sigC = make(chan os.Signal, 1)
var socksM = make(map[net.Conn]bool)

func init() {
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

	go runAsync(ctx)

	select {
	case <-sigC:
	case err := <-asyncErrC:
		return err
	}

	return nil
}

func runAsync(ctx context.Context) {
	listener, err := net.Listen("tcp", lib.Cfg.Log.GetHost())
	if err != nil {
		<-asyncErrC
		return
	}
	log.Printf("info listening on %s", lib.Cfg.Log.GetHost())

	for {
		soC := make(chan net.Conn, 1)
		go func() {
			so, err := listener.Accept()
			if err != nil {
				asyncErrC <- err
				return
			}
			soC <- so
		}()

		select {
		case <-ctx.Done():
			return
		case so := <-soC:
			socksM[so] = true
			fmt.Printf("%s | connected\n", so.RemoteAddr())
			go readSocket(ctx, so)
		}
	}
}

func readSocket(ctx context.Context, so net.Conn) {
	for {
		done := make(chan bool)
		errC := make(chan error)
		go func() {
			b := make([]byte, 1024)
			mLen, err := so.Read(b)
			if err != nil {
				errC <- err
			}
			fmt.Printf("%s | %s", so.RemoteAddr(), b[:mLen])
			done <- true
		}()

		select {
		case <-ctx.Done():
			return
		case err := <-errC:
			log.Printf("%v", err)
			delete(socksM, so)
			return
		case <-done:
		}
	}
}
