package server

import (
	"github.com/dearcode/crab/http/server"
)

func init() {
	server.RegisterPrefix(&staticServer{}, "/")
	server.RegisterPrefix(&debugServer{}, "/debug/")

	server.RegisterPrefix(&testServer{}, "/test/")

	server.Register(&user{})
}
