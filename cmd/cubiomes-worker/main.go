package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sfw/lib"
	"sfw/ws"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

var Idle = true

var FlagServer = flag.String("s", "127.0.0.1:3100", "server addr")

var Connection *websocket.Conn

func init() {
	if err := initE(); err != nil {
		log.Fatalf("error %v", err)
	}
}

func initE() error {
	flag.Parse()
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("error %v", err)
	}
	defer Connection.CloseNow()
}

func run() error {
	connErrC := make(chan error, 1)
	if Connection == nil {
		connErrC <- fmt.Errorf("error connection nil")
	}

	// todo use connErrC
	// todo for threads?
	go func() {
		for {
			if Idle {
				<-time.After(1 * time.Second)
				continue
			}

			gs, err := lib.Cubiomes()
			if err != nil {
				continue
			}

			if err := wsjson.Write(context.TODO(), Connection, &ws.NState{
				Foo:     "cubiomes:output",
				GodSeed: gs,
			}); err != nil {
				// todo check queue length
				connErrC <- err
				Idle = true
			}
		}
	}()

	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, os.Interrupt)

	for {
		select {
		case err := <-connErrC:
			log.Printf("warning connection error %v", err)
			Idle = true

			// todo move to ws?
			conn, _, err_ := websocket.Dial(
				context.TODO(),
				fmt.Sprintf("ws://%s", *FlagServer),
				nil,
			)
			if err_ != nil {
				connErrC <- err_
				goto Retry
			}
			Connection = conn
			if err := wsjson.Write(context.TODO(), Connection, &ws.NState{
				Foo: "cubiomes",
			}); err != nil {
				connErrC <- err
				goto Retry
			}
			goto Connected

		Retry:
			log.Printf("info retrying in 3 seconds")
			select {
			// todo use ratelimiter?
			case <-time.After(3 * time.Second):
				continue
			case sig := <-sigC:
				log.Printf("terminating: %v", sig)
				goto Stop
			}

		Connected:
			log.Printf("info connected!")
			// todo: context cancel
			go func() {
				for {
					_, b, err := Connection.Read(context.TODO())
					if err != nil {
						connErrC <- err
						return
					}
					switch {
					case string(b) == "start" && Idle:
						log.Printf("info changing state to active")
						Idle = false
					case string(b) == "stop" && !Idle:
						log.Printf("info changing state to idle")
						Idle = true
					}
				}
			}()

		case sig := <-sigC:
			log.Printf("terminating: %v", sig)
			goto Stop
		}
	}

Stop:
	// ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	// defer cancel()

	return Connection.Close(websocket.StatusNormalClosure, "")
}
