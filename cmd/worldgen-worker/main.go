package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sfw/lib"
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

var flagServer = flag.String("s", "127.0.0.1:3100", "server addr")

func init() {
	flag.Parse()

	signal.Notify(SigC, os.Interrupt)

	err := lib.Dial(flagServer)
	if err != nil {
		// return err
		log.Printf("warning connection failed %v", err)
		ConnErrC <- err
	}

	WorldgenC <- struct{}{}
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("error %v", err)
	}
	defer lib.Ws.CloseNow()
}

func run() error {
	// MainLoop:
	for {
		if len(ConnErrC) > 0 {
			log.Printf("warning connection error %v", <-ConnErrC)
			log.Printf("info retrying in 3 seconds")
			select {
			case <-time.After(3 * time.Second):
				log.Printf("info trying to reconnect")
				if err := lib.Dial(flagServer); err != nil {
					ConnErrC <- err
					continue
				} else {
					log.Printf("info connected!")
					continue
				}
			case sig := <-SigC:
				log.Printf("terminating: %v", sig)
				goto Stop
			}
		}

		select {
		case <-WorldgenC:
			go func() {
				for {
					cs := lib.GodSeed{}
					if err := wsjson.Read(context.TODO(), lib.Ws, &cs); err != nil {
						// log.Printf("warning decode %v", err)
						ConnErrC <- err
						WorldgenC <- struct{}{}
						return
					} else {
						log.Printf("info decoded %v", cs)
					}

				RetryWorldgen:
					gs, err := lib.Worldgen(cs, 4)
					if err != nil {
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

					// log.Printf("info THIS IS THE RESULT %v", gs)
					if err := wsjson.Write(context.TODO(), lib.Ws, &lib.NState{
						Foo:     "worldgen:output",
						GodSeed: gs,
					}); err != nil {
						ConnErrC <- err
					}

					IdleC <- struct{}{}
				}
			}()
			IdleC <- struct{}{}

		case <-IdleC:
			if err := wsjson.Write(context.TODO(), lib.Ws, &lib.NState{
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

	return lib.Ws.Close(websocket.StatusNormalClosure, "")
}
