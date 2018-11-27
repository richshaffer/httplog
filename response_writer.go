package httplog

import (
	"bufio"
	"net"
	"net/http"
)

// ResponseWriter augments the http.ResponseWriter type to enable getting the
// HTTP status code, size of the response and whether or not the connection has
// been hijacked.
type ResponseWriter interface {
	http.ResponseWriter
	Status() int
	Size() int64
	Hijacked() bool
}

// WrapResponseWriter wraps the http.ResponseWriter in a type that preserves
// implementations of http.ResponseWriter, http.CloseNotifier, http.Flusher,
// http.Hijacker and http.Pusher interfaces. For a given interface,
//   val, ok := rw.(http.InterfaceType)
// will return the same result as
//   val, ok := WrapResponseWriter(rw).(http.InterfaceType)
// This enables collecting logging statistics without losing the functionality
// provided by the interfaces.
func WrapResponseWriter(rw http.ResponseWriter) ResponseWriter {
	i := 0
	if _, ok := rw.(http.CloseNotifier); ok {
		i |= closeNotifier
	}
	if _, ok := rw.(http.Flusher); ok {
		i |= flusher
	}
	if _, ok := rw.(http.Hijacker); ok {
		i |= hijacker
	}
	if _, ok := rw.(http.Pusher); ok {
		i |= pusher
	}
	return types[i](rw)
}

const (
	closeNotifier int = 1 << iota
	flusher
	hijacker
	pusher
)

var types = [16]func(http.ResponseWriter) ResponseWriter{
	func(rw http.ResponseWriter) ResponseWriter {
		return &responseWriter{responseWriter: rw}
	},
	func(rw http.ResponseWriter) ResponseWriter {
		return responseWriterCloseNotifier{
			&responseWriter{responseWriter: rw},
		}
	},
	func(rw http.ResponseWriter) ResponseWriter {
		return responseWriterFlusher{
			&responseWriter{responseWriter: rw},
		}
	},
	func(rw http.ResponseWriter) ResponseWriter {
		return responseWriterCloseNotifierFlusher{
			&responseWriter{responseWriter: rw},
		}
	},
	func(rw http.ResponseWriter) ResponseWriter {
		return responseWriterHijacker{
			&responseWriter{responseWriter: rw},
		}
	},
	func(rw http.ResponseWriter) ResponseWriter {
		return responseWriterCloseNotifierHijacker{
			&responseWriter{responseWriter: rw},
		}
	},
	func(rw http.ResponseWriter) ResponseWriter {
		return responseWriterFlusherHijacker{
			&responseWriter{responseWriter: rw},
		}
	},
	func(rw http.ResponseWriter) ResponseWriter {
		return responseWriterCloseNotifierFlusherHijacker{
			&responseWriter{responseWriter: rw},
		}
	},
	func(rw http.ResponseWriter) ResponseWriter {
		return responseWriterPusher{
			&responseWriter{responseWriter: rw},
		}
	},
	func(rw http.ResponseWriter) ResponseWriter {
		return responseWriterCloseNotifierPusher{
			&responseWriter{responseWriter: rw},
		}
	},
	func(rw http.ResponseWriter) ResponseWriter {
		return responseWriterFlusherPusher{
			&responseWriter{responseWriter: rw},
		}
	},
	func(rw http.ResponseWriter) ResponseWriter {
		return responseWriterCloseNotifierFlusherPusher{
			&responseWriter{responseWriter: rw},
		}
	},
	func(rw http.ResponseWriter) ResponseWriter {
		return responseWriterHijackerPusher{
			&responseWriter{responseWriter: rw},
		}
	},
	func(rw http.ResponseWriter) ResponseWriter {
		return responseWriterCloseNotifierHijackerPusher{
			&responseWriter{responseWriter: rw},
		}
	},
	func(rw http.ResponseWriter) ResponseWriter {
		return responseWriterFlusherHijackerPusher{
			&responseWriter{responseWriter: rw},
		}
	},
	func(rw http.ResponseWriter) ResponseWriter {
		return responseWriterCloseNotifierFlusherHijackerPusher{
			&responseWriter{responseWriter: rw},
		}
	},
}

