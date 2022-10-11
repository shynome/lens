package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/lainio/err2/try"
	"github.com/shynome/lens"
)

var Version = "dev"

var args struct {
	PrintVersion bool

	Addr               string
	AutoCreate         bool
	ScopeCheckInterval string
	ScopeSurvivalTime  string
}

func init() {
	flag.BoolVar(&args.PrintVersion, "version", false, "print version")

	flag.StringVar(&args.Addr, "addr", ":7070", "listen addr")
	flag.BoolVar(&args.AutoCreate, "auto-create", true, "allow auto create user scope")
	flag.StringVar(&args.ScopeCheckInterval, "check-interval", "1s", "check user scope is should be delete interval")
	flag.StringVar(&args.ScopeSurvivalTime, "survival-time", "5m", "auto create user scope survival time")
}

func main() {
	flag.Parse()

	if args.PrintVersion {
		fmt.Println("version:", Version)
		return
	}

	scopes := lens.NewUserScopes()
	scopes.DisableAutoCreate = !args.AutoCreate
	scopes.ScopeCheckInterval = try.To1(
		time.ParseDuration(args.ScopeCheckInterval))
	scopes.ScopeSurvivalTime = try.To1(
		time.ParseDuration(args.ScopeSurvivalTime))

	for _, userStr := range flag.Args() {
		userArr := strings.Split(userStr, ":")
		if len(userArr) == 1 {
			continue
		}
		var (
			user string
			pass string
		)
		user = userArr[0]
		pass = strings.Join(userArr[1:], ":")
		try.To1(scopes.Create(lens.UserAuth{Username: user, Password: pass}))
	}

	e := echo.New()
	e.Use(middleware.BodyLimit("2M"))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		Skipper:      middleware.DefaultSkipper,
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"*"},
		AllowHeaders: []string{"*"},
	}))
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: 15 * time.Second,
		Skipper: func(c echo.Context) bool {
			r := c.Request()
			// skip sse subscribe
			if r.Method == http.MethodGet && r.URL.Path == "/" {
				return true
			}
			return false
		},
	}))

	scopes.WithEcho(e.Group(""))

	e.HideBanner = true
	try.To(e.Start(args.Addr))
}
