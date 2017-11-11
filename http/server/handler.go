package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/google/btree"
	"github.com/juju/errors"
	"github.com/zssky/log"
)

var (
	server = newHTTPServer()
)

type handlerRegexp struct {
	keys []string
	exp  *regexp.Regexp
	handler
}

// handler 对外服务接口, path格式：Method/URI
type handler struct {
	path string
	call func(http.ResponseWriter, *http.Request)
}

type prefix struct {
	path string
	exps []handlerRegexp
}

type httpServer struct {
	path     map[string]handler
	prefix   *btree.BTree
	filter   Filter
	listener net.Listener
	sync.RWMutex
}

func newHTTPServer() *httpServer {
	return &httpServer{
		path:   make(map[string]handler),
		prefix: btree.New(3),
		filter: defaultFilter,
	}
}

// Filter 请求过滤， 如果返回结果为nil,直接返回，不再进行后续处理.
type Filter func(http.ResponseWriter, *http.Request) *http.Request

func (p *prefix) Less(bi btree.Item) bool {
	return strings.Compare(p.path, bi.(*prefix).path) == 1
}

//NameToPath 类名转路径
func NameToPath(name string, depth int) string {
	buf := []byte(name)
	d := 0
	index := 0
	for i := range buf {
		if buf[i] == '.' || buf[i] == '*' {
			buf[i] = '/'
			d++
			if d == depth {
				index = i
			}
		}
	}
	return string(buf[index:])
}

//Register 只要struct实现了Get(),Post(),Delete(),Put()接口就可以自动注册
func Register(obj interface{}) error {
	return register(obj, "", false)
}

//RegisterPath 注册url完全匹配.
func RegisterPath(obj interface{}, path string) error {
	return register(obj, path, false)
}

//RegisterHandler 注册自定义url完全匹配.
func RegisterHandler(call func(http.ResponseWriter, *http.Request), method, path string) error {
	h := handler{
		path: fmt.Sprintf("%v%v", method, path),
		call: call,
	}

	server.Lock()
	defer server.Unlock()

	if _, ok := server.path[h.path]; ok {
		return errors.Errorf("exist url:%v %v", method, path)
	}

	server.path[h.path] = h

	log.Infof("add handler %v %v", method, path)

	return nil
}

//RegisterPrefix 注册url前缀.
func RegisterPrefix(obj interface{}, path string) error {
	return register(obj, path, true)
}

var (
	keysExp *regexp.Regexp
)

func init() {
	exp, err := regexp.Compile("{(\\w+)?}")
	if err != nil {
		panic(err.Error())
	}
	keysExp = exp
}

func newHandlerRegexp(h handler) handlerRegexp {
	hr := handlerRegexp{handler: h}

	for _, m := range keysExp.FindAllStringSubmatch(hr.path, -1) {
		hr.keys = append(hr.keys, m[1])
	}

	np := keysExp.ReplaceAllString(hr.path, "(.+)")
	exp, err := regexp.Compile(np)
	if err != nil {
		panic(err.Error())
	}

	hr.exp = exp

	return hr
}

