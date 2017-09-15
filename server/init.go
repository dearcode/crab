package server

import (
	"github.com/dearcode/crab/handler"
)

func init() {
	handler.Server.AddInterface(&staticServer{}, "/", true)
	handler.Server.AddInterface(&debugServer{}, "/debug/", true)

	handler.Server.AddInterface(&testServer{}, "/test/", false)

	handler.Server.AddInterface(&user{}, "", false)
}
