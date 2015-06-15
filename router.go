package core

import (
	"net/http"
	"time"

	"github.com/jzaikovs/t"
)

// Router is interface for implement specific routing engines,
// for example we have module user which handles users, we can then port
// module across other projects that uses core
type Router interface {
	Get(string, RouteFunc) *Route
	Post(string, RouteFunc) *Route
	Put(string, RouteFunc) *Route
	Delete(string, RouteFunc) *Route
	// main routing function
	Route(context Context) bool
	Handle(pattern string, handler http.Handler)
}

type defaultRouter struct {
	routes []*Route
}

// NewRouter is constructor for creating router instance for default core router
func NewRouter() Router {
	return &defaultRouter{routes: make([]*Route, 0)}
}

// Route if main method for dispatching routes
// returns true if found route
func (router *defaultRouter) Route(context Context) bool {

	//loggy.Log("ROUTE", context.RemoteAddr(), context.Method(), context.RequestURI())

	startTime := time.Now()

	// TODO: router can be more optimized, for example dividing in buckets for each method
	// TODO: try use trie (aka prefix-tree) as routing method
	for _, r := range router.routes {
		if !r.handler && context.Method() != r.method {
			continue // skip routes with different method
		}

		matches := r.pattern.FindStringSubmatch(context.RequestURI())

		//loggy.Trace.Println(matches)

		if len(matches) == 0 {
			continue // no match, go to next
		}

		if r.handler {
			r.callback(context)
			return true
		}

		// create arguments from groups in route pattern
		// each group is next argument in arguments
		matches = matches[1:]
		args := make([]t.T, len(matches))
		for i, match := range matches {
			args[i] = t.T{Value: match}
		}

		// so we found our request
		r.handle(args, startTime, context)
		return true
	}

	return false
}

func (router *defaultRouter) addRoute(method, pattern string, callback RouteFunc) *Route {
	r := newRoute(method, pattern, callback, router)
	router.routes = append(router.routes, r)
	return r
}

// Get adds router handler for GET request
func (router *defaultRouter) Get(pattern string, callback RouteFunc) *Route {
	return router.addRoute("GET", pattern, callback)
}

// Post adds router for POST request
func (router *defaultRouter) Post(pattern string, callback RouteFunc) *Route {
	return router.addRoute("POST", pattern, callback)
}

// Put adds router for PUT request
func (router *defaultRouter) Put(pattern string, callback RouteFunc) *Route {
	return router.addRoute("PUT", pattern, callback)
}

// Delete adds router for DELETE request
func (router *defaultRouter) Delete(pattern string, callback RouteFunc) *Route {
	return router.addRoute("DELETE", pattern, callback)
}

// Handle implemted to support 3rd party packages that uses http.Handler
func (router *defaultRouter) Handle(pattern string, handler http.Handler) {
	r := router.addRoute("?", pattern, func(context Context) {
		context.noFlush()
		handler.ServeHTTP(context.ResponseWriter(), context.Request())
	})
	// mark router as handler
	r.handler = true
}
