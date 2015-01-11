package loggy

import (
	"fmt"
	"log"
	"time"
)

type data struct {
	tx     time.Time
	tag    string
	format string
	args   []interface{}
}

var (
	q       = make(chan data)
	running bool
)

func init() {
	Start()
}

func Start() {
	if running {
		return
	}
	fmt.Println("loggy starting...")
	go func() {
		for {
			// reatriev next log in line
			d, ok := <-q
			if !ok {
				break
			}

			if len(d.format) == 0 {
				log.Println(d.tx.UnixNano(), d.tag, fmt.Sprint(d.args))
			} else {
				log.Println(d.tx.UnixNano(), d.tag, fmt.Sprintf(d.format, d.args...))
			}
		}
	}()

	running = true
}

// function for logging data
func Log(tag string, args ...interface{}) {
	q <- data{time.Now(), tag, "", args}
}

// function for lggin data with specific format
func Logf(tag string, format string, args ...interface{}) {
	q <- data{time.Now(), tag, format, args}
}

func Info(obj ...interface{}) {
	Log("INFO", obj...)
}

func Error(obj ...interface{}) {
	Log("ERR", obj...)
}
