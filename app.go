package core

import (
	"fmt"
	. "github.com/jzaikovs/t"
	"net"
	"net/http"
	"net/http/fcgi"
	"time"
)

// default App
var APP = New()

type App struct {
	Config *configs
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
	Log.Info("creating core...")
	//  create config if no initialized
	if this.Config == nil {
		this.Config = new(configs)
		this.Config.Load("config.json") // default configuration
		this.Config.Load("prod.json")   // production specific configuration
	}

	addr := fmt.Sprintf(":%d", this.Config.Port)
	Log.Info("core starting on", addr, "...")

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
	startTime := time.Now()

	// TODO: this can be more optimized, for example dividing in buckets for each method
	for _, r := range this.routes {
		if !r.handler && in.Method() != r.method {
			continue // skip routes with diferent methdo
		}

		matches := r.pattern.FindStringSubmatch(in.RequestURI())

		if len(matches) == 0 {
			continue // no match, go to next
		}

		defer out.Flush()

		if r.req_json {
			if in.ContentType() != ContentType_JSON {
				out.Response(Response_Unsupported_Media_Type)
				return true
			}
		}

		// create arguments from groups in route pattern
		matches = matches[1:]
		args := make([]T, len(matches))
		for i, match := range matches {
			args[i] = T{match}
		}
		in.link_args(args)
		in.link_session(session(in, out))

		// defer some cleanup when done routing
		defer in.Sess().strip()

		// testing rate limits
		if r.test_rate_limit(in, out, startTime) {
			out.Response(Response_Too_Many_Requests)
			return true
		}

		// testing if user is authorized
		if r.test_authorized && !in.Sess().IsAuth() {
			if r.doredirect {
				out.Redirect(r.redirect)
				return true
			}
			out.Response(Response_Unauthorized)
			return true
		}

		in.Data()["is_auth"] = in.Sess().IsAuth()

		if len(r.needs) > 0 {
			data := in.Data()
			for _, need := range r.needs {
				if _, ok := data[need]; !ok {
					out.Response(Response_Unprocessable_Entity)
					return true
				}
			}
		}

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
