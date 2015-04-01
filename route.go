package core

import (
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/jzaikovs/core/loggy"
	"github.com/jzaikovs/core/session"
	. "github.com/jzaikovs/t"
	"regexp"
	"time"
)

type RouteFunc func(Context)

type Route struct {
	app Router

	patternStr string
	pattern    *regexp.Regexp
	callback   RouteFunc
	method     string

	handler  bool
	req_json bool
	no_cache bool

	// authorized user test config
	test_authorized bool   // to call route function, session must be authorized
	redirect        string // if doredirect set then redirects to redirect value
	doredirect      bool
	// rate-limits for guest and authorized user
	limit      *ratelimit
	limitAuth  *ratelimit
	limits     map[string]*ratelimit // this is rate limit for each IP address
	limitsAuth map[string]*ratelimit

	rules []func(context Context) error

	need_valid_csrf_token bool
	emit_csrf_token       bool

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

func (this *Route) ReqAuth(args ...string) *Route {
	this.test_authorized = true
	if len(args) > 0 {
		this.redirect = args[0]
		this.doredirect = true
	}
	return this
}

func (this *Route) Need(fields ...string) *Route {
	this.rules = append(this.rules, func(context Context) error {
		data := context.Data()
		// for request we can add some mandatory fields
		// for example, we can add that for sign-in we need login and password
		for _, need := range fields {
			if _, ok := data[need]; !ok {
				// TODO: if string, then check for empty string?
				return errors.New(fmt.Sprintf("field [%s] required", need))
			}
		}

		return nil
	})

	return this
}

func (this *Route) RateLimitAuth(rate, per float32) *Route {
	this.limitAuth = new_ratelimit(rate, per)
	this.limitsAuth = make(map[string]*ratelimit)
	return this
}

func (this *Route) RateLimit(rate, per float32) *Route {
	this.limit = new_ratelimit(rate, per)
	this.limits = make(map[string]*ratelimit)
	return this
}

func (this *Route) Match(nameA, nameB string) *Route {
	this.rules = append(this.rules, func(context Context) error {
		data := context.Data()

		if data.Str(nameA) != data.Str(nameB) {
			return errors.New(fmt.Sprintf(`field [%s] not match field [%s]`, nameA, nameB))
		}

		return nil
	})

	return this
}

// request content type must be json
func (this *Route) JSON() *Route {
	this.req_json = true
	return this
}

// output of rout will not be cached in any way
// to client will be sent headers to not cache response
func (this *Route) NoCache() *Route {
	this.no_cache = true
	return this
}

// route option for setting CSRF validations
func (this *Route) CSRF(emit, need bool) *Route {
	this.emit_csrf_token = emit
	this.need_valid_csrf_token = need
	return this
}

func (this *Route) test_rate_limit(context Context, t time.Time) bool {
	var (
		limit *ratelimit
		ok    bool
	)

	if context.Session().IsAuth() {
		if this.limitAuth == nil {
			return false
		}

		if limit, ok = this.limitsAuth[context.RemoteAddr()]; !ok { // TODO: improve and remove race
			limit = new_ratelimit(this.limitAuth.rate, this.limitAuth.per)
			limit.lass_check = t
			this.limitsAuth[context.RemoteAddr()] = limit
		}

	} else {
		if this.limit == nil {
			return false
		}
		if limit, ok = this.limits[context.RemoteAddr()]; !ok { // TODO: improve and remove race
			limit = new_ratelimit(this.limit.rate, this.limit.per)
			limit.lass_check = t
			this.limits[context.RemoteAddr()] = limit
		}
	}
	ok = limit.test(t)
	context.AddHeader(Header_X_Rate_Limit_Limit, int(limit.rate))
	context.AddHeader(Header_X_Rate_Limit_Remaining, int(limit.allowance))
	return !ok
}

// route handler method
func (this *Route) handle(args []T, startTime time.Time, context Context) {
	// now defer that at the end we write data

	defer context.Flush()

	// route asks for JSON as content type
	if this.req_json {
		if context.ContentType() != ContentType_JSON {
			context.Response(Response_Unsupported_Media_Type)
			return
		}
	}

	// connect our request to session manager
	context.link_args(args)
	context.link_session(session.New(context))

	// defer some cleanup when done routing
	defer context.Session().Unlink()

	// testing rate limits
	// TODO: need testing
	if this.test_rate_limit(context, startTime) {
		context.Response(Response_Too_Many_Requests)
		return
	}

	// testing if user is authorized
	// route have flag that session must be authorize to access it
	if this.test_authorized && !context.Session().IsAuth() {
		// if we have set up redirect then on fail we redirect there
		if this.doredirect {
			context.Redirect(this.redirect)
			return
		}
		// else just say that we are unauthorized
		context.Response(Response_Unauthorized)
		return
	}

	if this.need_valid_csrf_token {
		csrf, ok := context.CookieValue("_csrf")
		if !ok || len(csrf) == 0 || csrf != context.Session().Data.Str("_csrf") {
			context.Response(Response_Forbidden) // TODO: what is best status code for CSRF violation
			return
		}
		delete(context.Session().Data, "_csrf")
		context.SetCookieValue("_csrf", "")
	}

	// this can be useful if we add session status in request data
	// TODO: need some mark to identify core added data, example, $is_auth, $base_url, etc..
	context.addData("is_auth", context.Session().IsAuth())

	if this.no_cache {
		// this is for IE to not cache JSON responses!
		context.AddHeader("If-Modified-Since", "01 Jan 1970 00:00:00 GMT")
		context.AddHeader("Cache-Control", "no-cache")
	}

	// TODO: verify that this is good way to emit CSRF tokens
	if this.emit_csrf_token {
		// generate csrf token
		b := make([]byte, 16)
		rand.Read(b)
		csrf := Base64Encode(b)
		context.Session().Data["_csrf"] = csrf
		context.SetCookieValue("_csrf", csrf)
	}

	// validate all added rules
	for _, rule := range this.rules {
		if err := rule(context); err != nil {
			loggy.Log("BAD", context.RemoteAddr(), err)
			context.WriteJSON(DefaultConfig.err_object_func(Response_Bad_Request, err))
			context.Response(Response_Bad_Request)
			return
		}
	}

	// call route function
	this.callback(context)
}
