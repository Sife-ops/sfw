package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"os/signal"
	"sfw/lib"
	"time"
)

var cubiomesC = make(chan error, 1)
var idle = true
var idleC = make(chan error, 1)
var sigC = make(chan os.Signal, 1)
var sockClient = lib.SockClient{}
var sockErrC = make(chan error, 1)
var threadsC chan struct{}

func init() {
	lib.FlagParse()
	threadsC = make(chan struct{}, *lib.FlagThreads)

	signal.Notify(sigC, os.Interrupt)

	err := sockClient.Connect(*lib.FlagWorker)
	if err != nil {
		sockErrC <- err
	}

	cubiomesC <- errors.New("startup")
	idleC <- errors.New("startup")
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("error %v", err)
	}
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())

	for {
		if len(sockErrC) > 0 {
			// for len(sockErrC) > 0 {
			log.Printf("warning connection error %v", <-sockErrC)
			// }
			log.Printf("info retrying in 3 seconds")
			select {
			case <-time.After(3 * time.Second):
				log.Printf("info trying to reconnect")
				if err := sockClient.Connect(*lib.FlagWorker); err != nil {
					// where are more errors coming from...
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
		case <-cubiomesC:
			for len(threadsC) < *lib.FlagThreads {
				go func() {
					threadsC <- struct{}{}
					gsC := make(chan lib.GodSeed, 1)
					for {
						if idle {
							<-time.After(1 * time.Second)
							continue
						}

						select {
						case <-ctx.Done():
							return

						case gs := <-gsC:
							m := lib.SockState{
								F0: "cubiomes:output",
								F1: gs,
							}
							j, err := json.Marshal(m)
							if err != nil {
								goto EndError
							}
							_, err = sockClient.Conn.Write(j)
							if err != nil {
								goto EndError
							}
							continue

						EndError:
							<-threadsC
							sockErrC <- err
							cubiomesC <- err
							// idle = true
							return

						default:
							gs, err := lib.Cubiomes()
							// todo check exit code
							if err != nil {
								continue
							}
							gsC <- gs
						}
					}
				}()
			}

		case <-idleC:
			go func() {
				m := lib.SockState{
					F0: "cubiomes",
				}
				j, err := json.Marshal(m)
				if err != nil {
					sockErrC <- err
					idleC <- err
					return
				}
				sockClient.Conn.Write(j)

				for {
					readC := make(chan struct {
						b    []byte
						mLen int
					})
					go func() {
						b := make([]byte, 1024)
						mLen, err := sockClient.Conn.Read(b)
						if err != nil {
							sockErrC <- err
							idleC <- err
							return
						}
						readC <- struct {
							b    []byte
							mLen int
						}{b, mLen}
					}()

					select {
					case <-ctx.Done():
						return
					case r := <-readC:
						msg := string(r.b[:r.mLen])
						// log.Printf("info got msg from %v: %s", sockClient.Conn.RemoteAddr(), msg)
						switch {
						case msg == "start" && idle:
							log.Printf("info changing state to active")
							idle = false
						case msg == "stop" && !idle:
							log.Printf("info changing state to idle")
							idle = true
						}
					}
				}
			}()
			break

		case sig := <-sigC:
			log.Printf("terminating: %v", sig)
			goto End
		}
	}

End:
	cancel()
	return sockClient.Conn.Close()
}
