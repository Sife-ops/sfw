package main

import (
	"context"
	"embed"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"sfw/lib"
	"strings"
	"text/template"
)

var asyncErrC = make(chan error)
var sigC = make(chan os.Signal, 1)

func init() {
	signal.Notify(sigC, os.Interrupt)
	lib.FlagParse()
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("%v", err)
	}
}

func run() error {
	s := http.Server{
		Addr:    *lib.FlagWebSrv,
		Handler: http.HandlerFunc(serve),
	}

	go func() {
		if err := s.ListenAndServe(); err != nil {
			asyncErrC <- err
		}
	}()

	select {
	case err := <-asyncErrC:
		return err
	case <-sigC:
		return nil
	}
}

///////////////////////////////////////////////////////////////////////////////

// https://github.com/benhoyt/go-routing/blob/9a2fa7a643ecb5681f504b95064d948ee2177c9a/retable/route.go

type ctxKey struct{}

var routes = []route{
	newRoute("GET", "/", root),
}

func newRoute(method, pattern string, handler http.HandlerFunc) route {
	return route{method, regexp.MustCompile("^" + pattern + "$"), handler}
}

type route struct {
	method  string
	regex   *regexp.Regexp
	handler http.HandlerFunc
}

func serve(w http.ResponseWriter, r *http.Request) {
	var allow []string
	for _, route := range routes {
		matches := route.regex.FindStringSubmatch(r.URL.Path)
		if len(matches) > 0 {
			if r.Method != route.method {
				allow = append(allow, route.method)
				continue
			}
			ctx := context.WithValue(r.Context(), ctxKey{}, matches[1:])
			route.handler(w, r.WithContext(ctx))
			return
		}
	}
	if len(allow) > 0 {
		w.Header().Set("Allow", strings.Join(allow, ", "))
		http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.NotFound(w, r)
}

//go:embed template
var fs embed.FS

///////////////////////////////////////////////////////////////////////////////

func root(w http.ResponseWriter, r *http.Request) {
	seeds := []lib.GodSeed{}
	if err := lib.Db.Select(&seeds,
		`SELECT *
		FROM seed`,
	); err != nil {
		log.Printf("0 %v", err)
		http.Error(w, "database", http.StatusInternalServerError)
		return
	}
	// log.Printf("%v", gs)

	t, err := template.New("root.html").ParseFS(fs, "template/root.html")
	if err != nil {
		log.Printf("1 %v", err)
		http.Error(w, "template", http.StatusInternalServerError)
		return
	}
	if err := t.Execute(w, map[string]interface{}{
		"seeds": seeds,
	}); err != nil {
		log.Printf("2 %v", err)
		http.Error(w, "template", http.StatusInternalServerError)
		return
	}
}
