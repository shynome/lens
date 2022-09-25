package signaler

import (
	"bytes"
	"context"
	"fmt"
	"io"
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
	MaxFileSize int64
}

var _ http.Handler = &Signaler{}

var mb = int64(math.Pow(2, 10))

func New(scopes *UserScopes) (s *Signaler) {
	if scopes == nil {
		scopes = NewUserScopes()
	}
	return &Signaler{
		scopes: scopes,

		CallTimeout: time.Minute,
		MaxFileSize: 4 * mb,
	}
}

func (s *Signaler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := s.serveHTTP(w, r)
	if err != nil {
		http.Error(w, fmt.Sprintf("err: %s", err), 500)
	}
}

func (s *Signaler) serveHTTP(w http.ResponseWriter, r *http.Request) (err error) {
	defer err2.Return(&err)

	r.Body = http.MaxBytesReader(w, r.Body, s.MaxFileSize)

	h := w.Header()
	h.Set("Access-Control-Allow-Origin", "*")
	h.Set("Access-Control-Allow-Methods", "*")

	user, pass, _ := r.BasicAuth()
	auth := UserAuth{user, pass, r.URL.Query().Get("w")}
	scope := try.To1(
		s.scopes.Get(auth))
	ctx := context.WithValue(r.Context(), UserScopeCtxKey, scope)
	r = r.WithContext(ctx)

	switch r.Method {
	case http.MethodGet:
		if scope == nil {
			r = try.To1(
				s.CreateScope(r, auth))
		}
		return s.SubTasks(w, r)
	case http.MethodPost:
		return s.HandleCall(w, r)
	case http.MethodDelete:
		return s.DialTask(w, r)
	default:
		http.Error(w, fmt.Sprintf("deny method: %s \r\n", r.Method), http.StatusMethodNotAllowed)
	}
	return
}

func (s *Signaler) CreateScope(r *http.Request, auth UserAuth) (r2 *http.Request, err error) {
	defer err2.Return(&err)

	scopes := s.scopes

	scope := try.To1(
		scopes.AutoCreate(auth))

	ctx := context.WithValue(r.Context(), UserScopeCtxKey, scope)
	r2 = r.WithContext(ctx)

	return
}

var (
	errTopicRequired = fmt.Errorf("topic is not selected")
	errTaskTimeout   = fmt.Errorf("too long to wait task dail")
)

func (s *Signaler) SubTasks(w http.ResponseWriter, r *http.Request) (err error) {
	defer err2.Return(&err)

	q := r.URL.Query()

	scope := try.To1(
		GetUserScope(r.Context()))
	try.To(
		scope.CheckWorkerToken(q.Get("w")))

	ess := scope.EventSourceServer()

	topic := q.Get("t")
	if topic == "" {
		return errTopicRequired
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
		return errTopicRequired
	}

	ess := scope.EventSourceServer()
	ess.Publish([]string{topic}, ev)

	select {
	case <-time.After(s.CallTimeout):
		http.Error(w, errTaskTimeout.Error(), http.StatusGatewayTimeout)

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

	q := r.URL.Query()

	scope := try.To1(
		GetUserScope(r.Context()))
	try.To(
		scope.CheckWorkerToken(q.Get("w")))

	id := r.Header.Get("X-Event-Id")
	b := try.To1(
		io.ReadAll(r.Body))

	try.To(
		scope.Events().Emit(id, b))

	return
}
