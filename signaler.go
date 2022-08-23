package signaler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
	"github.com/rs/xid"
	"github.com/shynome/signaler/sdk"
)

type Signaler struct {
	scopes *UserScopes

	CallTimeout time.Duration
}

var _ http.Handler = &Signaler{}

func New() (s *Signaler) {
	return &Signaler{
		scopes: NewUserScopes(),

		CallTimeout: time.Minute,
	}
}

func (s *Signaler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := s.serveHTTP(w, r)
	if err != nil {
		log.Println("err: ", err)
		http.Error(w, "server error", 500)
	}
}

var mb = int64(math.Pow(2, 10))
var maxFileSize = 4 * mb

func (s *Signaler) serveHTTP(w http.ResponseWriter, r *http.Request) (err error) {
	defer err2.Return(&err)

	r.Body = http.MaxBytesReader(w, r.Body, maxFileSize)

	user, pass, _ := r.BasicAuth()
	auth := UserAuth{user, pass}
	scope := try.To1(
		s.scopes.Get(auth))
	ctx := context.WithValue(r.Context(), UserScopeCtxKey, scope)
	r = r.WithContext(ctx)

	switch r.Method {
	case http.MethodGet:
		return s.SubTasks(w, r)
	case http.MethodPost:
		return s.HandleCall(w, r)
	case http.MethodDelete:
		return s.DialTask(w, r)
	default:
		http.Error(w, fmt.Sprintf("deny method: %s \r\n", r.Method), 400)
	}
	return
}

func (s *Signaler) SubTasks(w http.ResponseWriter, r *http.Request) (err error) {
	defer err2.Return(&err)

	scope := try.To1(
		GetUserScope(r.Context()))

	ess := scope.EventSourceServer()

	topic := r.URL.Query().Get("t")
	if topic == "" {
		return fmt.Errorf("topic is not selected")
	}
	ess.Handler(topic)(w, r)
	return
}

func (s *Signaler) HandleCall(w http.ResponseWriter, r *http.Request) (err error) {
	defer err2.Return(&err)

	scope := try.To1(
		GetUserScope(r.Context()))
	evs := scope.Events()

	id := xid.New().String()
	result := evs.On(id)
	defer evs.Off(id)

	b := try.To1(
		io.ReadAll(r.Body))

	ev := &sdk.Task{ID: id, Body: b}
	topic := r.URL.Query().Get("t")
	if topic == "" {
		return fmt.Errorf("topic is not selected")
	}

	ess := scope.EventSourceServer()
	ess.Publish([]string{topic}, ev)

	select {
	case <-time.After(s.CallTimeout):
		http.Error(w, "too long to wait task dail", http.StatusGatewayTimeout)

	case rbody := <-result:
		h := w.Header()
		h.Set("Content-Type", "application/octet-stream")
		try.To1(
			io.Copy(w, bytes.NewReader(rbody)))
	}

	return
}

func (s *Signaler) DialTask(w http.ResponseWriter, r *http.Request) (err error) {
	defer err2.Return(&err)

	scope := try.To1(
		GetUserScope(r.Context()))

	id := r.Header.Get("X-Event-Id")
	b := try.To1(
		io.ReadAll(r.Body))

	try.To(
		scope.Events().Emit(id, b))

	return
}
