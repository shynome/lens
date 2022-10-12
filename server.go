package lens

import (
	"io"

	"github.com/labstack/echo/v4"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
	"github.com/rs/xid"
	"github.com/shynome/lens/sse"
)

const UserScopeCtxKey = "UserScopeCtxKey"

func (scopes *UserScopes) WithEcho(e *echo.Group) {
	e.Use(UserRequired, TopicRequired)
	e.Use(scopes.InjectScopeIfExists)

	e.GET("/", SubTasks, scopes.AutoCreateMiddleware, VerifyScopePassword)
	e.POST("/", HandleCall, ScopeRequired)
	e.DELETE("/", DialTask, ScopeRequired, VerifyScopePassword)
}

func SubTasks(c echo.Context) (err error) {
	defer err2.Return(&err)

	scope := c.Get(UserScopeCtxKey).(*UserScope)
	ess := scope.EventSourceServer()

	topic := c.QueryParam("t")

	w := c.Response()
	r := c.Request()

	events, id := ess.Events()
	defer ess.OffEventsReceiver(id)

	sse.WriteHeader(w)

	done := r.Context().Done()
	for {
		select {
		case ev := <-events:
			if ev.Event == topic {
				sse.FlushEvent(w, ev)
			}
		case <-done:
			return
		}
	}
}

func HandleCall(c echo.Context) (err error) {
	defer err2.Return(&err)

	scope := c.Get(UserScopeCtxKey).(*UserScope)
	evs := scope.Events()

	id := xid.New().String()
	result := evs.On(id)
	defer evs.Off(id)

	r := c.Request()
	b := try.To1(io.ReadAll(r.Body))
	topic := c.QueryParam("t")
	ev := sse.Event{
		ID:    id,
		Data:  string(b),
		Event: topic,
	}

	ess := scope.EventSourceServer()
	ess.Publish(ev)

	select {
	case rbody := <-result:
		c.Blob(200, "application/octet-stream", rbody)
	case <-r.Context().Done():
	}

	return
}

func DialTask(c echo.Context) (err error) {
	defer err2.Return(&err)

	scope := c.Get(UserScopeCtxKey).(*UserScope)
	r := c.Request()
	id := r.Header.Get("X-Event-Id")

	b := try.To1(io.ReadAll(r.Body))
	try.To(scope.Events().Emit(id, b))

	return
}
