package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sfw/lib"
)

var monitorLog *os.File
var asyncErrC = make(chan error)
var sigC = make(chan os.Signal, 1)

func init() {
	signal.Notify(sigC, os.Interrupt)
}

func main() {
	var err error
	monitorLog, err = os.OpenFile("./tmp/monitor-log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		if err := monitorLog.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	listener, err := net.Listen("tcp", lib.Cfg.Log.GetHost())
	if err != nil {
		log.Fatalln(err)
		return
	}
	defer func() {
		if err := listener.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

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

	line := fmt.Sprintf("%s | %s", sock.RemoteAddr(), b[:mLen])
	fmt.Print(line)

	if _, err := monitorLog.Write([]byte(line)); err != nil {
		log.Println(err)
		return
	}
}
