package core

import (
	. "github.com/jzaikovs/t"
	"testing"
)

func TestCSRF(t *testing.T) {
	c := new_t_client()

	query := "/csrf"

	// 1. good situation, emit csrf then request
	test_app.Get(query, simple_resp("req#1")).CSRF(true, false)
	test_app.Post(query, simple_resp("req#2")).CSRF(false, true)

	assert_s(t, c.get(query), "200:req#1", "Got bad get response, wanted 200")
	assert(t, c.cookie("_csrf") != "", "Emited empty CSRF cookie")
	assert_s(t, c.post(query, Map{"x": "y"}), "200:req#2", "Got bad post response, wanted 200")
	assert(t, c.cookie("_csrf") == "", "After consuming cookie, response did not set cookie to empty")
	assert_s(t, c.post(query, Map{"x": "y"}), "403:", "Got goode response, wanted 403")
}

func TestMatchNeed(t *testing.T) {
	c := new_t_client()

	query := "/match"

	// 1. good situation, emit csrf then request
	test_app.Post(query, simple_resp("req#3")).Match("x", "y").Need("x", "y")

	assert_s(t, c.post(query, Map{"x": "1", "y": "1"}), "200:req#3", "Bad post request")
	assert_s(t, c.post(query, Map{"x": "1", "y": "2"}), `400:{"code":400,"error":"field [x] not match field [y]"}`, "Bad post request")
	assert_s(t, c.post(query, Map{"a": "1", "b": "1"}), `400:{"code":400,"error":"field [x] required"}`, "Bad post request")
}
