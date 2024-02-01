package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sfw/db"
	"sfw/lib"
	"sfw/ws"
	"strconv"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

var Idle = true
var WorldgenC = make(chan struct{}, 1)
var IdleC = make(chan struct{}, 1)
var ConnErrC = make(chan error, 1)
var SigC = make(chan os.Signal, 1)

var FlagServer = flag.String("s", "127.0.0.1:3100", "server addr")
var Connection *websocket.Conn

func init() {
	if err := initE(); err != nil {
		log.Fatalf("error %v", err)
	}
}

func initE() error {
	flag.Parse()

	signal.Notify(SigC, os.Interrupt)

	err := ConnF()
	if err != nil {
		// return err
		// log.Printf("warning connection failed %v", err)
		ConnErrC <- err
	}
	WorldgenC <- struct{}{}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("error %v", err)
	}
	defer Connection.CloseNow()
}

// todo move to ws
func ConnF() error {
	conn, _, err_ := websocket.Dial(
		context.TODO(),
		fmt.Sprintf("ws://%s", *FlagServer),
		nil,
	)
	if err_ != nil {
		return err_
	}
	Connection = conn
	return nil
}

func run() error {
	log.Printf("1")
MainLoop:
	for {
		select {
		case err := <-ConnErrC:
			log.Printf("warning connection error %v", err)
			for {
				log.Printf("info trying to reconnect")
				if err := ConnF(); err == nil {
					log.Printf("info connected!")
					continue MainLoop
				}

				// ConnErrC <- err
				log.Printf("info retrying in 3 seconds")
				select {
				// todo use ratelimiter?
				case <-time.After(3 * time.Second):
					continue
				case sig := <-SigC:
					log.Printf("terminating: %v", sig)
					goto Stop
				}
			}

		case <-WorldgenC:
			go func() {
				for {
					m := db.GodSeed{}
					if err := wsjson.Read(context.TODO(), Connection, &m); err != nil {
						// log.Printf("warning decode %v", err)
						ConnErrC <- err
						WorldgenC <- struct{}{}
						return
					} else {
						log.Printf("info decoded %v", m)
					}

				RetryWorldgen:
					if err := lib.Worldgen(m, 4); err != nil {
						fmt.Printf(">>> ***** WORLDGEN IS DILATING *****\n")
						fmt.Printf(">>> reason: %v\n", err)
						fmt.Printf(">>> 1) quit or something\n")
						fmt.Printf(">>> Enter) retry\n")
						var action string
						fmt.Scanln(&action)
						actionInt, err := strconv.Atoi(action)
						if err != nil || actionInt < 1 || actionInt > 1 {
							goto RetryWorldgen
						}
					}

					IdleC <- struct{}{}
				}
			}()
			IdleC <- struct{}{}

		case <-IdleC:
			if err := wsjson.Write(context.TODO(), Connection, &ws.NState{
				Foo: "worldgen:idle",
			}); err != nil {
				ConnErrC <- err
			}

		case sig := <-SigC:
			log.Printf("terminating: %v", sig)
			goto Stop
		}
	}

Stop:
	// ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	// defer cancel()

	return Connection.Close(websocket.StatusNormalClosure, "")
}
