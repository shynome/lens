package sse

import (
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/donovanhide/eventsource"
	"github.com/lainio/err2/try"
)

func TestSSE(t *testing.T) {
	var s = New()
	l := try.To1(net.Listen("tcp", ":0"))
	go http.Serve(l, s)
	endpoint := fmt.Sprintf("http://%s/", l.Addr().String())

	eid := "xxxx"

	wait := make(chan struct{})
	stream := try.To1(eventsource.Subscribe(endpoint, ""))
	go func() {
		for ev := range stream.Events {
			go func(ev eventsource.Event) {
				if eid != ev.Id() {
					t.Error(ev.Id())
				} else {
					t.Log(ev.Id())
				}
				close(wait)
			}(ev)
		}
	}()

	time.Sleep(time.Second)
	s.Publish(Event{
		ID:    eid,
		Data:  "yyyy",
		Event: "zzzz",
	})
	<-wait
}
