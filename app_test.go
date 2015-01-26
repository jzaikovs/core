package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	. "github.com/jzaikovs/t"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"testing"
	"time"
)

var _client = &http.Client{}

func init() {
	_client.Jar, _ = cookiejar.New(nil)
}

func _get(t *testing.T, url string) *http.Response {
	resp, err := _client.Get(url)
	if err != nil {
		t.Error(err)
		return nil
	}
	return resp
}

func _post(url string, data Map) *http.Response {
	b := new(bytes.Buffer)
	b.WriteString(_to_json(data))
	resp, err := _client.Post(url, ContentType_JSON, b)
	if err != nil {
		return nil
	}
	return resp
}

func _read_cmp(r io.Reader, want string) bool {
	b, _ := ioutil.ReadAll(r)
	return string(b) == want
}

func _to_json(x interface{}) string {
	p, _ := json.MarshalIndent(x, "", "  ")
	return string(p)
}

func TestBasic(t *testing.T) {

	APP.Get(`^/test$`, func(context Context) {
		context.WriteString("hello world")
	})

	APP.Get(`^/test/(\d+)/([^/]+)$`, func(context Context) {
		context.WriteString(context.Args(0).String())
		context.WriteString(context.Args(1).String())
	})

	APP.Get(`^/json/(\d+)/([^/]+)$`, func(context Context) {
		context.WriteJSON(map[string]interface{}{
			"0": context.Args(0).Int(),
			"1": context.Args(1).String(),
		})
	})

	APP.Get(`^/isauth$`, func(context Context) {
		if context.Session().IsAuth() {
			context.WriteString("auth")
		} else {
			context.WriteString("noauth")
		}
	})

	APP.Get(`^/auth$`, func(context Context) {
		context.WriteString("hello")
	}).ReqAuth()

	APP.Post(`^/auth$`, func(context Context) {
		context.Session().Authorize("")
	}).Need("username", "password")

	APP.Get(`^/limit$`, func(context Context) {
		context.WriteString("limit")
	}).RateLimit(5, 2)

	// testing limited only to localhost
	APP.Config = new_t_config()
	APP.Config.Host = "127.0.0.1"
	APP.Config.Port = 8080

	addr := fmt.Sprintf("%s:%d", APP.Config.Host, APP.Config.Port)

	go APP.Run()
	time.Sleep(time.Second)

	for k, v := range map[string]string{
		`http://` + addr + `/test`:        `hello world`,
		`http://` + addr + `/test/1/test`: `1test`,
		`http://` + addr + `/json/1/test`: `{"0":1,"1":"test"}`,
	} {
		if !_read_cmp(_get(t, k).Body, v) {
			t.Log(k, v)
			t.Fail()
		}
	}

	resp := _get(t, `http://127.0.0.1:8080/limit`)
	if val := resp.Header.Get(Header_X_Rate_Limit_Limit); val != "5" {
		t.Fatal("Ratelimit not working", val)
	}

	if val := resp.Header.Get(Header_X_Rate_Limit_Remaining); val != "4" {
		t.Fatal("Ratelimit not working #1", val)
	}
	resp = _get(t, `http://127.0.0.1:8080/limit`)
	if val := resp.Header.Get(Header_X_Rate_Limit_Remaining); val != "3" {
		t.Fatal("Ratelimit not working #2", val)
	}

	for i := 0; i < 3; i++ {
		// reach rate limit
		_get(t, `http://127.0.0.1:8080/limit`)
	}

	if _get(t, `http://127.0.0.1:8080/limit`).StatusCode != Response_Too_Many_Requests {
		t.Fatal("Ratelimit reached and no TooManyRequests")
	}

	if _post(`http://127.0.0.1:8080/auth`, Map{"password": "y"}).StatusCode != Response_Bad_Request {
		t.Fatal("Need not validated")
	}

	if _post(`http://127.0.0.1:8080/auth`, Map{"username": "x"}).StatusCode != Response_Bad_Request {
		t.Fatal("Need not validated")
	}

	if _get(t, `http://127.0.0.1:8080/auth`).StatusCode != Response_Unauthorized {
		t.Fatal("ReqAuth not validated")
	}

	if !_read_cmp(_get(t, `http://127.0.0.1:8080/isauth`).Body, "noauth") {
		t.Fatal("Session().IsAuth() not working")
	}

	if _post(`http://127.0.0.1:8080/auth`, Map{"username": "x", "password": "y"}).StatusCode != Response_Ok {
		t.Fatal("Post not passed")
	}

	if !_read_cmp(_get(t, `http://127.0.0.1:8080/auth`).Body, "hello") {
		t.Fatal("Session().Authorize not working")
	}

	if !_read_cmp(_get(t, `http://127.0.0.1:8080/isauth`).Body, "auth") {
		t.Fatal("Session().IsAuth() not working")
	}
}
