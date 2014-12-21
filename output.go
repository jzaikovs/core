package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

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

	no_flush()
}

type output struct {
	response      http.ResponseWriter
	buffer        bytes.Buffer
	response_code int
	noflush       bool
}

func new_output(response http.ResponseWriter) *output {
	this := &output{response: response, response_code: 200}
	this.SetContentType(MIME_HTML)
	this.AddHeader("Access-Control-Allow-Origin", "*")
	return this
}

func (this *output) Write(p []byte) (n int, err error) {
	return this.buffer.Write(p)
}

func (this *output) Response(code int) {
	this.response_code = code
}

func (this *output) Header() http.Header {
	return this.response.Header()
}

func (this *output) WriteString(str string) (n int, err error) {
	return this.buffer.WriteString(str)
}

func (this *output) WriteJSON(i interface{}) []byte {
	this.buffer.Reset()
	bytes, err := json.Marshal(i)
	if err != nil {
		this.response.WriteHeader(500)
		this.WriteString(err.Error())
	} else {
		this.SetContentType(MIME_JSON)
		this.Write(bytes)
	}
	return bytes
}

func (this *output) SetContentType(mime string) {
	this.response.Header().Set("Content-Type", mime)
}

func (this *output) SetCookieValue(name, value string) {
	cookie := &http.Cookie{}
	cookie.Expires = time.Now().Add(time.Hour * 24 * 30)
	cookie.Name = name
	cookie.Value = value
	cookie.Path = "/"
	http.SetCookie(this.response, cookie)
}

func (this *output) Redirect(url ...string) {
	if len(url) == 0 {
		this.response.Header().Set("Location", "/")
	} else {
		this.response.Header().Set("Location", url[0]) //todo: need testing!!!
	}
	this.response_code = 302
}

func (this *output) AddHeader(name string, value interface{}) {
	this.response.Header().Set(name, fmt.Sprint(value))
}

// write header and send buffer to response writer
func (this *output) Flush() {
	if this.noflush {
		return
	}

	this.response.WriteHeader(this.response_code)

	// write only if there is something to write
	if this.buffer.Len() > 0 {
		this.response.Write(this.buffer.Bytes())
		this.buffer.Reset()
	}
}

func (this *output) ResponseWriter() http.ResponseWriter {
	return this.response
}

func (this *output) no_flush() {
	this.noflush = true
}
