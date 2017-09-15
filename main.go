package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/dearcode/crab/handler"
	_ "github.com/dearcode/crab/server"
)

type index struct {
}

func (i *index) GET(w http.ResponseWriter, req *http.Request) {
	fmt.Printf("index:%p\n", i)
	w.Write([]byte("ok"))
    time.Sleep(time.Minute)

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
