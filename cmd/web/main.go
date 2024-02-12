package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"sfw/gen/config"
	"strings"
	"text/template"
)

var asyncErrC = make(chan error)
var sigC = make(chan os.Signal, 1)

func init() {
	signal.Notify(sigC, os.Interrupt)
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("%v", err)
	}
}

func run() error {
	log.Printf("info starting gud web server")

	cfg, err := config.LoadFromPath(context.Background(), "./config.pkl")
	if err != nil {
		return err
	}

	// log.Printf("debug %v", cfg.Lmao.GetHost())

	for {
		s := http.Server{
			Addr:    cfg.Web.GetHost(),
			Handler: http.HandlerFunc(serve),
		}

		go func() {
			if err := s.ListenAndServe(); err != nil {
				asyncErrC <- err
			}
		}()

		select {
		case err := <-asyncErrC:
			log.Printf("warning restarting due to error %v", err)
			// return err
		case <-sigC:
			return nil
		}
	}
}

///////////////////////////////////////////////////////////////////////////////

// https://github.com/benhoyt/go-routing/blob/9a2fa7a643ecb5681f504b95064d948ee2177c9a/retable/route.go

type ctxKey struct{}

var routes = []route{
	newRoute("GET", "/", wrapErr(root)),
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

///////////////////////////////////////////////////////////////////////////////

type compFn func(t *template.Template) (*template.Template, error)

func compRoot(t *template.Template) (*template.Template, error) {
	return template.New("root").Parse(`
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<title>sfw</title>
		</head>
		<body>
			<h1>sup sup sup</h1>
			{{ block "compBar" . }}<div>template fail!</div>{{ end }}
		</body>
		</html>
	`)
}

func compBar(t *template.Template) (*template.Template, error) {
	return t.Parse(`
		{{ define "compBar" }}
		<div>bar</div>
		{{ block "compFoo" . }} <div>failed!</div> {{ end }}
		{{ end }}
	`)
}

func compFoo(s string) compFn {
	return func(t *template.Template) (*template.Template, error) {
		return t.Parse(fmt.Sprintf(`
			{{ define "compFoo" }}
			<div>%s</div>
			{{ end }}`, s),
		)
	}
}

func comps(tfnRoot compFn, tfns ...compFn) (*template.Template, error) {
	t, err := tfnRoot(nil)
	for _, tfn := range tfns {
		if t, err = tfn(t); err != nil {
			return nil, err
		}
	}
	return t, nil
}

func wrapErr(
	hfn func(w http.ResponseWriter, r *http.Request) error,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := hfn(w, r); err != nil {
			log.Printf("error %v", err)
			http.Error(w, "error %v", http.StatusInternalServerError)
		}
	}
}

func root(w http.ResponseWriter, r *http.Request) error {
	// seeds := []lib.GodSeed{}
	// if err := lib.Db.Select(&seeds,
	// 	`SELECT *
	// 	FROM seed`,
	// ); err != nil {
	// 	// log.Printf("0 %v", err)
	// 	// http.Error(w, "database", http.StatusInternalServerError)
	// 	return err
	// }
	// log.Printf("%v", seeds)
	// return nil

	t, err := comps(
		compRoot,
		compBar,
		compFoo("fo sho"),
	)
	if err != nil {
		return err
	}

	log.Printf("info %s", t.DefinedTemplates())

	if err := t.Execute(w, map[string]interface{}{
		// "seeds": seeds,
	}); err != nil {
		return err
	}

	return nil
}
