package lens

import (
	"github.com/labstack/echo/v4"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

var (
	ErrTopicRequired = echo.NewHTTPError(400, "topic is not selected")
	ErrUserRequired  = echo.NewHTTPError(400, "user is required")
	ErrTaskTimeout   = echo.NewHTTPError(500, "too long to wait task dail")
)

func TopicRequired(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		topic := c.QueryParam("t")
		if topic == "" {
			return ErrTopicRequired
		}
		return next(c)
	}
}

func UserRequired(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, _, _ := c.Request().BasicAuth()
		if user == "" {
			return ErrUserRequired
		}
		return next(c)
	}
}

func ScopeRequired(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		_, ok := c.Get(UserScopeCtxKey).(*UserScope)
		if !ok {
			return ErrUserScopeNotReady
		}
		return next(c)
	}
}

func VerifyScopePassword(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		defer err2.Return(&err)
		scope := c.Get(UserScopeCtxKey).(*UserScope)
		_, pass, _ := c.Request().BasicAuth()
		try.To(scope.CheckPassword(pass))
		return next(c)
	}
}
