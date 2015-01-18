package core

import (
	"bytes"
	"encoding/json"
	. "github.com/jzaikovs/t"
	"io/ioutil"
	"net/http"
	"strings"
)

type Input interface {
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
	Session() *Session

	// returns body content, JSON post with JSON as content-body
	Body() (result string)

	Request() *http.Request

	link_args([]T)
	link_session(*Session)
	addData(string, interface{})
}

type input struct {
	app     *App
	request *http.Request
	args    []T
	session *Session
	data    Map
	parsed  bool
	body    []byte
}

func new_input(app *App, request *http.Request) (this *input) {
	this = &input{
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

	this.body, _ = ioutil.ReadAll(this.request.Body)
	this.request.Body.Close()

	buf := new(bytes.Buffer)
	buf.Write(this.body)
	this.request.Body = ioutil.NopCloser(buf)

	this.data["base_url"] = app.Config.BaseUrl
	return this
}

func (this *input) link_args(args []T) {
	this.args = args
}

func (this *input) link_session(session *Session) {
	this.session = session
}

func (this *input) Args(idx int) T {
	return this.args[idx]
}

func (this *input) Session() *Session {
	return this.session
}
func (this *input) ContentType() string {
	return this.HeaderValue("Content-Type")
}

func (this *input) Data() Map {
	// repeat request will return result of previous calls
	if this.parsed {
		return this.data
	}

	if this.ContentType() == ContentType_JSON {
		temp := make(Map)
		json.Unmarshal(this.body, &temp)
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

func (this *input) RequestURI() string {
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

func (this *input) HeaderValue(key string) string {
	return this.request.Header.Get(key)
}

func (this *input) Get(key string) (val string) {
	val = this.data.Str(key)
	return
}

func (this *input) Method() string {
	return this.request.Method
}

func (this *input) FormValue(name string) string {
	return this.request.FormValue(name)
}

func (this *input) CookieValue(name string) (string, bool) {
	cookie, err := this.request.Cookie(name)
	if err != nil {
		return "", false
	}
	return cookie.Value, true
}

func (this *input) UserAgent() string {
	return this.request.UserAgent()
}

func (this *input) RemoteAddr() string {
	return strings.Split(this.request.RemoteAddr, ":")[0]
}

func (this *input) Body() string {
	return string(this.body)
}

func (this *input) Ajax() bool {
	if this.ContentType() == "application/json" {
		return true
	}
	return strings.ToLower(this.request.Header.Get("X-Requested-With")) == "xmlhttprequest"
}

func (this *input) Request() *http.Request {
	return this.request
}

func (this *input) addData(k string, v interface{}) {
	this.data[k] = v
}
