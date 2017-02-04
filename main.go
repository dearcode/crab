package main

import (
    "flag"
    "net"
    "net/http"

    "github.com/dearcode/webgo/handler"
    _ "github.com/dearcode/webgo/server"
)

func main() {
    addr := flag.String("h", ":9000", "api listen address")
    flag.Parse()

    ln, err := net.Listen("tcp", *addr)
    if err != nil {
        panic(err.Error())
    }

    if err = http.Serve(ln, handler.Server); err != nil {
        panic(err.Error())
    }

}
