package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"sfw/lib"
	"strings"
	"time"
)

// ref https://www.developer.com/languages/intro-socket-programming-go/

var onUpdate = make(chan map[net.Conn]lib.SockState, 10) // todo length idk
var flagServer = flag.String("s", "0.0.0.0:3100", "server addr")
var sockErrC = make(chan error, 1)
var socks = map[net.Conn]lib.SockState{}
var sigC = make(chan os.Signal, 1)

func init() {
	flag.Parse()
	signal.Notify(sigC, os.Interrupt)
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	log.Printf("info listening on %s", *flagServer)
	listener, err := net.Listen("tcp", *flagServer)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			sockC := make(chan net.Conn)
			go func() {
			Retry:
				sockerino, err := listener.Accept()
				if err != nil {
					// sockErrC <- err
					// return
					log.Printf("warning listener error %v", err)
					goto Retry
				}
				log.Printf("info new socket connection %v", sockerino)
				sockC <- sockerino
			}()

			select {
			case sockerino := <-sockC:
				socks[sockerino] = lib.SockState{F0: "connected"}
				log.Printf("info socket %v state %v", sockerino, socks[sockerino])
				onUpdate <- map[net.Conn]lib.SockState{sockerino: socks[sockerino]}
				go rl(ctx, sockerino)
				break

			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		for {
			select {
			// todo how to stop dilating
			case <-time.After(3 * time.Second):
				gss := []lib.GodSeed{}
				if err := lib.Db.Select(&gss,
					`SELECT * 
					FROM seed 
					WHERE finished_worldgen IS NULL`,
				); err != nil {
					log.Fatalf("error db %v", err)
				}

				var msg string
				switch {
				case len(gss) < 6:
					msg = "start"
					break
				case len(gss) > 9:
					msg = "stop"
					break
				default:
					continue
				}

				for k, v := range socks {
					if strings.Contains(v.F0, "cubiomes") {
						if _, err := k.Write([]byte(msg)); err != nil {
							sockErrC <- err
							delete(socks, k)
							continue
						}
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case update := <-onUpdate:
				log.Printf("info onUpdate %v", update)
				for k, v := range update {
					log.Printf("info onUpdate connection %v", k)
					log.Printf("info onUpdate state %v", v)
					switch v.F0 {
					case "cubiomes:output":
						// log.Printf("info saving cubiomes output, %d queued", len(cubiomesOut))
						if _, err := lib.Db.NamedExec(
							`INSERT INTO seed 
								(seed, spawn_x, spawn_z, bastion_x, bastion_z, shipwreck_x, shipwreck_z, fortress_x, fortress_z, finished_cubiomes)
							VALUES 
								(:seed, :spawn_x, :spawn_z, :bastion_x, :bastion_z, :shipwreck_x, :shipwreck_z, :fortress_x, :fortress_z, :finished_cubiomes)`,
							&v.F1,
						); err != nil {
							// todo remove fatals??
							log.Fatalf("error onUpdate saving cubiomes output %s", err.Error())
						}
						// cubiomesOut <- v.F1
						// log.Printf("info onUpdate saved cubiomes output, queued %d", len(cubiomesOut))
						break

					case "worldgen:idle":
						// todo json omit empty
						if v.F1.Seed != nil {
							if _, err := lib.Db.NamedExec(
								`UPDATE 
									seed 
								SET 
									ravine_chunks=:ravine_chunks,
									iron_shipwrecks=:iron_shipwrecks,
									ravine_proximity=:ravine_proximity,
									avg_bastion_air=:avg_bastion_air,
									finished_worldgen=1 
								WHERE 
									seed=:seed`,
								&v.F1,
							); err != nil {
								log.Fatalf("error onUpdate updating record %v", err)
							}
							log.Printf("info onUpdate updated record %v", v.F1)
						}

						gss := []lib.GodSeed{}
						if err := lib.Db.Select(&gss,
							`SELECT * 
							FROM seed 
							WHERE finished_worldgen IS NULL`,
						); err != nil {
							log.Fatalf("error db %v", err)
						}
						log.Printf("info %d worldgen job remaining", len(gss))

						if len(gss) < 1 {
							break
						}

						j, err := json.Marshal(gss[0])
						if err != nil {
							sockErrC <- err
							break
						}
						if _, err := k.Write(j); err != nil {
							sockErrC <- err
							break
						}

						if _, err := lib.Db.Exec(
							`UPDATE seed
							SET finished_worldgen=0
							WHERE seed=$1`,
							gss[0].Seed,
						); err != nil {
							log.Fatalf("error db %v", err)
						}
					}
				}
			}
		}
	}()

	for {
		select {
		case sockErr := <-sockErrC:
			log.Printf("warning socket error %v", sockErr)
			break
		// case <-time.After(1 * time.Second):
		// 	log.Printf("info debug socks %v", socks)
		case sig := <-sigC:
			log.Printf("terminating: %v", sig)
			goto End
		}
	}

End:
	cancel()
	if err := listener.Close(); err != nil {
		return err
	}
	return nil
}

func rl(ctx context.Context, s net.Conn) {
	for {
		readC := make(chan struct {
			b    []byte
			mLen int
		})
		go func() {
			b := make([]byte, 1024)
			mLen, err := s.Read(b)
			if err != nil {
				sockErrC <- err
				delete(socks, s)
				return
			}
			readC <- struct {
				b    []byte
				mLen int
			}{b, mLen}
		}()

		select {
		case r := <-readC:
			j := lib.SockState{}
			if err := json.Unmarshal(r.b[:r.mLen], &j); err != nil {
				sockErrC <- err
				continue
			}
			log.Printf("info got msg from %v: %s", s.RemoteAddr(), string(r.b[:r.mLen]))
			socks[s] = j
			log.Printf("info socket %v state %v", s, socks[s])
			onUpdate <- map[net.Conn]lib.SockState{s: socks[s]}
			break

		case <-ctx.Done():
			return
		}
	}
}
