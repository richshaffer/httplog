package httplog

import "net/http"

// Response records information from the HTTP server response.
type Response struct {
	Status   int
	Size     int64
	Hijacked bool
	Header   http.Header
}

// Reset sets the receiver to its zero value.
func (r *Response) Reset() {
	*r = Response{}
}

// Update copies values from a ResponseWriter to the receiver.
func (r *Response) Update(w ResponseWriter) {
	r.Status = w.Status()
	r.Size = w.Size()
	r.Hijacked = w.Hijacked()
	r.Header = w.Header()
}