//
type responseWriter struct {
	responseWriter http.ResponseWriter
	status         int
	size           int64
	hijacked       bool
}

func (r *responseWriter) Write(p []byte) (int, error) {
	n, err := r.responseWriter.Write(p)
	r.size += int64(n)
	return n, err
}

func (r *responseWriter) WriteHeader(statusCode int) {
	r.status = statusCode
	r.responseWriter.WriteHeader(statusCode)
}

func (r *responseWriter) Header() http.Header {
	return r.responseWriter.Header()
}

func (r *responseWriter) Status() int {
	if r.status == 0 {
		return 200
	}
	return r.status
}

func (r *responseWriter) Size() int64 {
	return r.size
}

func (r *responseWriter) Hijacked() bool {
	return r.hijacked
}

//
type responseWriterCloseNotifier struct {
	*responseWriter
}

func (r responseWriterCloseNotifier) CloseNotify() <-chan bool {
	return r.responseWriter.responseWriter.(http.CloseNotifier).CloseNotify()
}

//
type responseWriterFlusher struct {
	*responseWriter
}

func (r responseWriterFlusher) Flush() {
	r.responseWriter.responseWriter.(http.Flusher).Flush()
}

//
type responseWriterCloseNotifierFlusher struct {
	*responseWriter
}

func (r responseWriterCloseNotifierFlusher) CloseNotify() <-chan bool {
	return r.responseWriter.responseWriter.(http.CloseNotifier).CloseNotify()
}

func (r responseWriterCloseNotifierFlusher) Flush() {
	r.responseWriter.responseWriter.(http.Flusher).Flush()
}

//
type responseWriterHijacker struct {
	*responseWriter
}

func (r responseWriterHijacker) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	r.responseWriter.hijacked = true
	return r.responseWriter.responseWriter.(http.Hijacker).Hijack()
}

//
type responseWriterCloseNotifierHijacker struct {
	*responseWriter
}

func (r responseWriterCloseNotifierHijacker) CloseNotify() <-chan bool {
	return r.responseWriter.responseWriter.(http.CloseNotifier).CloseNotify()
}

func (r responseWriterCloseNotifierHijacker) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	r.responseWriter.hijacked = true
	return r.responseWriter.responseWriter.(http.Hijacker).Hijack()
}

//
type responseWriterFlusherHijacker struct {
	*responseWriter
}

func (r responseWriterFlusherHijacker) Flush() {
	r.responseWriter.responseWriter.(http.Flusher).Flush()
}

func (r responseWriterFlusherHijacker) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	r.responseWriter.hijacked = true
	return r.responseWriter.responseWriter.(http.Hijacker).Hijack()
}

//
type responseWriterCloseNotifierFlusherHijacker struct {
	*responseWriter
}

func (r responseWriterCloseNotifierFlusherHijacker) CloseNotify() <-chan bool {
	return r.responseWriter.responseWriter.(http.CloseNotifier).CloseNotify()
}

func (r responseWriterCloseNotifierFlusherHijacker) Flush() {
	r.responseWriter.responseWriter.(http.Flusher).Flush()
}

func (r responseWriterCloseNotifierFlusherHijacker) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	r.responseWriter.hijacked = true
	return r.responseWriter.responseWriter.(http.Hijacker).Hijack()
}

//
type responseWriterPusher struct {
	*responseWriter
}

func (r responseWriterPusher) Push(target string, opts *http.PushOptions) error {
	// http.Server will start a new request handler for this which will be
	// logged separately.
	return r.responseWriter.responseWriter.(http.Pusher).Push(target, opts)
}

//
type responseWriterCloseNotifierPusher struct {
	*responseWriter
}

func (r responseWriterCloseNotifierPusher) CloseNotify() <-chan bool {
	return r.responseWriter.responseWriter.(http.CloseNotifier).CloseNotify()
}

func (r responseWriterCloseNotifierPusher) Push(target string, opts *http.PushOptions) error {
	// http.Server will start a new request handler for this which will be
	// logged separately.
	return r.responseWriter.responseWriter.(http.Pusher).Push(target, opts)
}

