package core

import (
	"fmt"
	"github.com/jzaikovs/core/loggy"
	. "github.com/jzaikovs/t"
	"net"
	"net/http"
	"net/http/fcgi"
	"time"
)

// default application
var APP = New()

type App struct {
	Config *t_configs
	routes []*t_route
	hooks  map[string][]func()
}

func New() *App {
	return &App{
		routes: make([]*t_route, 0),
		hooks:  make(map[string][]func()),
	}
}

func (this *App) Run() {
	loggy.Start()

	loggy.Info("creating core...")
	//  create configuration if no initialized
	if this.Config == nil {
		this.Config = new(t_configs)
		this.Config.Load("config.json") // default configuration
		this.Config.Load("prod.json")   // production specific configuration
	}

	addr := fmt.Sprintf("%s:%d", this.Config.Host, this.Config.Port)

	loggy.Info("core starting on", addr, "...")

	l, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	this.executeHook("core.done")

	if this.Config.FCGI {
		fcgi.Serve(l, this)
	} else {
		http.Serve(l, this)
	}
}

func (this *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	input := new_input(this, r)
	output := new_output(w)

	if this.route(input, output) {
		return
	}

	if this.Config.HandleContent {
		ServeFile(output, input.RequestURI())
	}
	return
}

func (this *App) Handle(pattern string, handler http.Handler) {
	r := newRoute("?", pattern, func(in Input, out Output) {
		out.no_flush()
		handler.ServeHTTP(out.ResponseWriter(), in.Request())
	})
	r.handler = true
	this.routes = append(this.routes, r)
}

// Main method for dispatching routes
// returns true if found route
func (this *App) route(in Input, out Output) bool {
	//loggy.Log("ROUTE", in.RemoteAddr(), in.Method(), in.RequestURI())

	startTime := time.Now()

	// TODO: this can be more optimized, for example dividing in buckets for each method
	// TODO: try use trie (aka prefix-tree) as routing method
	for _, r := range this.routes {
		if !r.handler && in.Method() != r.method {
			continue // skip routes with different method
		}

		matches := r.pattern.FindStringSubmatch(in.RequestURI())

		if len(matches) == 0 {
			continue // no match, go to next
		}

		// create arguments from groups in route pattern
		// each group is next argument in arguments
		matches = matches[1:]
		args := make([]T, len(matches))
		for i, match := range matches {
			args[i] = T{match}
		}

		// so we found our request
		r.handle(args, startTime, in, out)
		return true
	}

	return false
}

func (this *App) add_route(method, pattern string, callback func(Input, Output)) *t_route {
	r := newRoute(method, pattern, callback)
	this.routes = append(this.routes, r)
	return r
}

func (this *App) Get(pattern string, callback func(Input, Output)) *t_route {
	return this.add_route("GET", pattern, callback)
}

func (this *App) Post(pattern string, callback func(Input, Output)) *t_route {
	return this.add_route("POST", pattern, callback)
}

func (this *App) Put(pattern string, callback func(Input, Output)) *t_route {
	return this.add_route("PUT", pattern, callback)
}

func (this *App) Delete(pattern string, callback func(Input, Output)) *t_route {
	return this.add_route("DELETE", pattern, callback)
}
