package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sfw/lib"
	"strconv"
	"time"
)

var startWgC = make(chan lib.GodSeed, 1)
var startRlC = make(chan error, 1)
var sendIdleC = make(chan struct{}, 1)
var sigC = make(chan os.Signal, 1)
var sockClient = lib.SockClient{}
var sockErrC = make(chan error, 1)
var flagServer = flag.String("s", "127.0.0.1:3100", "server addr")

func init() {
	flag.Parse()

	signal.Notify(sigC, os.Interrupt)

	if err := sockClient.Connect(*flagServer); err != nil {
		sockErrC <- err
	}

	startRlC <- errors.New("startup")
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("error %v", err)
	}
}

func run() error {
	for {
		if len(sockErrC) > 0 {
			log.Printf("warning connection error %v", <-sockErrC)
			log.Printf("info retrying in 3 seconds")
			select {
			case <-time.After(3 * time.Second):
				log.Printf("info trying to reconnect")
				if err := sockClient.Connect(*flagServer); err != nil {
					for len(sockErrC) > 0 {
						log.Printf("info REEEEEE %v", <-sockErrC)
					}
					sockErrC <- err
					continue
				} else {
					log.Printf("info connected!")
					continue
				}
			case sig := <-sigC:
				log.Printf("terminating: %v", sig)
				goto End
			}
		}

		select {
		case <-startRlC:
			go func() {
				for {
					b := make([]byte, 1024)
					mLen, err := sockClient.Conn.Read(b)
					if err != nil {
						sockErrC <- err
						startRlC <- err
						return
					}
					cs := lib.GodSeed{}
					if err := json.Unmarshal(b[:mLen], &cs); err != nil {
						sockErrC <- err
						startRlC <- err
						return
					}
					// log.Printf("info decoded %v", cs)
					if len(startWgC) < 1 {
						startWgC <- cs
					}
				}
			}()
			sendIdleC <- struct{}{}

		case cs := <-startWgC:
			gsC := make(chan lib.GodSeed, 1)
			go func() {
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
				gsC <- gs
			}()

			select {
			case <-sigC:
				goto End
			case gs := <-gsC:
				j, err := json.Marshal(lib.SockState{
					F0: "worldgen:idle",
					F1: gs,
				})
				if err != nil {
					sockErrC <- err
					startRlC <- err
					continue
				}
				_, err = sockClient.Conn.Write(j)
				if err != nil {
					sockErrC <- err
					startRlC <- err
					continue
				}
			}

		case <-sendIdleC:
			j, err := json.Marshal(lib.SockState{
				F0: "worldgen:idle",
			})
			if err != nil {
				sockErrC <- err
				startRlC <- err
				break
			}
			_, err = sockClient.Conn.Write(j)
			if err != nil {
				sockErrC <- err
				startRlC <- err
			}

		case sig := <-sigC:
			log.Printf("terminating: %v", sig)
			goto End
		}
	}

End:
	return sockClient.Conn.Close()
}
