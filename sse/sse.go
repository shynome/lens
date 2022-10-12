package sse

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
)

type Stream interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	Publish(ev Event)
}

type Server struct {
	conns  map[uint64]chan Event
	cursor *uint64
	mu     *sync.RWMutex
}

var _ Stream = &Server{}
var _ http.Handler = &Server{}

func New() (s *Server) {
	cursor := uint64(0)
	s = &Server{
		conns:  map[uint64]chan Event{},
		cursor: &cursor,
		mu:     &sync.RWMutex{},
	}
	return
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	events, id := s.Events()
	defer s.OffEventsReceiver(id)

	WriteHeader(w)

	done := r.Context().Done()
	for {
		select {
		case ev := <-events:
			FlushEvent(w, ev)
		case <-done:
			return
		}
	}
}

func WriteHeader(w http.ResponseWriter) {
	defer w.(http.Flusher).Flush()
	h := w.Header()
	h.Set("Content-Type", "text/event-stream; charset=utf-8")
	h.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	h.Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	io.WriteString(w, ": a hack comment for pass caddy\n\n")
}

func FlushEvent(w http.ResponseWriter, ev Event) {
	defer w.(http.Flusher).Flush()
	io.WriteString(w, ev.String())
}

func (s *Server) Events() (ch chan Event, id uint64) {
	id = atomic.AddUint64(s.cursor, 1)
	ch = make(chan Event)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.conns[id] = ch
	return
}
func (s *Server) OffEventsReceiver(id uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.conns, id)
}

func (s *Server) Publish(ev Event) {
	go s.dispatch(ev)
}

func (s *Server) dispatch(ev Event) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, ch := range s.conns {
		ch <- ev
	}
}

type Event struct {
	ID    string
	Data  string
	Event string
}

func (ev *Event) String() string {
	s := ""
	s += fmt.Sprintf("id: %s\n", ev.ID)
	s += fmt.Sprintf("data: %s\n", ev.Data)
	if len(ev.Event) > 0 {
		s += fmt.Sprintf("event: %s\n", ev.Event)
	}
	s += "\n"
	return s
}
