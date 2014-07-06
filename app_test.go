package core

import (
	"bytes"
	"encoding/json"
	. "github.com/thejoker0/t"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	//"net/url"
	"io"
	"testing"
)

var _client = &http.Client{}

func init() {
	_client.Jar, _ = cookiejar.New(nil)
}

func _get(url string) *http.Response {
	resp, err := _client.Get(url)
	if err != nil {
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

	APP.Get(`^/test$`, func(in Input, out Output) {
		out.WriteString("hello world")
	})

	APP.Get(`^/test/(\d+)/([^/]+)$`, func(in Input, out Output) {
		out.WriteString(in.Args(0).String())
		out.WriteString(in.Args(1).String())
	})

	APP.Get(`^/json/(\d+)/([^/]+)$`, func(in Input, out Output) {
		out.WriteJSON(map[string]interface{}{
			"0": in.Args(0).Int(),
			"1": in.Args(1).String(),
		})
	})

	APP.Get(`^/isauth$`, func(in Input, out Output) {
		if in.Sess().IsAuth() {
			out.WriteString("auth")
		} else {
			out.WriteString("noauth")
		}
	})

	APP.Get(`^/auth$`, func(in Input, out Output) {
		out.WriteString("hello")
	}).ReqAuth()

	APP.Post(`^/auth$`, func(in Input, out Output) {
		in.Sess().Authorize("")
	}).Need("username", "password")

	APP.Get(`^/limit$`, func(in Input, out Output) {
		out.WriteString("limit")
	}).RateLimit(5, 2)

	go APP.Run()

	for k, v := range map[string]string{
		`http://127.0.0.1:8080/test`:        `hello world`,
		`http://127.0.0.1:8080/test/1/test`: `1test`,
		`http://127.0.0.1:8080/json/1/test`: `{"0":1,"1":"test"}`,
	} {
		if !_read_cmp(_get(k).Body, v) {
			t.Log(k, v)
			t.Fail()
		}
	}

	resp := _get(`http://127.0.0.1:8080/limit`)
	if val := resp.Header.Get(Header_X_Rate_Limit_Limit); val != "5" {
		t.Fatal("Ratelimit not working")
	}

	if val := resp.Header.Get(Header_X_Rate_Limit_Remaining); val != "4" {
		t.Fatal("Ratelimit not working #1")
	}
	resp = _get(`http://127.0.0.1:8080/limit`)
	if val := resp.Header.Get(Header_X_Rate_Limit_Remaining); val != "3" {
		t.Fatal("Ratelimit not working #2")
	}

	for i := 0; i < 3; i++ {
		// reach rate limit
		_get(`http://127.0.0.1:8080/limit`)
	}

	if _get(`http://127.0.0.1:8080/limit`).StatusCode != Response_Too_Many_Requests {
		t.Fatal("Ratelimit reached and no TooManyRequests")
	}

	if _post(`http://127.0.0.1:8080/auth`, Map{"password": "y"}).StatusCode != Response_Unprocessable_Entity {
		t.Fatal("Need not validated")
	}
	if _post(`http://127.0.0.1:8080/auth`, Map{"username": "x"}).StatusCode != Response_Unprocessable_Entity {
		t.Fatal("Need not validated")
	}

	if _get(`http://127.0.0.1:8080/auth`).StatusCode != Response_Unauthorized {
		t.Fatal("ReqAuth not validated")
	}

	if !_read_cmp(_get(`http://127.0.0.1:8080/isauth`).Body, "noauth") {
		t.Fatal("Sess().IsAuth() not working")
	}

	if _post(`http://127.0.0.1:8080/auth`, Map{"username": "x", "password": "y"}).StatusCode != Response_Ok {
		t.Fatal("Post not passed")
	}

	if !_read_cmp(_get(`http://127.0.0.1:8080/auth`).Body, "hello") {
		t.Fatal("Sess().Authorize not working")
	}

	if !_read_cmp(_get(`http://127.0.0.1:8080/isauth`).Body, "auth") {
		t.Fatal("Sess().IsAuth() not working")
	}

}
