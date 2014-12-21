package core

import (
	"regexp"
	"time"
)

type route struct {
	patternStr string
	pattern    *regexp.Regexp
	callback   func(Input, Output)
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

	needs []string
}

func newRoute(method, pattern string, callback func(Input, Output)) *route {
	r := new(route)
	r.patternStr = pattern
	r.pattern = regexp.MustCompile(pattern)
	r.callback = callback
	r.method = method
	return r
}

func (this *route) ReqAuth(args ...string) *route {
	this.test_authorized = true
	if len(args) > 0 {
		this.redirect = args[0]
		this.doredirect = true
	}
	return this
}

func (this *route) Need(fields ...string) {
	this.needs = append(this.needs, fields...)
}

func (this *route) RateLimitAuth(rate, per float32) *route {
	this.limitAuth = new_ratelimit(rate, per)
	this.limitsAuth = make(map[string]*ratelimit)
	return this
}

func (this *route) RateLimit(rate, per float32) *route {
	this.limit = new_ratelimit(rate, per)
	this.limits = make(map[string]*ratelimit)
	return this
}

func (this *route) JSON() *route {
	this.req_json = true
	return this
}

func (this *route) NoCache() {
	this.no_cache = true
}

func (this *route) test_rate_limit(in Input, out Output, t time.Time) bool {
	var (
		limit *ratelimit
		ok    bool
	)

	if in.Sess().IsAuth() {
		if this.limitAuth == nil {
			return false
		}
		if limit, ok = this.limitsAuth[in.RemoteAddr()]; !ok { // TODO: improve and remove race
			limit = new_ratelimit(this.limitAuth.rate, this.limitAuth.per)
			limit.lass_check = t
			this.limitsAuth[in.RemoteAddr()] = limit
		}

	} else {
		if this.limit == nil {
			return false
		}
		if limit, ok = this.limits[in.RemoteAddr()]; !ok { // TODO: improve and remove race
			limit = new_ratelimit(this.limit.rate, this.limit.per)
			limit.lass_check = t
			this.limits[in.RemoteAddr()] = limit
		}
	}
	ok = limit.test(t)
	out.AddHeader(Header_X_Rate_Limit_Limit, limit.rate)
	out.AddHeader(Header_X_Rate_Limit_Remaining, limit.allowance)
	return !ok
}
