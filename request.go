package httplog

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Request contains information from a client request.
type Request struct {
	Method        string
	URI           string
	URL           *url.URL
	Proto         string
	Header        http.Header
	ContentLength int64
	Host          string
	RemoteAddr    string
	User          string
}

// NewRequest returns a new Request from an *http.Request variable.
func NewRequest(req *http.Request) *Request {
	r := new(Request)
	r.Update(req)
	return r
}

// Reset sets the receiver to its zero value.
func (r *Request) Reset() {
	*r = Request{}
}

// Update sets values on the receiver from an *http.Request value.
func (r *Request) Update(req *http.Request) {
	r.Method = req.Method
	r.URI = req.RequestURI
	r.URL = req.URL
	r.Proto = req.Proto
	r.Header = req.Header
	r.ContentLength = req.ContentLength
	r.Host = req.Host
	r.RemoteAddr = req.RemoteAddr
	r.User, _, _ = req.BasicAuth()
}

// FirstForwardedFor attempts to parse an IP address from Forwarded and
// X-Forwarded-For headers. If no client IP address is found in those headers,
// it returns an empty string.
func (r *Request) FirstForwardedFor() string {
	forwardHeaders := r.Header["Forwarded"]
	for i := range forwardHeaders {
		pairs, _ := ParsePairs(forwardHeaders[i], true)
		if f := pairs["for"]; f != "" {
			return f
		}
	}
	forwardHeaders = r.Header["X-Forwarded-For"]
	for i := range forwardHeaders {
		hops := strings.SplitN(forwardHeaders[i], ",", 2)
		if len(hops) != 0 {
			return strings.TrimSpace(hops[0])
		}
	}
	return ""
}

// ClientAddr returns the IP address (or possibly host name) of the client for
// the request. If an identifier is found in Forwarded or X-Forwarded-For
// headers, it is returned. Otherwise, the remote IP address of the connection
// the request was received on is returned.
func (r *Request) ClientAddr() string {
	if f := r.FirstForwardedFor(); f != "" {
		return f
	}
	return r.RemoteAddr
}

// ParsePairs parses 'token=quoted-string' pairs from HTTP headers. The first
// parameter is the header value without the header name. The second parameter
// controls case-insensitivity. If it is true, all keys in the returned map
// will be lowercased.
func ParsePairs(text string, ci bool) (map[string]string, error) {
	// This is more complex than splitting on ';' and '=', because
	// quoted-strings may contain both of those characters.
	var pairs map[string]string
	s, e := 0, len(text)
	for s < e {
		var ns, ne, vs, ve int // name start, name end, val start, val end
		ns = s
		for ne = s; ne < e && text[ne] != '=' && text[ne] != ';'; ne++ {
		}
		if ne == ns {
			return nil, fmt.Errorf("empty attribute name at %d", ns)
		}
		if ne == e {
			return nil, errors.New("no '=' before end of string")
		}
		if text[ne] == ';' {
			return nil, fmt.Errorf("found unexpected ';' at %d", ne)
		}
		if vs = ne + 1; vs < e && text[vs] == '"' {
			// value is a quoted-string; find end quote
			for vs, ve = vs+1, vs+2; ve < e && text[ve] != '"'; ve++ {
			}
			// after closing quote should be end of string or comma
			if s = ve + 1; s < len(text) && text[s] != ';' {
				return nil, fmt.Errorf("trailing data after quoted string at %d", s)
			}
			s++ // advance to one character after comma (ok if end-of-string)
		} else {
			// not a quoted string; find next comma or end of string.
			for ve = vs; ve < e && text[ve] != ';'; ve++ {
			}
			s = ve + 1 // advance s to one character after comma
		}
		if vs == ve {
			return nil, fmt.Errorf("empty attribute value at %d", vs)
		}
		// at end of string, s == e if last character is a comma, else e + 1
		if s == e {
			return nil, fmt.Errorf("trailing ';' at %d", s)
		}
		if pairs == nil {
			pairs = make(map[string]string, 1)
		}
		if ci {
			pairs[strings.ToLower(text[ns:ne])] = text[vs:ve]
		} else {
			pairs[text[ns:ne]] = text[vs:ve]
		}
	}
	return pairs, nil
}
