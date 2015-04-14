package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Output is interface for route output handler
type Output interface {
	// writing function
	Write([]byte) (n int, err error)
	WriteString(string) (n int, err error)
	WriteJSON(interface{}) []byte

	// response config
	SetContentType(string)
	Response(int)

	// some heper functions
	SetCookieValue(string, string)
	Redirect(url ...string)
	AddHeader(string, interface{})
	Header() http.Header

	ResponseWriter() http.ResponseWriter
	Flush()

	noFlush()
}

type output struct {
	response     http.ResponseWriter
	buffer       bytes.Buffer
	responseCode int
	noflush      bool
}

func newOutput(response http.ResponseWriter) *output {
	out := &output{response: response, responseCode: 200}
	out.SetContentType(MIME_HTML)
	out.AddHeader("Access-Control-Allow-Origin", "*")
	return out
}

func (out *output) Write(p []byte) (n int, err error) {
	return out.buffer.Write(p)
}

func (out *output) Response(code int) {
	out.responseCode = code
}

func (out *output) Header() http.Header {
	return out.response.Header()
}

func (out *output) WriteString(str string) (n int, err error) {
	return out.buffer.WriteString(str)
}

func (out *output) WriteJSON(i interface{}) []byte {
	out.buffer.Reset()
	bytes, err := json.Marshal(i)
	if err != nil {
		out.response.WriteHeader(500)
		out.WriteString(err.Error())
	} else {
		out.SetContentType(MIME_JSON)
		out.Write(bytes)
	}
	return bytes
}

func (out *output) SetContentType(mime string) {
	out.response.Header().Set("Content-Type", mime)
}

func (out *output) SetCookieValue(name, value string) {
	cookie := &http.Cookie{}
	cookie.Expires = time.Now().Add(time.Hour * 24 * 30)
	cookie.Name = name
	cookie.Value = value
	cookie.Path = "/"
	http.SetCookie(out.response, cookie)
}

func (out *output) Redirect(url ...string) {
	if len(url) == 0 {
		out.response.Header().Set("Location", "/")
	} else {
		out.response.Header().Set("Location", url[0]) //todo: need testing!!!
	}
	out.responseCode = 302
}

func (out *output) AddHeader(name string, value interface{}) {
	out.response.Header().Set(name, fmt.Sprint(value))
}

// write header and send buffer to response writer
func (out *output) Flush() {
	if out.noflush {
		return
	}

	out.noflush = true

	out.response.WriteHeader(out.responseCode)

	// write only if there is something to write
	if out.buffer.Len() > 0 {
		out.response.Write(out.buffer.Bytes())
		out.buffer.Reset()
	}
}

func (out *output) ResponseWriter() http.ResponseWriter {
	return out.response
}

func (out *output) noFlush() {
	out.noflush = true
}
