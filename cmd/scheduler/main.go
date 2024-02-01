package main

import (
	"context"
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

var Connections = map[*websocket.Conn]lib.NState{}
var OnMessage = make(chan lib.ConnNState, 10) // todo length idk
var CubiomesOut = make(chan lib.GodSeed, 20)

func run() error {
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
			case len(CubiomesOut) < 6:
				msg = "start"
				break
			case len(CubiomesOut) > 10:
				msg = "stop"
				break
			default:
				continue
			}
			// log.Printf("info sending %s messages to cubiomes", msg)
			for k, v := range Connections {
				if strings.Contains(v.Foo, "cubiomes") {
					if err := k.Write(context.TODO(), websocket.MessageText, []byte(msg)); err != nil {
						log.Printf("warning starting problem %v", err)
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
					gs := <-CubiomesOut
					log.Printf("info send cubiomes output, queued %d", len(CubiomesOut))
					if err := wsjson.Write(context.TODO(), s.Conn, gs); err != nil {
						log.Printf("todo REEEE %v", err)
						return
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
				CubiomesOut <- s.NState.GodSeed
				log.Printf("info saving cubiomes output, %d queued", len(CubiomesOut))
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
	// 		log.Printf("info cubiomes out %d", len(CubiomesOut))
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
			log.Printf("info LOOOOOL %v", err)
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

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
