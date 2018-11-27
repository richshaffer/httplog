// Package httplog implements logging of HTTP requests.
package httplog

import (
	"log"
	"net/http"
	"sync"
)

// LogFn is a function responsible for logging an HTTP request/response.
type LogFn func(*Record)

// LoggingHandler wraps an http.Handler in order to log processed requests
// using the provided function. If the logging function needs to reference
// the passed in *Record, it must make a copy before returning.
type LoggingHandler struct {
	http.Handler
	LogFn
}

// NewLoggingHandler returns an http.Handler that logs completed requests
// using the given LogFn. If the second parameter is nil, it uses DefaultLogFn.
func NewLoggingHandler(handler http.Handler, fn LogFn) http.Handler {
	if fn == nil {
		return &LoggingHandler{Handler: handler, LogFn: DefaultLogFn}
	}
	return &LoggingHandler{Handler: handler, LogFn: fn}
}

func (l *LoggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	record := recordPool.Get().(*Record)
	record.Start()
	record.Request.Update(r)
	rw := WrapResponseWriter(w)
	l.Handler.ServeHTTP(rw, r)
	record.Response.Update(rw)
	record.End()
	if l.LogFn != nil {
		l.LogFn(record)
	}
	record.Reset()
	recordPool.Put(record)
}

// DefaultLogFn logs the record to log.Println using BasicLogFormat.
func DefaultLogFn(record *Record) {
	log.Println(record.Format(BasicLogFormat))
}

var recordPool = sync.Pool{New: func() interface{} { return new(Record) }}
