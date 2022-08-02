package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"dearcode.net/crab/http/client"
	"dearcode.net/crab/http/server"
)

type regexpTest struct {
}

func (rt *regexpTest) GET(w http.ResponseWriter, req *http.Request) {
	user, ok := server.RESTValue(req, "user")
	if !ok {
		panic("context `user` not found")
	}
	id, ok := server.RESTValue(req, "id")
	if !ok {
		panic("context `user` not found")
	}

	_, ok = server.RESTValue(req, "idx")
	if ok {
		panic("find invalid key idx")
	}

	fmt.Fprintf(w, "user:%v, id:%v", user, id)
}

type index struct {
	r *http.Request
}

func (i *index) GET(w http.ResponseWriter, req *http.Request) {
	i.r = req
	w.Write([]byte(fmt.Sprintf("client:%v addr:%p", i.r.RemoteAddr, i)))
}

func testRESTClient() {
	url := "http://127.0.0.1:9000/regexp/mailchina/test/"
	//url := fmt.Sprintf("http://127.0.0.1:9000/regexp/mailchina/test/u1%v", time.Now().UnixNano())
	buf, err := client.New().Get(url, nil, nil)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("response:%s\n", buf)
}

func testHTTPClient() {
	url := fmt.Sprintf("http://127.0.0.1:9000/main/index/?id=111&tm=%v", time.Now().UnixNano())
	buf, err := client.New().Get(url, nil, nil)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("response:%s\n", buf)
}

func testHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "testHandler:%v", req.RemoteAddr)
}

func main() {
	addr := flag.String("h", ":9000", "api listen address")
	flag.Parse()

	server.Register(&index{})
	server.RegisterPrefix(&regexpTest{}, "/regexp/{user}/test/{id}")

	server.RegisterHandler(testHandler, "GET", "/testHandler/")

	ln, err := server.Start(*addr)
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("listen:%v\n", ln.Addr())

	for i := 0; i < 5; i++ {
		time.Sleep(time.Second)
		testHTTPClient()
		testRESTClient()
	}

}
