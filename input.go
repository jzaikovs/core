package core

import (
	"bytes"
	"encoding/json"
	"github.com/jzaikovs/core/loggy"
	"github.com/jzaikovs/core/session"
	. "github.com/jzaikovs/t"
	"io/ioutil"
	"net/http"
	"strings"
)

type Input interface {
	App() *App

	RequestURI() string
	// Get used for getting GET passed parameters
	Get(key string) (val string)
	// Returns method of request (GET, POST, PUT, ...)
	Method() string
	// Returns header value
	HeaderValue(string) string

	ContentType() string
	// wrapper for http.Request.FormValue
	FormValue(string) string

	CookieValue(string) (string, bool)
	// Return user agent
	UserAgent() string
	// Returns remote IP address
	RemoteAddr() string

	// Returns true if request is marked as AJAX based
	Ajax() bool
	// Access to URL parameters
	Args(int) T
	// Access to posted data
	Data() Map
	// Provides access to session data
	Session() *session.Session

	// returns body content, JSON post with JSON as content-body
	Body() (result string)

	Request() *http.Request

	link_args([]T)
	link_session(*session.Session)
	addData(string, interface{})

	Segments() []T
}

type t_input struct {
	app      *App
	request  *http.Request
	args     []T
	session  *session.Session
	data     Map
	parsed   bool
	body     []byte
	segments []T
}

func new_input(app *App, request *http.Request) (this *t_input) {
	this = &t_input{
		request: request,
		data:    make(Map),
		args:    make([]T, 0),
		app:     app,
	}

	parts := strings.Split(this.RequestURI(), "?")
	if len(parts) > 1 {
		for _, part := range strings.Split(parts[1], "&") {
			idx := strings.Index(part, "=")
			if idx <= 0 {
				continue
			}
			key := part[:idx]
			if len(part) > idx+1 {
				val := part[idx+1:]
				this.data[key] = val
			} else {
				this.data[key] = ""
			}
		}
	}

	this.segments = make([]T, 0)

	if len(parts) > 0 {
		for _, segment := range strings.Split(strings.Trim(parts[0], "/"), "/") {
			this.segments = append(this.segments, T{segment})
		}
	}

	this.body, _ = ioutil.ReadAll(this.request.Body)
	this.request.Body.Close()

	buf := new(bytes.Buffer)
	buf.Write(this.body)
	this.request.Body = ioutil.NopCloser(buf)

	this.data["base_url"] = app.Config.BaseUrl
	return this
}

func (this *t_input) App() *App {
	return this.app
}

func (this *t_input) link_args(args []T) {
	this.args = args
}

func (this *t_input) link_session(session *session.Session) {
	this.session = session
}

func (this *t_input) Args(idx int) T {
	return this.args[idx]
}

func (this *t_input) Session() *session.Session {
	return this.session
}
func (this *t_input) ContentType() string {
	return this.HeaderValue("Content-Type")
}

func (this *t_input) Data() Map {
	// repeat request will return result of previous calls
	if this.parsed {
		return this.data
	}

	if strings.Contains(this.ContentType(), ContentType_JSON) {
		temp := make(Map)

		// parse body for JSON data into temporal map
		if err := json.Unmarshal(this.body, &temp); err != nil {
			loggy.Error("core.input.data err", this.RemoteAddr(), err.Error(), string(this.body))
		}

		// clean all incoming values, to protect as from some bad injections
		// TODO: do we really need?
		for k, v := range temp {
			switch val := v.(type) {
			case string:
				this.data[k] = Clean(val)
			default:
				this.data[k] = v
			}
		}

		// mark that data is parsed and can be returned on next call without parsing
		this.parsed = true

		return this.data
	}

	if this.request.Form == nil {
		this.request.ParseMultipartForm(32 << 20) // 32MB
	}

	// read all form values if we have post-data
	for k, v := range this.request.Form {
		switch len(v) {
		case 1:
			this.data[k] = v[0] // remove from slice if single value
		case 0:
		default:
			this.data[k] = v
		}
	}

	this.parsed = true

	return this.data
}

func (this *t_input) RequestURI() string {
	if this.app.Config.FCGI && len(this.request.URL.Opaque) > 0 {
		// using Nginx hack
		// fastcgi_param REQUEST_URI "$scheme: $request_uri";
		// fastcgi_param HTTP_HOST "";
		return this.request.URL.Opaque[1:]
	}
	if len(this.request.RequestURI) > 0 {
		return this.request.RequestURI
	} else if this.app.Config.FCGI {
		// using Nginx hack
		// fastcgi_param HTTP_REQUEST_URI $request_uri;
		return strings.TrimRight(this.request.Header.Get("Request-Uri"), "?")[len(this.app.Config.Subdir):]
	}
	return this.request.URL.Path
}

func (this *t_input) Segments() []T {
	return this.segments
}

func (this *t_input) HeaderValue(key string) string {
	return this.request.Header.Get(key)
}

func (this *t_input) Get(key string) (val string) {
	val = this.data.Str(key)
	return
}

func (this *t_input) Method() string {
	return this.request.Method
}

func (this *t_input) FormValue(name string) string {
	return this.request.FormValue(name)
}

func (this *t_input) CookieValue(name string) (string, bool) {
	cookie, err := this.request.Cookie(name)
	if err != nil {
		return "", false
	}
	return cookie.Value, true
}

func (this *t_input) UserAgent() string {
	return this.request.UserAgent()
}

func (this *t_input) RemoteAddr() string {
	return strings.Split(this.request.RemoteAddr, ":")[0]
}

func (this *t_input) Body() string {
	return string(this.body)
}

func (this *t_input) Ajax() bool {
	if this.ContentType() == "application/json" {
		return true
	}
	return strings.ToLower(this.request.Header.Get("X-Requested-With")) == "xmlhttprequest"
}

func (this *t_input) Request() *http.Request {
	return this.request
}

func (this *t_input) addData(k string, v interface{}) {
	this.data[k] = v
}
