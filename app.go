package core

import (
	"fmt"
	"github.com/jzaikovs/core/loggy"
	"net"
	"net/http"
	"net/http/fcgi"
)

var (
	APP                      = &App{Router: NewRouter()} // default application
	DefaultConfig *t_configs = nil
)

// structure represents single application on server
// server can route multiple applications,
// in different sub-directory or sub-domain
type App struct {
	Router
	Config    *t_configs // TODO: for each there is application specific configuration
	name      string
	subdomain bool
}

// module is par of app, for each app there is module instance
type Module struct {
}

func New(name string, subdomain bool) *App {
	app := new(App)
	app.name = name
	app.subdomain = subdomain
	return app
}

type ServerOption struct {
}

func Run() {
	loggy.Start()

	loggy.Info("core.create...")
	//  create configuration if no initialized
	if DefaultConfig == nil {
		DefaultConfig = new_t_config()
		DefaultConfig.Load("config.json") // default configuration
		DefaultConfig.Load("prod.json")   // production specific configuration
	}

	addr := fmt.Sprintf("%s:%d", DefaultConfig.Host, DefaultConfig.Port)

	loggy.Info("core.starting.on:", addr, "...")

	l, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	//TODO: make it simple, do we really need hooks?
	//this.executeHook("core.done")

	if DefaultConfig.FCGI {
		fcgi.Serve(l, APP)
	} else {
		http.Serve(l, APP)
	}
}

func (this *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if this.Config == nil {
		this.Config = DefaultConfig
	}

	input := new_input(this, r)
	output := new_output(w)

	if this.Route(t_context{input, output}) {
		return
	}

	if DefaultConfig.HandleContent {
		ServeFile(output, input.RequestURI())
	}
	return
}

/*
func (this *App) Handle(pattern string, handler http.Handler) {
	r := newRoute("?", pattern, func(context Context) {
		context.no_flush()
		handler.ServeHTTP(context.ResponseWriter(), context.Request())
	}, this)
	r.handler = true
	this.routes = append(this.routes, r)
}
*/
func (this *App) Module(pattern string, router Router) {

}
