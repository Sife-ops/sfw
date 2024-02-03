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

var flagServer = flag.String("s", "127.0.0.1:3100", "server addr")
var rlStartC = make(chan error, 1)
var sendIdleC = make(chan lib.GodSeed, 1)
var sigC = make(chan os.Signal, 1)
var sockClient = lib.SockClient{}
var sockErrC = make(chan error, 1)
var wbBusyC = make(chan error, 1)
var wgStartC = make(chan lib.GodSeed, 1)

func init() {
	flag.Parse()
	signal.Notify(sigC, os.Interrupt)
	if err := sockClient.Connect(*flagServer); err != nil {
		sockErrC <- err
	}
	rlStartC <- errors.New("startup")
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
		case <-rlStartC:
			go func() {
				for {
					b := make([]byte, 1024)
					mLen, err := sockClient.Conn.Read(b)
					if err != nil {
						sockErrC <- err
						rlStartC <- err
						return
					}
					cs := lib.GodSeed{}
					if err := json.Unmarshal(b[:mLen], &cs); err != nil {
						sockErrC <- err
						rlStartC <- err
						return
					}
					// log.Printf("info decoded %v", cs)
					if len(wgStartC) < 1 {
						wgStartC <- cs
					}
				}
			}()
			sendIdleC <- lib.GodSeed{}

		case cs := <-wgStartC:
			wbBusyC <- errors.New("busy")
			gsC := make(chan lib.GodSeed, 1)

			go func() {
			RetryWorldgen:
				gs, err := lib.Worldgen(cs, 4)
				if err != nil {
					fmt.Printf(">>> ***** WORLDGEN IS DILATING *****\n")
					fmt.Printf(">>> reason: %v\n", err)
					fmt.Printf(">>> 1) next\n")
					fmt.Printf(">>> Enter) retry\n")

					action := make(chan string)
					go func() {
						var a string
						fmt.Scanln(&a)
						action <- a
					}()

					select {
					case <-time.After(30 * time.Second):
						gsC <- lib.GodSeed{}
						return
					case a := <-action:
						aInt, err := strconv.Atoi(a)
						if err != nil {
							goto RetryWorldgen
						}
						switch {
						case aInt == 1:
							gsC <- lib.GodSeed{}
							return
						default:
							goto RetryWorldgen
						}
					}
				}
				gsC <- gs
			}()

			select {
			case <-sigC:
				goto End
			case gs := <-gsC:
				<-wbBusyC
				sendIdleC <- gs
			}

		case <-time.After(5 * time.Second):
			sendIdleC <- lib.GodSeed{}

		case gs := <-sendIdleC:
			if len(wbBusyC) > 0 {
				break
			}

			j, err := json.Marshal(lib.SockState{
				F0: "worldgen:idle",
				F1: gs,
			})
			if err != nil {
				sockErrC <- err
				rlStartC <- err
				break
			}
			_, err = sockClient.Conn.Write(j)
			if err != nil {
				sockErrC <- err
				rlStartC <- err
			}

		case sig := <-sigC:
			log.Printf("terminating: %v", sig)
			goto End
		}
	}

End:
	return sockClient.Conn.Close()
}
