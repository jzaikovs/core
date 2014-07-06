package core

/*
	TODO: REVIEW
*/

import (
	"encoding/json"
	"fmt"
)

type RESTResponse struct {
	Status  string `json:"status"`
	Message string `json:"msg"`
}

func (this RESTResponse) Error() string {
	p, _ := json.Marshal(this)
	return string(p)
}

func NewSuccess(message string) RESTResponse {
	return RESTResponse{Status: "SUCCESS", Message: message}
}

func NewErr(msg string) RESTResponse {
	return RESTResponse{Status: "ERROR", Message: msg}
}

func ENotFound() RESTResponse {
	return NewErr("NOT_FOUND")
}

func NewErrf(msg string, args ...interface{}) RESTResponse {
	return NewErr(fmt.Sprintf(msg, args...))
}
