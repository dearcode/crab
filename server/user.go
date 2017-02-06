package server

import (
	"net/http"
)

type user struct {
}

//DoGet 默认get方法
func (u *user) DoGet(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Get user"))
}

//DoPost 默认post方法
func (u *user) DoPost(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Post user"))
}
