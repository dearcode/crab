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

	"github.com/dearcode/crab/log"
)

//UserKey 用户自定义的key
type UserKey string

type handlerRegexp struct {
	keys []UserKey
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
	mu       sync.RWMutex
}

var (
	server  = newHTTPServer()
	keysExp *regexp.Regexp
)

func init() {
	exp, err := regexp.Compile(`{(\w+)?}`)
	if err != nil {
		panic(err.Error())
	}
	keysExp = exp
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

//RegisterMust 只要struct实现了Get(),Post(),Delete(),Put()接口就可以自动注册, 如果添加失败panic.
func RegisterMust(obj interface{}) {
	if err := register(obj, "", false); err != nil {
		panic(err.Error())
	}

}

//RegisterPath 注册url完全匹配.
func RegisterPath(obj interface{}, path string) error {
	return register(obj, path, false)
}

//RegisterPathMust 注册url完全匹配，如果遇到错误panic.
func RegisterPathMust(obj interface{}, path string) {
	if err := register(obj, path, false); err != nil {
		panic(err.Error())
	}
}

//RegisterHandler 注册自定义url完全匹配.
func RegisterHandler(call func(http.ResponseWriter, *http.Request), method, path string) error {
	h := handler{
		path: fmt.Sprintf("%v%v", method, path),
		call: call,
	}

	server.mu.Lock()
	defer server.mu.Unlock()

	if _, ok := server.path[h.path]; ok {
		return errors.Errorf("exist url:%v %v", method, path)
	}

	server.path[h.path] = h

	log.Infof("handler %v %v", method, path)

	return nil
}

//RegisterPrefix 注册url前缀.
func RegisterPrefix(obj interface{}, path string) error {
	return register(obj, path, true)
}

//RegisterPrefixMust 注册url前缀并保证成功.
func RegisterPrefixMust(obj interface{}, path string) {
	if err := RegisterPrefix(obj, path); err != nil {
		panic(err.Error())
	}
}

func newHandlerRegexp(h handler) handlerRegexp {
	hr := handlerRegexp{handler: h}

	for _, m := range keysExp.FindAllStringSubmatch(hr.path, -1) {
		hr.keys = append(hr.keys, UserKey(m[1]))
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

	server.mu.Lock()
	defer server.mu.Unlock()

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
			log.Warningf("ignore func:%v %v %v", method, path, rt)
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

			log.Infof("prefix %v %v %v %v", method, path, exp.keys, rt)
			continue
		}

		//全路径匹配
		if _, ok := server.path[h.path]; ok {
			return errors.Errorf("exist url:%v %v", method, path)
		}

		server.path[h.path] = h
		log.Infof("path %v %v %v", method, path, rt)
	}

	return nil
}

//AddFilter 添加过滤函数.
func AddFilter(filter Filter) {
	server.mu.Lock()
	defer server.mu.Unlock()
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
	server.mu.RLock()
	defer server.mu.RUnlock()

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
	//	log.Debugf("ok:%v, prefix:%v", ok, p)

	if !ok {
		return nil, nil
	}

	for _, ue := range p.exps {
		if ue.exp.MatchString(path) {
			if len(ue.keys) == 0 {
				return ue.call, nil
			}
			//			log.Debugf("uri exp:%v, path:%v", ue, path)
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
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if ctx != nil {
		nr = nr.WithContext(ctx)
	}

	log.Debugf("%v %v %v h:%p", r.RemoteAddr, r.Method, r.URL, h)

	h(w, nr)
}

func defaultFilter(_ http.ResponseWriter, r *http.Request) *http.Request {
	return r
}

//Start 启动httpServer.
func Start(addr string) (net.Listener, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, errors.Trace(err)
	}

	server.mu.Lock()
	server.listener = ln
	server.mu.Unlock()

	go func() {
		if err = http.Serve(ln, server); err != nil {
			log.Errorf("Serve error:%v", err)
		}
	}()
	return ln, nil
}
