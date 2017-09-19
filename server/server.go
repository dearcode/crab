package server

import (
	"fmt"
	"net/http"
	"net/http/pprof"
)

type testServer struct {
}

func (s *testServer) GET(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Get test"))
}

func (s *testServer) POST(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Post test"))
}

func (s *testServer) PUT(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Put test"))
}

func (s *testServer) DELETE(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Delete test"))
}

type staticServer struct {
}

func (s *staticServer) GET(w http.ResponseWriter, r *http.Request) {
	path := fmt.Sprintf("/var/www/html/%s", r.URL.Path)
	w.Header().Add("Cache-control", "no-store")
	http.ServeFile(w, r, path)
}

type debugServer struct {
}

func (s *debugServer) GET(w http.ResponseWriter, r *http.Request) {
	pprof.Index(w, r)
}
