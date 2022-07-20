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

	Status() int

	Size() int

	// WriteString writes the string into the response body.
	WriteString(string) (int, error)

	Written() bool

	setContext(ctx *Context)
}

type responseWriter struct {
	http.ResponseWriter
	size   int
	status int
	ctx    *Context
}

func newResponseWriter(origin http.ResponseWriter) ResponseWriter {
	return &responseWriter{
		ResponseWriter: origin,
		size:           -1,
		status:         404, // default status 404
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
	if code > 0 && w.status != code {
		w.status = code

		w.ctx.StatusCode = code
	}
}

func (w *responseWriter) writeHeaderNow(isEmpty bool) {
	if !w.Written() {
		// @TODO io.Copy response write will not trigger writeHeader
		if !isEmpty && w.status == 404 {
			w.status = 200
			w.ctx.StatusCode = 200
		}

		w.size = 0
		w.ResponseWriter.WriteHeader(w.status)
	}
}

func (w *responseWriter) Write(b []byte) (n int, err error) {
	w.writeHeaderNow(len(b) == 0)
	n, err = w.ResponseWriter.Write(b)
	w.size += n
	return
}

func (w *responseWriter) WriteString(s string) (n int, err error) {
	w.writeHeaderNow(len(s) == 0)
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
	w.writeHeaderNow(true)

	w.ResponseWriter.(http.Flusher).Flush()
}
