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
	routes []*route
	hooks  map[string][]func()
}

func New() *App {
	return &App{
		routes: make([]*route, 0),
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
	fmt.Println("core starting on", addr, "...")

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

		// so we found our request
		// now defer that at the end we write data

		defer out.Flush()

		// route asks for JSON as content type
		if r.req_json {
			if in.ContentType() != ContentType_JSON {
				out.Response(Response_Unsupported_Media_Type)
				return true
			}
		}

		// create arguments from groups in route pattern
		// each group is next argument in arguments
		matches = matches[1:]
		args := make([]T, len(matches))
		for i, match := range matches {
			args[i] = T{match}
		}

		// connect our request to session manager
		in.link_args(args)
		in.link_session(session(in, out))

		// defer some cleanup when done routing
		defer in.Sess().strip()

		// testing rate limits
		// TODO: need testing
		if r.test_rate_limit(in, out, startTime) {
			out.Response(Response_Too_Many_Requests)
			return true
		}

		// testing if user is authorized
		// route have flag that session must be authorize to access it
		if r.test_authorized && !in.Sess().IsAuth() {
			// if we have set up redirect then on fail we redirect there
			if r.doredirect {
				out.Redirect(r.redirect)
				return true
			}
			// else just say that we are unauthorized
			out.Response(Response_Unauthorized)
			return true
		}

		// for request we can add some mandatory fields
		// for example, we can add that for sign-in we need login and password
		if len(r.needs) > 0 {
			data := in.Data()
			for _, need := range r.needs {
				if _, ok := data[need]; !ok {
					out.Response(Response_Unprocessable_Entity)
					return true
				}
			}
		}

		// this can be useful if we add session status in request data
		// TODO: need some mark to identify core added data, example, $is_auth, $base_url, etc..
		in.addData("is_auth", in.Sess().IsAuth())

		if r.no_cache {
			// this is for IE to not cache JSON responses!
			out.AddHeader("If-Modified-Since", "01 Jan 1970 00:00:00 GMT")
			out.AddHeader("Cache-Control", "no-cache")
		}

		// call route function
		r.callback(in, out)
		return true
	}
	return false
}

func (this *App) addRoute(method, pattern string, callback func(Input, Output)) *route {
	r := newRoute(method, pattern, callback)
	this.routes = append(this.routes, r)
	return r
}

func (this *App) Get(pattern string, callback func(Input, Output)) *route {
	return this.addRoute("GET", pattern, callback)
}

func (this *App) Post(pattern string, callback func(Input, Output)) *route {
	return this.addRoute("POST", pattern, callback)
}

func (this *App) Put(pattern string, callback func(Input, Output)) *route {
	return this.addRoute("PUT", pattern, callback)
}

func (this *App) Delete(pattern string, callback func(Input, Output)) *route {
	return this.addRoute("DELETE", pattern, callback)
}
