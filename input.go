package core

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/jzaikovs/core/loggy"
	"github.com/jzaikovs/core/session"
	"github.com/jzaikovs/t"
)

// Input is interface for routes input handler
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
	Args(int) t.T
	// Access to posted data
	Data() t.Map
	// Provides access to session data
	Session() *session.Session

	// returns body content, JSON post with JSON as content-body
	Body() (result string)

	Request() *http.Request

	linkArgs([]t.T)
	linkSession(*session.Session)
	addData(string, interface{})
}

type defaultInput struct {
	app     *App
	request *http.Request
	args    []t.T
	session *session.Session
	data    t.Map
	parsed  bool
	body    []byte
	reqURI  string
}

func newInput(app *App, request *http.Request) (in *defaultInput) {
	in = &defaultInput{
		request: request,
		data:    make(t.Map),
		args:    make([]t.T, 0),
		app:     app,
	}

	in.parseSegments()

	in.body, _ = ioutil.ReadAll(in.request.Body)
	in.request.Body.Close()

	buf := new(bytes.Buffer)
	buf.Write(in.body)
	in.request.Body = ioutil.NopCloser(buf)

	in.data["base_url"] = app.Config.BaseURL
	return in
}

func (in *defaultInput) parseSegments() {
	u, err := url.ParseQuery(in.RequestURI())
	if err != nil {
		loggy.Error.Println(err)
		return
	}

	for k, v := range u {
		if len(v) == 1 {
			in.data[k] = v[0]
		} else {
			in.data[k] = v
		}
	}
}

func (in *defaultInput) App() *App {
	return in.app
}

func (in *defaultInput) linkArgs(args []t.T) {
	in.args = args
}

func (in *defaultInput) linkSession(session *session.Session) {
	in.session = session
}

func (in *defaultInput) Args(idx int) t.T {
	return in.args[idx]
}

func (in *defaultInput) Session() *session.Session {
	return in.session
}
func (in *defaultInput) ContentType() string {
	return in.HeaderValue("Content-Type")
}

func (in *defaultInput) Data() t.Map {
	// repeat request will return result of previous calls
	if in.parsed {
		return in.data
	}

	if strings.Contains(in.ContentType(), ContentType_JSON) {
		temp := make(t.Map)

		// parse body for JSON data into temporal map
		if err := json.Unmarshal(in.body, &temp); err != nil {
			loggy.Error.Println("core.input.data err", in.RemoteAddr(), err.Error(), string(in.body))
		}

		// clean all incoming values, to protect as from some bad injections
		// TODO: do we really need?
		for k, v := range temp {
			switch val := v.(type) {
			case string:
				in.data[k] = Clean(val)
			default:
				in.data[k] = v
			}
		}

		// mark that data is parsed and can be returned on next call without parsing
		in.parsed = true

		return in.data
	}

	if in.request.Form == nil {
		in.request.ParseMultipartForm(32 << 20) // 32MB
	}

	// read all form values if we have post-data
	for k, v := range in.request.Form {
		switch len(v) {
		case 1:
			in.data[k] = v[0] // remove from slice if single value
		case 0:
		default:
			in.data[k] = v
		}
	}

	in.parsed = true

	return in.data
}

func (in *defaultInput) RequestURI() (uri string) {
	if in.reqURI != "" {
		return in.reqURI
	}

	uri = in.request.URL.Path

	if in.app.Config.FCGI && len(in.request.URL.Opaque) > 0 {
		// using Nginx hack
		// fastcgi_param REQUEST_URI "$scheme: $request_uri";
		// fastcgi_param HTTP_HOST "";
		uri = in.request.URL.Opaque[1:]
	} else {
		if len(in.request.RequestURI) > 0 {
			uri = in.request.RequestURI
		} else if in.app.Config.FCGI {
			// using Nginx hack
			// fastcgi_param HTTP_REQUEST_URI $request_uri;
			uri = strings.TrimRight(in.request.Header.Get("Request-Uri"), "?")[len(in.app.Config.Subdir):]
		}
	}

	in.reqURI = uri
	return
}

func (in *defaultInput) HeaderValue(key string) string {
	return in.request.Header.Get(key)
}

func (in *defaultInput) Get(key string) (val string) {
	val = in.data.Str(key)
	return
}

func (in *defaultInput) Method() string {
	return in.request.Method
}

func (in *defaultInput) FormValue(name string) string {
	return in.request.FormValue(name)
}

func (in *defaultInput) CookieValue(name string) (string, bool) {
	cookie, err := in.request.Cookie(name)
	if err != nil {
		return "", false
	}
	return cookie.Value, true
}

func (in *defaultInput) UserAgent() string {
	return in.request.UserAgent()
}

func (in *defaultInput) RemoteAddr() string {
	return strings.Split(in.request.RemoteAddr, ":")[0]
}

func (in *defaultInput) Body() string {
	return string(in.body)
}

func (in *defaultInput) Ajax() bool {
	if in.ContentType() == "application/json" {
		return true
	}
	return strings.ToLower(in.request.Header.Get("X-Requested-With")) == "xmlhttprequest"
}

func (in *defaultInput) Request() *http.Request {
	return in.request
}

func (in *defaultInput) addData(k string, v interface{}) {
	in.data[k] = v
}
