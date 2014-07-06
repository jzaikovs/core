package core

/*
	TODO: introduce log levels
*/

import (
	"log"
)

var Log = loger{}

type loger struct {
}

func (this *loger) Info(obj ...interface{}) {
	log.Println(obj...)
}

func (this *loger) Error(obj ...interface{}) {
	log.Println(obj...)
}
