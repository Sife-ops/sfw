package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sfw/lib"
)

var FlagPort = flag.Int("p", 3000, "port")

func init() {
	flag.Parse()
}

func main() {
	go func() {
		sigchan := make(chan os.Signal)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan
		log.Printf("info shutting down")
		os.Exit(0)
	}()

	log.Fatal(http.ListenAndServe(
		fmt.Sprintf(":%d", *FlagPort),
		http.HandlerFunc(ServeFactory(routes))),
	)
}

var routes = []route{
	newRoute("POST", "/", generate),
}

func generate(w http.ResponseWriter, r *http.Request) {
	var j struct{ Id string }
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		log.Printf("warning %v", err)
		http.Error(w, "generate", http.StatusInternalServerError)
		return
	}

	log.Println("lmao!")
	log.Println(j)

	lib.Db.Get(`SELECT * from seed WHERE id=$1`, j.Id)
}
