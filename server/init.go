package server

import (
	"github.com/dearcode/petrel/handler"
)

func init() {
	handler.Server.AddHandler(handler.GET, "/", true, onStaticGet)
	handler.Server.AddHandler(handler.GET, "/debug/", true, onDebugGet)

	handler.Server.AddHandler(handler.GET, "/test/", false, onTestGet)
	handler.Server.AddHandler(handler.POST, "/test/", false, onTestPost)
	handler.Server.AddHandler(handler.DELETE, "/test/", false, onTestDelete)

	handler.Server.AddInterface(&user{}, "")
}
