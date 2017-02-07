package handler

import (
	"fmt"
	"net/http"
	"reflect"
	"runtime/debug"
	"strings"

	"github.com/davygeek/log"
)

var (
	//Server 默认服务
	Server = newHTTPServer()
)

func newHTTPServer() *server {
	return &server{
		post:   make(map[string]iface),
		get:    make(map[string]iface),
		put:    make(map[string]iface),
		delete: make(map[string]iface),
		prefix: make(map[string]iface),
	}
}

// Callback 用户接口
type Callback func(http.ResponseWriter, *http.Request)

// iface 对外服务接口
type iface struct {
	method Method
	path   string
	call   Callback
}

type server struct {
	post   map[string]iface
	get    map[string]iface
	put    map[string]iface
	delete map[string]iface

	prefix map[string]iface
}

//nameToPath 类名转路径
func (s *server) nameToPath(name string) string {
	buf := []byte(name)
	for i := range buf {
		if buf[i] == '.' || buf[i] == '*' {
			buf[i] = '/'
		}
	}
	buf = append(buf, '/')
	return string(buf)
}

//AddInterface 自动注册接口
//只要struct实现了DoGet(),DoPost(),DoDelete(),DoPut()接口就可以自动注册
func (s *server) AddInterface(iface interface{}) error {
	rt := reflect.TypeOf(iface)
	if rt.Kind() != reflect.Ptr {
		return fmt.Errorf("need ptr")
	}
	rv := reflect.ValueOf(iface)
	for i := 0; i < rv.NumMethod(); i++ {
		mt := rt.Method(i)
		mv := rv.Method(i)
		path := s.nameToPath(rt.String())

		switch mt.Name {
		case "DoPost":
			s.AddHandler(POST, path, false, mv.Interface().(func(http.ResponseWriter, *http.Request)))
		case "DoGet":
			s.AddHandler(GET, path, false, mv.Interface().(func(http.ResponseWriter, *http.Request)))
		case "DoPut":
			s.AddHandler(PUT, path, false, mv.Interface().(func(http.ResponseWriter, *http.Request)))
		case "DoDelete":
			s.AddHandler(DELETE, path, false, mv.Interface().(func(http.ResponseWriter, *http.Request)))
		}
		log.Debugf("%v %v", mt.Name, path)
	}

	return nil
}

//AddHandler 注册接口
func (s *server) AddHandler(method Method, path string, isPrefix bool, call Callback) {
	var ms map[string]iface
	switch method {
	case GET:
		ms = s.get
	case POST:
		ms = s.post
	case PUT:
		ms = s.put
	case DELETE:
		ms = s.delete
	}
	if isPrefix {
		ms = s.prefix
	}

	if _, ok := ms[path]; ok {
		panic(fmt.Sprintf("exist url:%v %v", method, path))
	}

	ms[path] = iface{method, path, call}
}

//ServeHTTP 真正对外服务接口
func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if p := recover(); p != nil {
			log.Errorf("panic:%v req:%v, stack:%s", p, r, debug.Stack())
			SendResponse(w, http.StatusInternalServerError, "%v", p)
			return
		}
	}()
	var i iface
	var ok bool
	log.Debugf("%v %v", r.Method, r.URL)

	switch r.Method {
	case "GET":
		i, ok = s.get[r.URL.Path]
	case "POST":
		i, ok = s.post[r.URL.Path]
	case "PUT":
		i, ok = s.put[r.URL.Path]
	case "DELETE":
		i, ok = s.delete[r.URL.Path]
	default:
		log.Errorf("invalid request req:%v", r)
		SendResponse(w, http.StatusBadRequest, "invalid method:%v", r.Method)
		return
	}

	//如果完全匹配没找到，再找前缀的
	if !ok {
		for k, v := range s.prefix {
			if strings.HasPrefix(r.URL.Path, k) {
				i = v
				ok = true
				break
			}
		}
	}

	if !ok {
		log.Errorf("handler not found, req:%v", r)
		SendResponse(w, http.StatusBadRequest, "invalid request:%v", r)
		return
	}

	i.call(w, r)
	return
}
