package main

import (
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
	listener, err := net.Listen("tcp", lib.Cfg.Log.GetHost())
	if err != nil {
		log.Fatalln(err)
		return
	}
	log.Printf("info listening on %s\n", lib.Cfg.Log.GetHost())

	for {
		sockC := make(chan net.Conn)
		errC := make(chan error)

		go func() {
			sock, err := listener.Accept()
			if err != nil {
				errC <- err
				return
			}
			sockC <- sock
		}()

		select {
		case <-sigC:
			return
		case err := <-errC:
			log.Println(err)
		case so := <-sockC:
			go readSocket(so)
		}
	}
}

func readSocket(sock net.Conn) {
	defer func() {
		if err := sock.Close(); err != nil {
			log.Println(err)
		}
	}()

	b := make([]byte, 1024)
	mLen, err := sock.Read(b)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("%s | %s", sock.RemoteAddr(), b[:mLen])

	// todo write to file???
}
