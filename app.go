package core

import (
	"fmt"
	"net"
	"net/http"
	"net/http/fcgi"
	"strings"

	"github.com/jzaikovs/core/loggy"
)

// Default application handler
var APP = &App{Router: NewRouter()} // default application

// Default global configuration
var DefaultConfig *configStruct

func init() {
	DefaultConfig = newConfigStruct()
	DefaultConfig.Load("config.json") // default configuration
	DefaultConfig.Load("prod.json")   // production specific configuration
}

// App structure represents single application on server
// server can route multiple applications,
// in different sub-directory or sub-domain
type App struct {
	Router
	Config    *configStruct // TODO: for each there is application specific configuration
	name      string
	subdomain bool
	subs      map[string]*App
}

// Module is par of app, for each app there is module instance
type Module struct {
}

// New functions is constructor for app structure
func New(name string, subdomain bool) *App {
	return &App{
		name:      name,
		subdomain: subdomain,
		subs:      make(map[string]*App),
	}
}

// Run function will initiaate default config load and start listening for requests
func Run() {
	loggy.Info.Println("Starting core...")

	addr := fmt.Sprintf("%s:%d", DefaultConfig.Host, DefaultConfig.Port)

	l, err := net.Listen("tcp", addr)
	if err != nil {
		loggy.Error.Println(err)
		return
	}

	loggy.Info.Println("Listening on:", addr)

	if DefaultConfig.FCGI {
		fcgi.Serve(l, APP)
	} else {
		http.Serve(l, APP)
	}
}

// Sub is used to link application module to main application module
func (app *App) Sub(name string, sub *App) {
	app.subs[strings.ToLower(name)] = sub
}

func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if app.Config == nil {
		app.Config = DefaultConfig
	}

	input := newInput(app, r)
	output := newOutput(w)

	if app.Route(context{input, output}) {
		return
	}

	// subs only executes when base routing failed
	parts := strings.Split(input.RequestURI(), "/")
	if len(parts) > 0 {
		if sub, ok := app.subs[strings.ToLower(parts[0])]; ok {
			// sub-app router will work as if it is main router
			input.reqURI = "/" + strings.Join(parts[1:], "/")
			if sub.Route(context{input, output}) {
				return
			}
		}
	}

	if DefaultConfig.HandleContent {
		ServeFile(output, input.RequestURI())
	}
	return
}

// TODO: this functions should be reworked
/*
func (app *App) Handle(pattern string, handler http.Handler) {
	r := newRoute("?", pattern, func(context Context) {
		context.noFlush()
		handler.ServeHTTP(context.ResponseWriter(), context.Request())
	}, app)
	r.handler = true
	app. = append(app.routes, r)
}
*/
