package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"sfw/lib"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// ref https://github.com/nhooyr/websocket/blob/master/internal/examples/echo/server.go
// ref https://www.developer.com/languages/intro-socket-programming-go/

var Connections = map[*websocket.Conn]lib.NState{}
var cubiomesOut = make(chan lib.GodSeed, 20)
var OnMessage = make(chan lib.ConnNState, 10) // todo length idk

var onUpdate = make(chan map[net.Conn]lib.SockState, 10) // todo length idk
var flagServer = flag.String("s", "0.0.0.0:3100", "server addr")
var sockErrC = make(chan error)
var socks = map[net.Conn]lib.SockState{}
var sigC = make(chan os.Signal, 1)

func init() {
	flag.Parse()
	signal.Notify(sigC, os.Interrupt)
}

func main() {
	if err := run2(); err != nil {
		log.Fatal(err)
	}
}

func run2() error {
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
				sockerino, err := listener.Accept()
				if err != nil {
					sockErrC <- err
					return
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
			case <-time.After(500 * time.Millisecond):
				var msg string
				switch {
				case len(cubiomesOut) < 6:
					msg = "start"
					break
				case len(cubiomesOut) > 10:
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
			case update := <-onUpdate:
				log.Printf("info onUpdate %v", update)
				for k, v := range update {
					log.Printf("info onUpdate connection %v", k)
					log.Printf("info onUpdate state %v", v)
					if v.F0 == "cubiomes:output" {
						cubiomesOut <- v.F1
						log.Printf("info onUpdate got cubiomes output %d: %+v", len(cubiomesOut), v.F1)
					}
				}
				break
			case <-ctx.Done():
				return
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

// todo finish migrating this

func run() error {
	// todo use 0.0.0.0
	l, err := net.Listen("tcp", "127.0.0.1:3100")
	if err != nil {
		return err
	}
	log.Printf("listening on http://%v", l.Addr())

	s := &http.Server{
		Handler: fooServer{
			logf: log.Printf,
		},
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}
	errc := make(chan error, 1)
	go func() {
		errc <- s.Serve(l)
	}()

	go func() {
		for {
			<-time.After(500 * time.Millisecond)
			var msg string
			switch {
			case len(cubiomesOut) < 6:
				msg = "start"
				break
			case len(cubiomesOut) > 10:
				msg = "stop"
				break
			default:
				continue
			}
			// log.Printf("info sending %s messages to cubiomes", msg)
			for k, v := range Connections {
				if strings.Contains(v.Foo, "cubiomes") {
					if err := k.Write(context.TODO(), websocket.MessageText, []byte(msg)); err != nil {
						log.Printf("warning write %v", err)
					}
				}
			}
		}
	}()

	go func() {
		for {
			s := <-OnMessage
			switch {
			case s.NState.Foo == "worldgen:idle":
				go func() {
					// todo confirm success
					gs := <-cubiomesOut
					log.Printf("info send cubiomes output, queued %d", len(cubiomesOut))
					if err := wsjson.Write(context.TODO(), s.Conn, gs); err != nil {
						log.Printf("warning wsjson write %v", err)
					}
				}()
				break

			case s.NState.Foo == "worldgen:output":
				log.Printf("info saving worldgen output %v", s.NState.GodSeed)
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
					&s.NState.GodSeed,
				); err != nil {
					log.Fatalf("error saving worldgen output %v", err)
				}
				break

			case s.NState.Foo == "cubiomes:output":
				cubiomesOut <- s.NState.GodSeed
				log.Printf("info saving cubiomes output, %d queued", len(cubiomesOut))
				// log.Printf("info saving cubiomes results %v", godSeed)
				if _, err := lib.Db.NamedExec(
					`INSERT INTO seed 
						(seed, spawn_x, spawn_z, bastion_x, bastion_z, shipwreck_x, shipwreck_z, fortress_x, fortress_z, finished_cubiomes)
					VALUES 
						(:seed, :spawn_x, :spawn_z, :bastion_x, :bastion_z, :shipwreck_x, :shipwreck_z, :fortress_x, :fortress_z, :finished_cubiomes)`,
					&s.NState.GodSeed,
				); err != nil {
					log.Fatalf("error saving cubiomes output %s", err.Error())
				}
				break

			default:
				log.Printf("info did nothing %s", s.NState.Foo)
			}
		}
	}()

	// // todo internal state monitoring
	// go func() {
	// 	// return // disabled
	// 	for {
	// 		<-time.After(1 * time.Second)
	// 		log.Printf("info cubiomes out %d", len(cubiomesOut))
	// 		continue
	// 		log.Printf("info Connections %v", Connections)
	// 		for k, v := range Connections {
	// 			log.Printf("conn %v state %v", &k, v)
	// 		}
	// 	}
	// }()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	select {
	case err := <-errc:
		log.Printf("failed to serve: %v", err)
	case sig := <-sigs:
		log.Printf("terminating: %v", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	return s.Shutdown(ctx)
}

type fooServer struct {
	// logf controls where logs are sent.
	logf func(f string, v ...interface{})
}

func (s fooServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		Subprotocols: []string{"foo"},
	})
	if err != nil {
		s.logf("%v", err)
		return
	}

	Connections[c] = lib.NState{}

	for {
		m := lib.NState{}
		if err := wsjson.Read(context.TODO(), c, &m); err != nil {
			log.Printf("warning wsjson read %v", err)
			break
		}
		Connections[c] = m
		OnMessage <- lib.ConnNState{Conn: c, NState: m}
	}

	// todo this may not be needed
	log.Printf("info closing connection %v", &c)
	if err := c.CloseNow(); err != nil {
		log.Printf("warning close %v", err)
	}

	log.Printf("info deleting connection %v", &c)
	delete(Connections, c)
}
