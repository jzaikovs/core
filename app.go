package core

import (
	"fmt"
	"net"
	"net/http"
	"net/http/fcgi"

	"github.com/jzaikovs/core/loggy"
)

// Default application handler
var APP = &App{Router: NewRouter()} // default application

// Default global configuration
var DefaultConfig *configStruct

// App structure represents single application on server
// server can route multiple applications,
// in different sub-directory or sub-domain
type App struct {
	Router
	Config    *configStruct // TODO: for each there is application specific configuration
	name      string
	subdomain bool
}

// Module is par of app, for each app there is module instance
type Module struct {
}

// New functions is constructor for app structure
func New(name string, subdomain bool) *App {
	app := new(App)
	app.name = name
	app.subdomain = subdomain
	return app
}

// Run function will initiaate default config load and start listening for requests
func Run() {
	loggy.Start()

	loggy.Info("core.create...")
	//  create configuration if no initialized
	if DefaultConfig == nil {
		DefaultConfig = newConfigStruct()
		DefaultConfig.Load("config.json") // default configuration
		DefaultConfig.Load("prod.json")   // production specific configuration
	}

	addr := fmt.Sprintf("%s:%d", DefaultConfig.Host, DefaultConfig.Port)

	loggy.Info("core.starting.on:", addr, "...")

	l, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	if DefaultConfig.FCGI {
		fcgi.Serve(l, APP)
	} else {
		http.Serve(l, APP)
	}
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
