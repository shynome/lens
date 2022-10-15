package lens

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/donovanhide/eventsource"
	"github.com/labstack/echo/v4"
	"github.com/lainio/err2/try"
	"github.com/shynome/lens/sdk"
)

var serviceInit = &sync.Once{}
var l = try.To1(net.Listen("tcp", ":0"))

func initLens() {
	serviceInit.Do(func() {
		e := echo.New()
		scopes := NewUserScopes()
		scopes.WithEcho(e.Group(""))

		go http.Serve(l, e)
	})
}

func TestServer(t *testing.T) {
	initLens()

	addr := l.Addr().String()
	endpoint := fmt.Sprintf("http://user@%s?t=webrtc", addr)
	// endpoint = strings.Replace(endpoint, "[::]", "127.0.0.1", 1)
	sdk := sdk.New(endpoint)

	runHandleCallService(sdk)
	time.Sleep(2 * time.Second)

	var b = make([]byte, 17)
	rand.Read(b)
	output := try.To1(sdk.Call("webrtc", b))
	if !bytes.Equal(output, b) {
		t.Error(string(b), string(output))
	}
}

func TestWorker(t *testing.T) {
	initLens()
	addr := l.Addr().String()
	endpoint := fmt.Sprintf("http://user:pass@%s/?t=webrtc", addr)
	sdk := sdk.New(endpoint)
	runHandleCallService(sdk)
}

func runHandleCallService(client *sdk.Sdk) {
	endpoint := try.To1(sdk.WithTopic(client.Endpoint, ""))
	stream := try.To1(eventsource.Subscribe(endpoint, ""))
	go func() {
		for ev := range stream.Events {
			go func(ev eventsource.Event) {
				if err := client.Dial("", ev.Id(), []byte(ev.Data())); err != nil {
					return
				}
			}(ev)
		}
	}()
}

func TestScopeKeepAlive(t *testing.T) {
	e := echo.New()
	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest(http.MethodGet, "http://aaa:bbb@127.0.0.1/", nil)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	scopes := NewUserScopes()
	scopes.ScopeCheckInterval = time.Second
	scopes.ScopeSurvivalTime = 7 * time.Second

	scope := try.To1(scopes.AutoCreate(UserAuth{Username: "aaa", Password: "bbb"}))
	c.Set(UserScopeCtxKey, scope)
	go SubTasks(c)

	time.Sleep(5 * time.Second)
	if time.Since(scope.LastAliveAt) > time.Second {
		t.Error("last alive should less than 1s.")
	}

	cancel()
	time.Sleep(5 * time.Second)
	if scopes.Get("aaa") == nil {
		t.Error("scope should be alive")
	}

	time.Sleep(5 * time.Second)
	if scopes.Get("aaa") != nil {
		t.Error("scope should be deleted")
	}
}
