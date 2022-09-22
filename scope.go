package signaler

import (
	"context"
	"fmt"
	"time"

	"github.com/donovanhide/eventsource"
	"github.com/shynome/signaler/events"
)

type UserAuth struct {
	Username string
	Password string
	WToken   string
}

type UserScope struct {
	Auth UserAuth

	ess *eventsource.Server
	evs *events.Events[[]byte]

	LastAliveAt time.Time
}

var UserScopeCtxKey struct{}

var (
	ErrPasswordIncorrect    = fmt.Errorf("password is incorrect")
	ErrWorkerTokenIncorrect = fmt.Errorf("worker token is incorrect")
	ErrUserScopeNotReady    = fmt.Errorf("get *UserScope failed")
)

func GetUserScope(ctx context.Context) (s *UserScope, err error) {
	s, ok := ctx.Value(UserScopeCtxKey).(*UserScope)
	if !ok {
		return nil, ErrUserScopeNotReady
	}
	return
}

func NewUserScope(auth UserAuth) (s *UserScope, err error) {

	if auth.Username == "" {
		return nil, errUsernameRequired
	}

	s = &UserScope{
		Auth: auth,

		ess: eventsource.NewServer(),
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

func (scope *UserScope) CheckWorkerToken(wt string) (err error) {
	if scope.Auth.Password == "" {
		return
	}
	if scope.Auth.WToken != wt {
		return ErrWorkerTokenIncorrect
	}
	return
}

func (scope *UserScope) KeepAlive() {
	scope.LastAliveAt = time.Now()
}

func (scope *UserScope) Events() *events.Events[[]byte] {
	return scope.evs
}

func (scope *UserScope) EventSourceServer() *eventsource.Server {
	return scope.ess
}