func register(obj interface{}, path string, isPrefix bool) error {
	rt := reflect.TypeOf(obj)
	if rt.Kind() != reflect.Ptr {
		return fmt.Errorf("need ptr")
	}

	if path == "" {
		path = NameToPath(rt.String(), 0) + "/"
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	server.Lock()
	defer server.Unlock()

	rv := reflect.ValueOf(obj)
	for i := 0; i < rv.NumMethod(); i++ {
		method := rt.Method(i).Name
		//log.Debugf("rt:%v, %d, method:%v", rt, i, method)
		switch method {
		case http.MethodPost:
		case http.MethodGet:
		case http.MethodPut:
		case http.MethodDelete:
		default:
			log.Warnf("ignore func:%v %v %v", method, path, rt)
			continue
		}

		mt := rv.MethodByName(method)
		if mt.Type().NumIn() != 2 || mt.Type().In(0).String() != "http.ResponseWriter" || mt.Type().In(1).String() != "*http.Request" {
			log.Debugf("ignore func:%v %v %v", method, path, mt.Type())
			continue
		}

		h := handler{
			path: fmt.Sprintf("%v%v", method, path),
			call: mt.Interface().(func(http.ResponseWriter, *http.Request)),
		}

		//前缀匹配
		if isPrefix {
			p := prefix{
				path: fmt.Sprintf("%v%v", method, path),
			}

			if idx := strings.Index(path, "{"); idx > 0 {
				p.path = fmt.Sprintf("%v%v", method, path[:idx])
			}

			//	log.Debugf("path:%v", p.path)
			if server.prefix.Has(&p) {
				p = *(server.prefix.Get(&p).(*prefix))
			}

			exp := newHandlerRegexp(h)

			p.exps = append(p.exps, exp)

			server.prefix.ReplaceOrInsert(&p)

			log.Infof("add prefix %v %v %v %v", method, exp.path, exp.keys, rt)
			continue
		}

		//全路径匹配
		if _, ok := server.path[h.path]; ok {
			return errors.Errorf("exist url:%v %v", method, path)
		}

		server.path[h.path] = h
		log.Infof("add path %v %v %v", method, path, rt)
	}

	return nil
}

//AddFilter 添加过滤函数.
func AddFilter(filter Filter) {
	server.Lock()
	defer server.Unlock()
	server.filter = filter
}

func parseRequestValues(path string, ur handlerRegexp) context.Context {
	ctx := context.Background()
	for i, v := range ur.exp.FindAllStringSubmatch(path, -1) {
		//log.Debugf("i:%v, v:%#v, keys:%#v", i, v, ur.keys)
		ctx = context.WithValue(ctx, ur.keys[i], v[1])
	}
	return ctx
}

func getHandler(method, path string) (func(http.ResponseWriter, *http.Request), context.Context) {
	server.RLock()
	defer server.RUnlock()

	path = method + path

	if i, ok := server.path[path]; ok {
		//log.Debugf("find path:%v", path)
		return i.call, nil
	}

	var p prefix
	var ok bool
	//如果完全匹配没找到，再找前缀的
	server.prefix.AscendGreaterOrEqual(&prefix{path: path}, func(item btree.Item) bool {
		p = *(item.(*prefix))
		ok = strings.HasPrefix(path, p.path)
		//log.Debugf("path:%v, prefix:%v, ok:%v", path, p.path, ok)
		return !ok
	})

	if !ok {
		return nil, nil
	}

	for _, ue := range p.exps {
		if ue.exp.MatchString(path) {
			if len(ue.keys) == 0 {
				return ue.call, nil
			}
			//	log.Debugf("uri exp:%v, path:%v", ue, path)
			return ue.call, parseRequestValues(path, ue)
		}
	}

	return nil, nil
}

//ServeHTTP 真正对外服务接口
func (s *httpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if p := recover(); p != nil {
			log.Errorf("panic:%v req:%v, stack:%s", p, r, debug.Stack())
			Abort(w, "%v\n%s", p, debug.Stack())
			return
		}
	}()

	nr := s.filter(w, r)
	if nr == nil {
		log.Debugf("%v %v %v ignore", r.RemoteAddr, r.Method, r.URL)
		return
	}

	h, ctx := getHandler(r.Method, r.URL.Path)
	if h == nil {
		log.Errorf("%v %v %v not found.", r.RemoteAddr, r.Method, r.URL)
		SendResponse(w, http.StatusNotFound, "invalid request")
		return
	}

	if ctx != nil {
		nr = nr.WithContext(ctx)
	}

	log.Debugf("%v %v %v %v", r.RemoteAddr, r.Method, r.URL, h)

	h(w, nr)

	return
}

func defaultFilter(_ http.ResponseWriter, r *http.Request) *http.Request {
	return r
}

//Start 启动httpServer.
func Start(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return errors.Trace(err)
	}

	server.Lock()
	server.listener = ln
	server.Unlock()

	return http.Serve(ln, server)
}

//Listener 获取监听地址.
func Listener() net.Listener {
	server.Lock()
	defer server.Unlock()
	return server.listener
}

//Stop 停止httpServer监听, 进行中的任务并不会因此而停止.
func Stop() error {
	server.Lock()
	defer server.Unlock()

	return server.listener.Close()
}
