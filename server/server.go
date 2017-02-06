package server

import (
	"fmt"
	"net/http"
	"net/http/pprof"
)

//onTestGet 获取配置文件中域名
func onTestGet(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Get test"))
}

//onTestPost 获取配置文件中域名
func onTestPost(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Post test"))
}

//onTestDelete 获取配置文件中域名
func onTestDelete(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Delete test"))
}

//onStaticGet 静态文件
func onStaticGet(w http.ResponseWriter, r *http.Request) {
	path := fmt.Sprintf("/var/www/html/%s", r.URL.Path)
	w.Header().Add("Cache-control", "no-store")
	http.ServeFile(w, r, path)
}

//onDebugGet debug接口
func onDebugGet(w http.ResponseWriter, r *http.Request) {
	pprof.Index(w, r)
}
