package lens

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/shynome/lens/events"
	"github.com/shynome/lens/sse"
)

type UserAuth struct {
	Username string
	Password string
}

type UserScope struct {
	Auth UserAuth

	ess *sse.Server
	evs *events.Events[[]byte]

	LastAliveAt time.Time
}

var (
	ErrPasswordIncorrect = echo.NewHTTPError(400, "password is incorrect")
	ErrUserScopeNotReady = echo.NewHTTPError(404, "get *UserScope failed")
)

func NewUserScope(auth UserAuth) (s *UserScope, err error) {

	if auth.Username == "" {
		return nil, ErrUserRequired
	}

	ess := sse.New()

	s = &UserScope{
		Auth: auth,

		ess: ess,
		evs: events.New[[]byte](),

		LastAliveAt: time.Now(),
	}

	return
}

func (scope *UserScope) CheckPassword(pass string) (err error) {
	if scope.Auth.Password == "" {
		return
	}
	if scope.Auth.Password != pass {
		return ErrPasswordIncorrect
	}
	return
}

func (scope *UserScope) KeepAlive() {
	scope.LastAliveAt = time.Now()
}

func (scope *UserScope) Events() *events.Events[[]byte] {
	return scope.evs
}

func (scope *UserScope) EventSourceServer() *sse.Server {
	return scope.ess
}
