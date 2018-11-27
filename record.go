package httplog

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Record represents data collected from processing one HTTP request.
type Record struct {
	Request
	Response
	StartTime, EndTime time.Time
	Duration           time.Duration
}

// Reset resets the received to its zero value.
func (r *Record) Reset() {
	r.Request.Reset()
	r.Response.Reset()
	r.StartTime, r.EndTime, r.Duration = time.Time{}, time.Time{}, 0
}

// Start should be called before processing a request to record the start time.
func (r *Record) Start() {
	r.StartTime = time.Now()
}

// End should be called when done processing a request to update the end time
// and duration.
func (r *Record) End() {
	r.EndTime = time.Now()
	r.Duration = r.EndTime.Sub(r.StartTime)
}

const (
	// SimpleLogFormat provides a basic log format.
	SimpleLogFormat = `%a "%r" %s %B`
	// BasicLogFormat provides a moderately verbose format.
	BasicLogFormat = `%a %v %u "%r" %s %T %B "%{User-Agent}i"`
	// CommonLogFormat is a format string that provides Common Log Format
	// output.
	CommonLogFormat = `%a - %u %t "%r" %s %B`
	// NCSALogFormat can be used to provide NCSA extended log format that
	// includes referer and user-agent headers.
	NCSALogFormat = `%a - %u %t "%r" %s %B "%{Referer}i" "%{User-Agent}i"`
)

// Format formats the log record according to a format string. The format
// directives are loosely based on Apache HTTP Server's mod_log_config:
//   %% - A literal '%'.
//   %B - The size in bytes of the response body, not including headers.
//   %D - The duration of the request, in microseconds (as a floating point
//        value).
//   %H - The request protocol, e.g., "HTTP/1.1".
//   %T - The duration of the request, in seconds (as a floating point value).
//   %U - The URL path requested, without any query string.
//   %a - The client IP address. If the request contains a Forwarded or
//        X-Forwarded-For header, that address (or name) will be used.
//        Otherwise, the value will be the remote IP address of the connection.
//   %m - The request method, e.g. "GET".
//   %q - The URL query, if any, including the leading '?'.
//   %r - The first line of the request, e.g., "GET /path HTTP/1.1".
//   %s - The numeric response status code.
//   %t - The request start time, in format "[02/Jan/2006:15:04:05 -0700]".
//   %u - The user name from the request, if any.
//   %v - The server name from the Host header or request URL.
//   %{NAME}C - The value of the cookie with name NAME (case-sensitive).
//   %{UNIT}T - The request duration in the given UNIT. UNIT must be one of
//              "ms", "us" or "s" for microseconds, milliseconds or seconds.
//   %{NAME}i - The value of the request header with the given name (case-
//              insensitive).
//   %{NAME}o - The value of the response header with the given name (case-
//              insensitive). Note that this won't include headers added by the
//              http package automatically, such as Date, Content-Length,
//              Content-Type, Transfer-Encoding and Connection.
//   %{FORMAT}t - The request time in the provided FORMAT. FORMAT should be a
//                format string understood by time.Time.Format. If the format
//                begins with 'end:', the time will be when the request
//                finished. If the format begins with 'begin:' or has no prefix,
//                the time will be when the request was started.
//
// Invalid format directives will be passed through unchanged.
func (r *Record) Format(format string) string {
	var b strings.Builder
	for i, l := 0, len(format); i < l; i++ {
		switch format[i] {
		case '%':
			if i++; i == l {
				b.WriteByte('%')
				return b.String()
			}
			switch format[i] {
			case '%':
				b.WriteByte('%')
			case 'B':
				b.WriteString(strconv.FormatInt(r.Size, 10))
			case 'D':
				ms := float64(r.Duration) / float64(time.Microsecond)
				b.WriteString(strconv.FormatFloat(ms, 'f', -1, 64))
			case 'H':
				b.WriteString(r.Proto)
			case 'T':
				s := r.Duration.Seconds()
				b.WriteString(strconv.FormatFloat(s, 'f', -1, 64))
			case 'U':
				b.WriteString(r.URL.Path)
			case 'a':
				b.WriteString(r.ClientAddr())
			case 'm':
				b.WriteString(r.Method)
			case 'q':
				if r.URL.RawQuery != "" {
					b.WriteByte('?')
					b.WriteString(r.URL.RawQuery)
				}
			case 'r':
				b.WriteString(r.Method)
				b.WriteByte(' ')
				b.WriteString(r.Proto)
				b.WriteByte(' ')
				b.WriteString(r.URL.String())
			case 's':
				b.WriteString(strconv.FormatInt(int64(r.Status), 10))
			case 't':
				b.WriteString(r.StartTime.Format("[02/Jan/2006:15:04:05 -0700]"))
			case 'u':
				if r.Request.User != "" {
					b.WriteString(r.Request.User)
				} else {
					b.WriteByte('-')
				}
			case 'v':
				b.WriteString(r.Host)
			case '{':
				j := i + 1
				for ; j < l && format[j] != '}'; j++ {
				}
				if j == l {
					b.WriteByte('%')
					b.WriteString(format[i:])
					return b.String()
				}
				key := format[i+1 : j]
				j++
				if j == l {
					b.WriteByte('%')
					b.WriteString(format[i:])
					return b.String()
				}
				i = j
				switch format[j] {
				case 'C':
					cookies, _ := ParsePairs(r.Request.Header.Get("Cookie"), false)
					b.WriteString(cookies[key])
				case 'T':
					switch key {
					case "ms":
						ms := float64(r.Duration) / float64(time.Millisecond)
						b.WriteString(strconv.FormatFloat(ms, 'f', -1, 64))
					case "us":
						ms := float64(r.Duration) / float64(time.Microsecond)
						b.WriteString(strconv.FormatFloat(ms, 'f', -1, 64))
					case "s":
						s := r.Duration.Seconds()
						b.WriteString(strconv.FormatFloat(s, 'f', -1, 64))
					default:
						b.WriteString("%{")
						b.WriteString(key)
						b.WriteString("}T")
					}
				case 'i':
					headers := r.Request.Header[http.CanonicalHeaderKey(key)]
					b.WriteString(strings.Join(headers, ","))
				case 'o':
					headers := r.Response.Header[http.CanonicalHeaderKey(key)]
					b.WriteString(strings.Join(headers, ","))
				case 't':
					if strings.HasPrefix(key, "end:") {
						b.WriteString(r.EndTime.Format(strings.TrimPrefix(key, "end:")))
					} else {
						b.WriteString(r.StartTime.Format(strings.TrimPrefix(key, "begin:")))
					}
				default:
					b.WriteString("%{")
					b.WriteString(key)
					b.WriteByte('}')
					b.WriteByte(format[j])
				}
			default:
				b.WriteByte('%')
				b.WriteByte(format[i])
			}
		default:
			b.WriteByte(format[i])
		}
	}
	return b.String()
}
