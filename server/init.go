package server

import (
	"github.com/dearcode/crab/http/server"
)

func init() {
	server.AddInterface(&staticServer{}, "/", true)
	server.AddInterface(&debugServer{}, "/debug/", true)

	server.AddInterface(&testServer{}, "/test/", false)

	server.AddInterface(&user{}, "", false)
}
