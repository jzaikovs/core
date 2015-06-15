package loggy

import (
	"io"
	"io/ioutil"
	"log"
	"os"
)

var Trace = log.New(ioutil.Discard, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)

var Info = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

var Warning = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)

var Error = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

func EnableTrace(dest io.Writer) {
	Trace = log.New(dest, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
}
