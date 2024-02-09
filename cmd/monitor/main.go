package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sfw/lib"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

var asyncErrC = make(chan error)
var sigC = make(chan os.Signal, 1)

func init() {
	lib.FlagParse()
	signal.Notify(sigC, os.Interrupt)
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("%v", err)
	}
}

func run() error {
	log.Printf("starting monitor")

	server := http.Server{
		Addr:    *lib.FlagWsSrv,
		Handler: http.HandlerFunc(acceptWebsocket),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			asyncErrC <- err
		}
	}()

	select {
	case err := <-asyncErrC:
		// todo return nil if ErrServerClosed?
		return err
	case <-sigC:
		if err := server.Shutdown(context.TODO()); err != nil {
			return err
		}
		return nil
	}
}

func acceptWebsocket(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Printf("%v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var m struct {
		Hostname string
		Message  string
	}
	if err := wsjson.Read(ctx, conn, &m); err != nil {
		log.Printf("%v", err)
	}

	fmt.Printf("%s | %s", m.Hostname, m.Message)

	if err := conn.CloseNow(); err != nil {
		log.Printf("%v", err)
	}
}
