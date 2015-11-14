package core

import (
	"crypto/rand"
	"fmt"
	"regexp"
	"time"

	"github.com/jzaikovs/core/loggy"
	"github.com/jzaikovs/core/session"
	"github.com/jzaikovs/t"
	"github.com/jzaikovs/tokenbucket"
)

const (
	HeaderXRateLimit          = `X-Rate-Limit-Limit`
	HeaderXRateLimitRemaining = `X-Rate-Limit-Remaining`
)

// RouteFunc is function type used in routes
type RouteFunc func(Context)

// Route handles single route
type Route struct {
	app Router

	patternStr string
	pattern    *regexp.Regexp
	callback   RouteFunc
	method     string

	handler     bool
	jsonRequest bool
	noCache     bool

	// authorized user test config
	authRequest bool   // to call route function, session must be authorized
	redirect    string // if doredirect set then redirects to redirect value
	doredirect  bool

	// rate-limits for guest and authorized user
	limits     *tokenbucket.Buckets // this is rate limit for each IP address
	limitsAuth *tokenbucket.Buckets

	rules []func(context Context) error

	validateCSRFToken bool
	emitCSRFToken     bool

	needs []string
}

func newRoute(method, pattern string, callback RouteFunc, router Router) *Route {
	return &Route{
		app:        router,
		patternStr: pattern,
		pattern:    regexp.MustCompile(pattern),
		callback:   callback,
		method:     method,
	}
}

// ReqAuth marks route so that it can be accessed only by authorized session
// if session is not authorized request is redirected to route that is passed in argument
func (route *Route) ReqAuth(args ...string) *Route {
	route.authRequest = true
	if len(args) > 0 {
		route.redirect = args[0]
		route.doredirect = true
	}
	return route
}

// Need functions adds validation for mandatory fields
func (route *Route) Need(fields ...string) *Route {
	route.rules = append(route.rules, func(context Context) error {
		data := context.Data()
		// for request we can add some mandatory fields
		// for example, we can add that for sign-in we need login and password
		for _, need := range fields {
			if _, ok := data[need]; !ok {
				// TODO: if string, then check for empty string?
				return fmt.Errorf("field [%s] required", need)
			}
		}

		return nil
	})

	return route
}

// RateLimitAuth sets routes maximum request rate per time for authorized users
func (route *Route) RateLimitAuth(rate, per float32) *Route {
	route.limitsAuth = tokenbucket.NewBuckets(int(rate), rate/per)
	return route
}

// RateLimit sets routes maximum request rate per second from specific remote IP
func (route *Route) RateLimit(rate, per float32) *Route {
	route.limits = tokenbucket.NewBuckets(int(rate), rate/per)
	return route
}

// Match sets a rule that two input data fields should match
func (route *Route) Match(nameA, nameB string) *Route {
	route.rules = append(route.rules, func(context Context) error {
		data := context.Data()

		if data.Str(nameA) != data.Str(nameB) {
			return fmt.Errorf(`field [%s] not match field [%s]`, nameA, nameB)
		}

		return nil
	})

	return route
}

// JSON adds validation for request content so that only requests with content type json is handled
func (route *Route) JSON() *Route {
	route.jsonRequest = true
	return route
}

// NoCache marks request handler output of route will not be cached in any way
// to client will be sent headers to not cache response
func (route *Route) NoCache() *Route {
	route.noCache = true
	return route
}

// CSRF route option for setting CSRF validations
func (route *Route) CSRF(emit, need bool) *Route {
	route.emitCSRFToken = emit
	route.validateCSRFToken = need
	return route
}

func (route *Route) exeedsRateLimit(context Context, t time.Time) bool {

	// if session is authorized then check auth rate limits
	if context.Session().IsAuth() {
		if route.limitsAuth != nil {
			space, ok := route.limitsAuth.Add(context.RemoteAddr(), t)
			if !ok {
				return true
			}

			context.AddHeader(HeaderXRateLimit, route.limitsAuth.Capacity())
			context.AddHeader(HeaderXRateLimitRemaining, space)
		}

		return false
	}

	if route.limits != nil {
		space, ok := route.limits.Add(context.RemoteAddr(), t)
		if !ok { // reached guest rate limie for IP, to many request from this IP
			return true
		}

		context.AddHeader(HeaderXRateLimit, route.limits.Capacity())
		context.AddHeader(HeaderXRateLimitRemaining, space)
	}

	return false
}

// route handler method
func (route *Route) handle(args []t.T, startTime time.Time, context Context) {
	// now defer that at the end we write data

	defer context.Flush()

	// route asks for JSON as content type
	if route.jsonRequest {
		if context.ContentType() != ContentType_JSON {
			context.Response(Response_Unsupported_Media_Type)
			return
		}
	}

	// connect our request to session manager
	context.linkArgs(args)
	context.linkSession(session.New(context))

	// defer some cleanup when done routing
	defer context.Session().Unlink()

	// testing rate limits
	// TODO: need testing
	if route.exeedsRateLimit(context, startTime) {
		context.Response(Response_Too_Many_Requests)
		return
	}

	// testing if user is authorized
	// route have flag that session must be authorize to access it
	if route.authRequest && !context.Session().IsAuth() {
		// if we have set up redirect then on fail we redirect there
		if route.doredirect {
			context.Redirect(route.redirect)
			return
		}
		// else just say that we are unauthorized
		context.Response(Response_Unauthorized)
		return
	}

	if route.validateCSRFToken {
		csrf, ok := context.CookieValue("_csrf")
		if !ok || len(csrf) == 0 || csrf != context.Session().Data.Str("_csrf") {
			context.Response(Response_Forbidden) // TODO: what is best status code for CSRF violation
			return
		}
		delete(context.Session().Data, "_csrf")
		context.SetCookieValue("_csrf", "")
	}

	// route can be useful if we add session status in request data
	// TODO: need some mark to identify core added data, example, $is_auth, $base_url, etc..
	context.addData("is_auth", context.Session().IsAuth())

	if route.noCache {
		// route is for IE to not cache JSON responses!
		context.AddHeader("If-Modified-Since", "01 Jan 1970 00:00:00 GMT")
		context.AddHeader("Cache-Control", "no-cache")
	}

	// TODO: verify that route is good way to emit CSRF tokens
	if route.emitCSRFToken {
		// generate csrf token
		b := make([]byte, 16)
		rand.Read(b)
		csrf := Base64Encode(b)
		context.Session().Data["_csrf"] = csrf
		context.SetCookieValue("_csrf", csrf)
	}

	// validate all added rules
	for _, rule := range route.rules {
		if err := rule(context); err != nil {
			loggy.Warning.Println(context.RemoteAddr(), err)
			context.WriteJSON(DefaultConfig.err_object_func(Response_Bad_Request, err))
			context.Response(Response_Bad_Request)
			return
		}
	}

	// call route function
	route.callback(context)
}
