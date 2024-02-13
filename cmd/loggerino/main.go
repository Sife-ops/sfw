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

func init() {
	signal.Notify(sigC, os.Interrupt)
}

func main() {
	fmt.Println("starting loggerino")
	if err := run(); err != nil {
		log.Fatalf("error %v", err)
	}
}

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
		soC := make(chan net.Conn)
		errC := make(chan error)

		go func() {
			so, err := listener.Accept()
			if err != nil {
				errC <- err
				return
			}
			soC <- so
		}()

		select {
		case <-ctx.Done():
			return
		case err := <-errC:
			fmt.Println(err)
		case so := <-soC:
			go readSocket(ctx, so)
		}
	}
}

func readSocket(ctx context.Context, so net.Conn) {
	defer func() {
		if err := so.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	b := make([]byte, 1024)
	mLen, err := so.Read(b)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%s | %s", so.RemoteAddr(), b[:mLen])

	// todo write to file???
}
