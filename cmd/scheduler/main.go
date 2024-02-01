package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"sfw/db"
	"sfw/ws"

	"nhooyr.io/websocket"
)

// ref https://github.com/nhooyr/websocket/blob/master/internal/examples/echo/server.go

var Connections = map[*websocket.Conn]ws.NState{}
var OnMessage = make(chan ws.ConnNState, 10) // todo length idk
var CubiomesOut = make(chan db.GodSeed, 20)

func run() error {

	// //
	// CubiomesOut <- db.GodSeed{
	// 	Id: 123,
	// }
	// CubiomesOut <- db.GodSeed{
	// 	Id: 456,
	// }
	// //

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

	// todo on update
	go func() {
		for {
			s := <-OnMessage
			// log.Printf("info update %v", s)
			switch {
			case s.NState.Foo == "worldgen:idle":
				go func() {
					// todo confirm success
					gs := <-CubiomesOut
					log.Printf("info send cubiomes output, queued %d", len(CubiomesOut))
					// todo use wsjson.Write
					w, err := s.Conn.Writer(context.TODO(), websocket.MessageText)
					if err != nil {
						log.Printf("warning writer %v", err)
						return
					}
					enc := json.NewEncoder(w)
					if err := enc.Encode(gs); err != nil {
						log.Printf("warning encode %v", err)
						return
					}
					if err := w.Close(); err != nil {
						log.Printf("warning writer close %v", err)
						return
					}
				}()
				break
			case s.NState.Foo == "worldgen:output":
				break
			case s.NState.Foo == "cubiomes:output":
				CubiomesOut <- s.NState.GodSeed
				log.Printf("info recv cubiomes output, queued %d", len(CubiomesOut))
				// todo insert
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

	Connections[c] = ws.NState{}

	for {
		_, r, err := c.Reader(context.TODO())
		if err != nil {
			log.Printf("warning reader sumting wong! %v", err)
			break
		}
		// log.Printf("info message type %v", typ)

		dec := json.NewDecoder(r)
		m := ws.NState{}
		if err := dec.Decode(&m); err != nil {
			log.Printf("warning decode %v", err)
			break
		}
		// log.Printf("info state %v", m)
		Connections[c] = m
		OnMessage <- ws.ConnNState{Conn: c, NState: m}
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
