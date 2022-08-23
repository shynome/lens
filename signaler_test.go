package signaler_test

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/donovanhide/eventsource"
	"github.com/lainio/err2/try"
	"github.com/shynome/signaler"
	"github.com/shynome/signaler/sdk"
)

func TestUrlAuth(t *testing.T) {
	u, err := http.NewRequest(http.MethodGet, "http://a:b@127.0.0.1:3030?t=7&t=8", nil)
	if err != nil {
		t.Error(err)
		return
	}
	user := u.URL.User.Username()
	if user != "a" {
		t.Errorf("expect a, got %s", user)
		return
	}
	topics := u.URL.Query()["t"]
	fmt.Println(topics)
}

var testData = []byte("hello world")

func TestSignaler(t *testing.T) {
	var l = try.To1(
		net.Listen("tcp", "127.0.0.1:0"))
	defer l.Close()
	s := signaler.New(nil)
	s.CallTimeout = 10 * time.Second
	endpoint := fmt.Sprintf("http://a:b@%s/signaler?t=7", l.Addr())
	http.Handle("/signaler", s)
	go http.Serve(l, nil)
	handleTask(endpoint)

	sdk := sdk.New(endpoint)
	rData := try.To1(sdk.Call(testData))

	if !bytes.Equal(rData, testData) {
		t.Errorf("got %s, expect %s", rData, testData)
		return
	}
}

func handleTask(endpoint string) {
	sdk := sdk.New(endpoint)
	s := try.To1(
		eventsource.Subscribe(endpoint, ""))
	go func() {
		for ev := range s.Events {
			go func(ev eventsource.Event) {
				id := ev.Id()
				sdk.Dial(id, []byte(ev.Data()))
			}(ev)
		}
	}()
}
