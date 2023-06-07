package sse

import (
	"fmt"
	"net/http"
	"time"
)

// SSE ...
type SSE interface {
	Retry(delat time.Duration)
	Event(name, data string)
	Comment(comment string)
}

// sse ...
type sse struct {
	id      int
	writer  http.ResponseWriter
	flusher http.Flusher
}

func New(rw http.ResponseWriter) SSE {
	rw.WriteHeader(200)
	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("Access-Control-Allow-Origin", "*")

	return &sse{
		writer:  rw,
		flusher: rw.(http.Flusher),
	}
}

func (s *sse) Retry(delay time.Duration) {
	s.writer.Write([]byte(fmt.Sprintf("retry: %d\n", delay/time.Millisecond)))
	s.flusher.Flush()
}

func (s *sse) Event(name string, data string) {
	s.id++

	s.writer.Write([]byte("event: " + name + "\n"))
	s.writer.Write([]byte(fmt.Sprintf("id: %d\n", s.id)))
	s.writer.Write([]byte("data: " + data + "\n\n"))

	s.flusher.Flush()
}

func (s *sse) Comment(comment string) {
	s.writer.Write([]byte(": " + comment + "\n"))
	s.flusher.Flush()
}
