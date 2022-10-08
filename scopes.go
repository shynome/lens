package lens

import (
	"sync"
	"time"

	"github.com/lainio/err2/try"
)

type UserScopes struct {
	mu                *sync.RWMutex
	Scopes            map[string]*UserScope
	DisableAutoCreate bool

	ScopeSurvivalTime  time.Duration
	ScopeCheckInterval time.Duration
}

func NewUserScopes() *UserScopes {
	return &UserScopes{
		mu:                &sync.RWMutex{},
		Scopes:            map[string]*UserScope{},
		DisableAutoCreate: false,

		ScopeSurvivalTime:  5 * time.Minute,
		ScopeCheckInterval: time.Second,
	}
}

func (scopes *UserScopes) AutoCreate(auth UserAuth) (s *UserScope, err error) {
	if scopes.DisableAutoCreate {
		return nil, ErrPasswordIncorrect // 使用密码错误, 避免用户探测
	}
	s, err = scopes.Create(auth)
	if err != nil {
		return
	}
	go scopes.DeleteAfter(s, scopes.ScopeSurvivalTime)
	return
}

func (scopes *UserScopes) Create(auth UserAuth) (s *UserScope, err error) {
	scopes.mu.Lock()
	defer scopes.mu.Unlock()

	s = try.To1(NewUserScope(auth))

	user := s.Auth.Username
	scopes.Scopes[user] = s

	return
}

// Get if not exists will return nil
func (scopes *UserScopes) Get(user string) (s *UserScope) {
	scopes.mu.Lock()
	defer scopes.mu.Unlock()
	s = scopes.Scopes[user]
	return
}

func (scopes *UserScopes) DeleteAfter(scope *UserScope, d time.Duration) {
	for {
		time.Sleep(scopes.ScopeCheckInterval)
		now := time.Now()
		expiresAt := scope.LastAliveAt.Add(d)
		if expiresAt.Before(now) {
			scopes.Delete(scope.Auth.Username)
			break
		}
	}
}

func (scopes *UserScopes) Delete(user string) {
	scopes.mu.Lock()
	defer scopes.mu.Unlock()
	_, ok := scopes.Scopes[user]
	if !ok { //deleted
		return
	}
	delete(scopes.Scopes, user)
}
