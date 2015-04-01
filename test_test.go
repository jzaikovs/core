package core

import (
	"bytes"
	"fmt"
	. "github.com/jzaikovs/t"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"testing"
)

func simple_resp(ret string) func(context Context) {
	return func(context Context) {
		context.WriteString(ret)
	}
}

func assert(t *testing.T, exp bool, msg string) {
	if !exp {
		t.Error(msg)
	}
}

func assert_s(t *testing.T, a, b string, msg string) {
	t.Log(a)
	t.Log(b)
	if a != b {

		t.Error(msg)
	}
}

// test client structure
type t_test_client struct {
	raw *http.Client
}

// function for creating testing client
func new_t_client() *t_test_client {
	this := new(t_test_client)
	this.raw = new(http.Client)
	this.raw.Jar, _ = cookiejar.New(nil)
	return this
}

// make client get request
func (this *t_test_client) get(query string) string {
	resp, err := this.raw.Get(test_server_addr + query)
	if err != nil {
		return "ERR"
	}

	defer resp.Body.Close()

	p, _ := ioutil.ReadAll(resp.Body)

	return fmt.Sprintf("%d:%s", resp.StatusCode, string(p))
}

// make client post request
func (this *t_test_client) post(query string, data Map) string {
	b := new(bytes.Buffer)
	b.WriteString(_to_json(data))

	resp, err := this.raw.Post(test_server_addr+query, ContentType_JSON, b)
	if err != nil {
		return "ERR"
	}

	defer resp.Body.Close()

	p, _ := ioutil.ReadAll(resp.Body)

	return fmt.Sprintf("%d:%s", resp.StatusCode, string(p))
}

func (this *t_test_client) cookie(name string) string {
	u, err := url.Parse(test_server_addr)
	if err != nil {
		panic(err)
	}
	for _, c := range this.raw.Jar.Cookies(u) {
		if c.Name == name {
			return c.Value
		}
	}
	return ""
}
