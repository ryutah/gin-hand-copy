package gin

import (
	"bufio"
	"io"
	"net"
	"net/http"
)

const (
	noWritten     = -1
	defaultStatus = http.StatusOK
)

type responseWriterBase interface {
	http.ResponseWriter
	http.Hijacker
	http.Flusher
	http.CloseNotifier

	// Returns the HTTP response status code of the current request.
	Status() int

	// Returns the number of bytes already written into the response http body.
	// See Written()
	Size() int

	// Writes the string into the response body.
	WriteString(string) (int, error)

	// Returns true if the response body was already written.
	Written() bool

	// Forces to write the http header (status code + headers).
	WriteHeaderNow()
}

type responseWriter struct {
	http.ResponseWriter
	size   int
	status int
}

var _ ResponseWriter = &responseWriter{}

func (r *responseWriter) reset(writer http.ResponseWriter) {
	r.ResponseWriter = writer
	r.size = noWritten
	r.status = defaultStatus
}

func (r *responseWriter) WriteHeader(code int) {
	if code > 0 && r.status != code {
		if r.Written() {
			// Debug print
		}
		r.status = code
	}
}

func (r *responseWriter) WriteHeaderNow() {
	if !r.Written() {
		r.size = 0
		r.ResponseWriter.WriteHeader(r.status)
	}
}

func (r *responseWriter) Write(data []byte) (n int, err error) {
	r.WriteHeaderNow()
	n, err = r.ResponseWriter.Write(data)
	r.size += n
	return
}

func (r *responseWriter) WriteString(s string) (n int, err error) {
	r.WriteHeaderNow()
	n, err = io.WriteString(r.ResponseWriter, s)
	r.size += n
	return
}

func (r *responseWriter) Status() int {
	return r.status
}

func (r *responseWriter) Size() int {
	return r.size
}

func (r *responseWriter) Written() bool {
	return r.size != noWritten
}

// Hijack implements the http.Hijacker interface.
func (r *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if r.size < 0 {
		r.size = 0
	}
	return r.ResponseWriter.(http.Hijacker).Hijack()
}

// CloseNotifier implements the http.CloseNotifier interface.
func (r *responseWriter) CloseNotify() <-chan bool {
	return r.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

// Flush implements the http.Flush interface.
func (r *responseWriter) Flush() {
	r.WriteHeaderNow()
	r.ResponseWriter.(http.Flusher).Flush()
}
