package core

import (
	. "github.com/jzaikovs/t"
	"time"
)

// this will be router interface for modules,
// for example we have module user which handles users, we can then port
// module across other projects that uses core
type Router interface {
	Get(string, RouteFunc) *Route
	Post(string, RouteFunc) *Route
	Put(string, RouteFunc) *Route
	Delete(string, RouteFunc) *Route
	// main routing function
	Route(context Context) bool
}

type t_default_router struct {
	routes []*Route
}

func new_router() Router {
	return &t_default_router{routes: make([]*Route, 0)}
}

// Main method for dispatching routes
// returns true if found route
func (this *t_default_router) Route(context Context) bool {

	//loggy.Log("ROUTE", context.RemoteAddr(), context.Method(), context.RequestURI())

	startTime := time.Now()

	// TODO: this can be more optimized, for example dividing in buckets for each method
	// TODO: try use trie (aka prefix-tree) as routing method
	for _, r := range this.routes {
		if !r.handler && context.Method() != r.method {
			continue // skip routes with different method
		}

		matches := r.pattern.FindStringSubmatch(context.RequestURI())

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
		r.handle(args, startTime, context)
		return true
	}

	return false
}

func (this *t_default_router) add_route(method, pattern string, callback RouteFunc) *Route {
	r := newRoute(method, pattern, callback, this)
	this.routes = append(this.routes, r)
	return r
}

func (this *t_default_router) Get(pattern string, callback RouteFunc) *Route {
	return this.add_route("GET", pattern, callback)
}

func (this *t_default_router) Post(pattern string, callback RouteFunc) *Route {
	return this.add_route("POST", pattern, callback)
}

func (this *t_default_router) Put(pattern string, callback RouteFunc) *Route {
	return this.add_route("PUT", pattern, callback)
}

func (this *t_default_router) Delete(pattern string, callback RouteFunc) *Route {
	return this.add_route("DELETE", pattern, callback)
}
