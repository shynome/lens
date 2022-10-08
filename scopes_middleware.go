package lens

import (
	"github.com/labstack/echo/v4"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

func (scopes *UserScopes) InjectScopeIfExists(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, _, _ := c.Request().BasicAuth()
		s := scopes.Get(user)
		if s != nil {
			c.Set(UserScopeCtxKey, s)
		}
		return next(c)
	}
}

func (scopes *UserScopes) AutoCreateMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		defer err2.Return(&err)
		_, ok := c.Get(UserScopeCtxKey).(*UserScope)
		if ok {
			return next(c)
		}
		user, pass, _ := c.Request().BasicAuth()
		s := try.To1(scopes.AutoCreate(UserAuth{user, pass}))
		c.Set(UserScopeCtxKey, s)
		return next(c)
	}
}
