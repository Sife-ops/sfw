package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"
	"sfw/lib"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

var flagServer = flag.String("s", "127.0.0.1:3100", "server addr")
var flagThreads = flag.Int("t", 1, "threads")

var idle = true
var connErrC = make(chan error, 1)
var cubiomesC = make(chan error, 1) // todo use err?
var idleC = make(chan error, 1)
var sigC = make(chan os.Signal, 1)
var threadsC chan struct{}

func init() {
	flag.Parse()
	threadsC = make(chan struct{}, *flagThreads)

	signal.Notify(sigC, os.Interrupt)

	err := lib.Dial(flagServer)
	if err != nil {
		connErrC <- err
	}

	cubiomesC <- errors.New("startup")
	idleC <- errors.New("startup")
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("error %v", err)
	}
	defer lib.Ws.CloseNow() // todo nil ref
}

func run() error {
	for {
		if len(connErrC) > 0 {
			log.Printf("warning connection error %v", <-connErrC)
			log.Printf("info retrying in 3 seconds")
			select {
			case <-time.After(3 * time.Second):
				log.Printf("info trying to reconnect")
				if err := lib.Dial(flagServer); err != nil {
					connErrC <- err
					continue
				} else {
					log.Printf("info connected!")
					continue
				}
			case sig := <-sigC:
				log.Printf("terminating: %v", sig)
				goto Stop
			}
		}

		select {
		case <-cubiomesC:
			for len(threadsC) < *flagThreads {
				go func() {
					threadsC <- struct{}{}
					for {
						if idle {
							<-time.After(1 * time.Second)
							continue
						}

						var gs lib.GodSeed
						for {
							gs_, err := lib.Cubiomes()
							// todo check exit code
							if err != nil {
								continue
							}
							gs = gs_
							break
						}

						if err := wsjson.Write(context.TODO(), lib.Ws, &lib.NState{
							Foo:     "cubiomes:output",
							GodSeed: gs,
						}); err != nil {
							// todo check queue length
							<-threadsC
							connErrC <- err
							cubiomesC <- err
							idle = true
							return
						}
					}
				}()
			}

		case <-idleC:
			go func() {
				for {
					_, b, err := lib.Ws.Read(context.TODO())
					if err != nil {
						connErrC <- err
						idleC <- err
						return
					}
					switch {
					case string(b) == "start" && idle:
						log.Printf("info changing state to active")
						idle = false
					case string(b) == "stop" && !idle:
						log.Printf("info changing state to idle")
						idle = true
					}
				}
			}()
			if err := wsjson.Write(context.TODO(), lib.Ws, &lib.NState{
				Foo: "cubiomes",
			}); err != nil {
				connErrC <- err
			}

		case sig := <-sigC:
			log.Printf("terminating: %v", sig)
			goto Stop
		}
	}

Stop:
	// ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	// defer cancel()

	return lib.Ws.Close(websocket.StatusNormalClosure, "")
}
