package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/lainio/err2/try"
	"github.com/shynome/signaler"
)

var Version = "dev"

var args struct {
	PrintVersion bool

	Addr               string
	AutoCreate         bool
	ScopeCheckInterval string
	ScopeSurvivalTime  string
	TaskTimeout        string
}

func init() {
	flag.BoolVar(&args.PrintVersion, "version", false, "print version")

	flag.StringVar(&args.Addr, "addr", ":7070", "listen addr")
	flag.BoolVar(&args.AutoCreate, "auto-create", true, "allow auto create user scope")
	flag.StringVar(&args.ScopeCheckInterval, "check-interval", "1s", "check user scope is should be delete interval")
	flag.StringVar(&args.ScopeSurvivalTime, "survival-time", "5m", "auto create user scope survival time")
	flag.StringVar(&args.TaskTimeout, "task-timeout", "10s", "dail task timeout")
}

func main() {
	flag.Parse()

	if args.PrintVersion {
		fmt.Println("version:", Version)
		return
	}

	scopes := signaler.NewUserScopes()
	scopes.DisableAutoCreate = !args.AutoCreate
	scopes.ScopeCheckInterval = try.To1(
		time.ParseDuration(args.ScopeCheckInterval))
	scopes.ScopeSurvivalTime = try.To1(
		time.ParseDuration(args.ScopeSurvivalTime))

	s := signaler.New(scopes)
	s.CallTimeout = try.To1(
		time.ParseDuration(args.TaskTimeout))

	for _, userStr := range flag.Args() {
		userArr := strings.Split(userStr, ":")
		var (
			user string
			pass string
			wt   string
		)
		switch len(userArr) {
		case 1:
			fallthrough
		default:
			continue
		case 2:
			user = userArr[0]
			pass = userArr[1]
		case 3:
			user = userArr[0]
			wt = userArr[1]
			pass = strings.Join(userArr[2:], ":")
		}
		try.To1(scopes.Create(signaler.UserAuth{Username: user, Password: pass, WToken: wt}))
	}

	http.ListenAndServe(args.Addr, s)
}
