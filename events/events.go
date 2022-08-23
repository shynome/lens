package events

import (
	"fmt"
	"sync"
)

type Events[T any] struct {
	mu     *sync.RWMutex
	events map[string]chan T
}

func New[T any]() *Events[T] {
	return &Events[T]{
		mu:     &sync.RWMutex{},
		events: map[string]chan T{},
	}
}

func (s *Events[T]) Emit(id string, data T) (err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ch, ok := s.events[id]
	if !ok {
		return fmt.Errorf("event handler [%s] is empty", id)
	}
	ch <- data
	return
}

func (s *Events[T]) On(id string) chan T {
	s.mu.Lock()
	s.events[id] = make(chan T)
	s.mu.Unlock()
	return s.events[id]
}

func (s *Events[T]) Off(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.events, id)
}
