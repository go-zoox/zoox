package zoox

import (
	"bufio"
	"net"
	"net/http"

	"github.com/go-zoox/logger"
)

const (
	noWritten = -1
	// io.Copy response write will not trigger writeHeader => default status should be 200
	defaultStatus = http.StatusOK
)

// ResponseWriter ...
type ResponseWriter interface {
	http.ResponseWriter
	http.Hijacker
	http.CloseNotifier
	http.Flusher

	// Status returns the HTTP response status code of the current request.
	Status() int

	// Size returns the number of bytes already written into the response http body.
	// See Written()
	Size() int

	// WriteString writes the string into the response body.
	WriteString(string) (int, error)

	// Written returns true if the response body was already written.
	Written() bool

	// Pusher get the http.Pusher for server push
	Pusher() http.Pusher

	// WriteHeaderNow forces to write the http header (status code + headers).
	WriteHeaderNow()
}

type responseWriter struct {
	http.ResponseWriter
	//
	//
	size   int
	status int
}

func newResponseWriter(origin http.ResponseWriter) ResponseWriter {
	return &responseWriter{
		ResponseWriter: origin,
		size:           noWritten,
		status:         defaultStatus, // default status 404
	}
}

func (w *responseWriter) WriteHeader(code int) {
	if code > 0 && w.status != code {
		if w.Written() {
			logger.Debugf("[WARNING] Headers were already written. Wanted to override status code %d with %d", w.status, code)
			return
		}

		w.status = code
	}
}

func (w *responseWriter) Write(b []byte) (n int, err error) {
	// write status
	w.WriteHeaderNow()

	// write body
	n, err = w.ResponseWriter.Write(b)

	// record size
	w.size += n
	return
}

///////////////////////

// WriteHeaderNow forces to write the http header (status code + headers).
func (w *responseWriter) WriteHeaderNow() {
	if !w.Written() {
		w.size = 0
		w.ResponseWriter.WriteHeader(w.status)
	}
}

func (w *responseWriter) reset(writer http.ResponseWriter) {
	w.ResponseWriter = writer
	w.size = noWritten
	w.status = defaultStatus
}

///////////////////////

func (w *responseWriter) WriteString(s string) (n int, err error) {
	w.Write([]byte(s))
	return
}

///////////////////////

func (w *responseWriter) Status() int {
	return w.status
}

func (w *responseWriter) Size() int {
	return w.size
}

func (w *responseWriter) Written() bool {
	return w.size != noWritten
}

// Hijack implements the http.Hijacker interface.
func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if w.size < 0 {
		w.size = 0
	}

	return w.ResponseWriter.(http.Hijacker).Hijack()
}

func (w *responseWriter) CloseNotify() <-chan bool {
	return w.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

// Flush implements the http.Flusher interface.
func (w *responseWriter) Flush() {
	w.WriteHeaderNow()

	w.ResponseWriter.(http.Flusher).Flush()
}

func (w *responseWriter) Pusher() (pusher http.Pusher) {
	if pusher, ok := w.ResponseWriter.(http.Pusher); ok {
		return pusher
	}
	return nil
}