//
type responseWriterFlusherPusher struct {
	*responseWriter
}

func (r responseWriterFlusherPusher) Flush() {
	r.responseWriter.responseWriter.(http.Flusher).Flush()
}

func (r responseWriterFlusherPusher) Push(target string, opts *http.PushOptions) error {
	// http.Server will start a new request handler for this which will be
	// logged separately.
	return r.responseWriter.responseWriter.(http.Pusher).Push(target, opts)
}

//
type responseWriterCloseNotifierFlusherPusher struct {
	*responseWriter
}

func (r responseWriterCloseNotifierFlusherPusher) CloseNotify() <-chan bool {
	return r.responseWriter.responseWriter.(http.CloseNotifier).CloseNotify()
}

func (r responseWriterCloseNotifierFlusherPusher) Flush() {
	r.responseWriter.responseWriter.(http.Flusher).Flush()
}

func (r responseWriterCloseNotifierFlusherHijacker) Push(target string, opts *http.PushOptions) error {
	// http.Server will start a new request handler for this which will be
	// logged separately.
	return r.responseWriter.responseWriter.(http.Pusher).Push(target, opts)
}

//
type responseWriterHijackerPusher struct {
	*responseWriter
}

func (r responseWriterHijackerPusher) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	r.responseWriter.hijacked = true
	return r.responseWriter.responseWriter.(http.Hijacker).Hijack()
}

func (r responseWriterHijackerPusher) Push(target string, opts *http.PushOptions) error {
	// http.Server will start a new request handler for this which will be
	// logged separately.
	return r.responseWriter.responseWriter.(http.Pusher).Push(target, opts)
}

//
type responseWriterCloseNotifierHijackerPusher struct {
	*responseWriter
}

func (r responseWriterCloseNotifierHijackerPusher) CloseNotify() <-chan bool {
	return r.responseWriter.responseWriter.(http.CloseNotifier).CloseNotify()
}

func (r responseWriterCloseNotifierHijackerPusher) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	r.responseWriter.hijacked = true
	return r.responseWriter.responseWriter.(http.Hijacker).Hijack()
}

func (r responseWriterCloseNotifierHijackerPusher) Push(target string, opts *http.PushOptions) error {
	// http.Server will start a new request handler for this which will be
	// logged separately.
	return r.responseWriter.responseWriter.(http.Pusher).Push(target, opts)
}

//
type responseWriterFlusherHijackerPusher struct {
	*responseWriter
}

func (r responseWriterFlusherHijackerPusher) Flush() {
	r.responseWriter.responseWriter.(http.Flusher).Flush()
}

func (r responseWriterFlusherHijackerPusher) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	r.responseWriter.hijacked = true
	return r.responseWriter.responseWriter.(http.Hijacker).Hijack()
}

func (r responseWriterFlusherHijackerPusher) Push(target string, opts *http.PushOptions) error {
	// http.Server will start a new request handler for this which will be
	// logged separately.
	return r.responseWriter.responseWriter.(http.Pusher).Push(target, opts)
}

//
type responseWriterCloseNotifierFlusherHijackerPusher struct {
	*responseWriter
}

func (r responseWriterCloseNotifierFlusherHijackerPusher) CloseNotify() <-chan bool {
	return r.responseWriter.responseWriter.(http.CloseNotifier).CloseNotify()
}

func (r responseWriterCloseNotifierFlusherHijackerPusher) Flush() {
	r.responseWriter.responseWriter.(http.Flusher).Flush()
}

func (r responseWriterCloseNotifierFlusherHijackerPusher) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	r.responseWriter.hijacked = true
	return r.responseWriter.responseWriter.(http.Hijacker).Hijack()
}

func (r responseWriterCloseNotifierFlusherHijackerPusher) Push(target string, opts *http.PushOptions) error {
	// http.Server will start a new request handler for this which will be
	// logged separately.
	return r.responseWriter.responseWriter.(http.Pusher).Push(target, opts)
}
