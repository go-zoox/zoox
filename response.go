package zoox

import (
	"bufio"
	"io"
	"net"
	"net/http"
)

// ResponseWriter ...
type ResponseWriter interface {
	http.ResponseWriter
	http.Hijacker
	http.CloseNotifier
	http.Flusher

	setContext(ctx *Context)

	// Status returns the HTTP response status code of the current request.
	Status() int

	// Size returns the number of bytes already written into the response http body.
	// See Written()
	Size() int

	// WriteString writes the string into the response body.
	WriteString(string) (int, error)

	// Written returns true if the response body was already written.
	Written() bool

	// WriteHeaderNow forces to write the http header (status code + headers).
	WriteHeaderNow()

	// Pusher get the http.Pusher for server push
	Pusher() http.Pusher
}

type responseWriter struct {
	http.ResponseWriter
	size   int
	status int
	ctx    *Context
	//
	isStatusWritten bool
	//
	isEmpty bool
}

func newResponseWriter(origin http.ResponseWriter) ResponseWriter {
	return &responseWriter{
		ResponseWriter: origin,
		size:           -1,
		status:         404, // default status 404
		isEmpty:        true,
	}
}

func (w *responseWriter) Status() int {
	return w.status
}

func (w *responseWriter) Size() int {
	return w.size
}

func (w *responseWriter) Written() bool {
	return w.size != -1
}

func (w *responseWriter) WriteHeader(code int) {
	if !w.isStatusWritten {
		w.isStatusWritten = true
	}

	if code > 0 && w.status != code {
		w.status = code

		w.ctx.StatusCode = code
	}
}

func (w *responseWriter) WriteHeaderNow() {
	if !w.Written() {
		// @TODO io.Copy response write will not trigger writeHeader
		if !w.isEmpty && !w.isStatusWritten {
			w.status = 200
			w.ctx.StatusCode = 200
		}

		w.size = 0
		w.ResponseWriter.WriteHeader(w.status)
	}
}

func (w *responseWriter) Write(b []byte) (n int, err error) {
	if w.isEmpty {
		w.isEmpty = len(b) == 0
	}

	w.WriteHeaderNow()
	n, err = w.ResponseWriter.Write(b)
	w.size += n
	return
}

func (w *responseWriter) WriteString(s string) (n int, err error) {
	if w.isEmpty {
		w.isEmpty = len(s) == 0
	}

	w.WriteHeaderNow()
	n, err = io.WriteString(w.ResponseWriter, s)
	w.size += n
	return
}

// Hijack implements the http.Hijacker interface.
func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if w.size < 0 {
		w.size = 0
	}

	return w.ResponseWriter.(http.Hijacker).Hijack()
}

func (w *responseWriter) setContext(ctx *Context) {
	w.ctx = ctx
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
