package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"

	"github.com/dearcode/crab/handler"
	_ "github.com/dearcode/crab/server"
)

type index struct {
    r *http.Request
}

func test(r *http.Request) {
  fmt.Printf("%v\n",    r.RemoteAddr)
}

func (i *index) GET(w http.ResponseWriter, req *http.Request) {
	fmt.Printf("index:%p\n", i)
    test(req)
	w.Write([]byte("ok"))
}

func main() {
	addr := flag.String("h", ":9000", "api listen address")
	flag.Parse()

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		panic(err.Error())
	}

	handler.Server.AddInterface(&index{}, "/index", false)

	if err = http.Serve(ln, handler.Server); err != nil {
		panic(err.Error())
	}
}
