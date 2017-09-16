package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/dearcode/crab/http/client"
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

func testHTTPClient() {
	url := "http://127.0.0.1:9000/index"
	buf, _, err := client.New(time.Second).Get(url, nil, nil)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("response:%s\n", buf)
}

func main() {
	addr := flag.String("h", ":9000", "api listen address")
	flag.Parse()

	server.AddInterface(&index{}, "/index", false)

	go func() {
		for i := 0; i < 5; i++ {
			time.Sleep(time.Second)
			testHTTPClient()
		}
	}()

	if err := server.Start(*addr); err != nil {
		panic(err.Error())
	}
}
