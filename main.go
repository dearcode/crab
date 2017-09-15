package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"

	"github.com/dearcode/crab/http/server"
	_ "github.com/dearcode/crab/server"
)

type index struct {
	r *http.Request
}

func (i *index) GET(w http.ResponseWriter, req *http.Request) {
	i.r = req
	w.Write([]byte(fmt.Sprintf("client:%v addr:%p", i.r.RemoteAddr, i)))
}

func main() {
	addr := flag.String("h", ":9000", "api listen address")
	flag.Parse()

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		panic(err.Error())
	}

	server.AddInterface(&index{}, "/index", false)
	if err = server.Start(ln); err != nil {
		panic(err.Error())
	}
}
