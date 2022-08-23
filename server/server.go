package main

import (
	"flag"
	"net/http"
	"strings"
	"time"

	"github.com/lainio/err2/try"
	"github.com/shynome/signaler"
)

var args struct {
	Addr               string
	AutoCreate         bool
	ScopeCheckInterval string
	ScopeSurvivalTime  string
	TaskTimeout        string
}

func init() {
	flag.StringVar(&args.Addr, "addr", ":7070", "listen addr")
	flag.BoolVar(&args.AutoCreate, "auto-create", true, "allow auto create user scope")
	flag.StringVar(&args.ScopeCheckInterval, "check-interval", "1s", "check user scope is should be delete interval")
	flag.StringVar(&args.ScopeSurvivalTime, "survival-time", "5m", "auto create user scope survival time")
	flag.StringVar(&args.TaskTimeout, "task-timeout", "10s", "dail task timeout")
}

func main() {
	flag.Parse()

	scopes := signaler.NewUserScopes()
	scopes.DisableAutoCreate = args.AutoCreate
	scopes.ScopeCheckInterval = try.To1(
		time.ParseDuration(args.ScopeCheckInterval))
	scopes.ScopeSurvivalTime = try.To1(
		time.ParseDuration(args.ScopeSurvivalTime))

	s := signaler.New(scopes)
	s.CallTimeout = try.To1(
		time.ParseDuration(args.TaskTimeout))

	for _, userStr := range flag.Args() {
		userArr := strings.Split(userStr, ":")
		if len(userArr) == 1 {
			continue
		}
		pass := strings.Join(userArr[1:], ":")
		user := userArr[0]
		try.To1(scopes.Create(signaler.UserAuth{Username: user, Password: pass}))
	}

	http.ListenAndServe(args.Addr, s)
}
