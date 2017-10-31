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

type uriRegexp struct {
	path   string
	keys   []string
	exp    *regexp.Regexp
	source reflect.Type
}

// iface 对外服务接口, path格式：Method/URI
type iface struct {
	path   string
	source reflect.Type
}

type prefix struct {
	path    string
	uriExps []uriRegexp
}

type httpServer struct {
	path     map[string]iface
	prefix   *btree.BTree
	filter   Filter
	listener net.Listener
	sync.RWMutex
}

func newHTTPServer() *httpServer {
	return &httpServer{
		path:   make(map[string]iface),
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

func newURIRegexp(path string, source reflect.Type) uriRegexp {
	ur := uriRegexp{
		path:   path,
		source: source,
	}

	//log.Debugf("path:%v", path)
	for _, m := range keysExp.FindAllStringSubmatch(path, -1) {
		//log.Debugf("path:%v, key:%#v", path, m)
		ur.keys = append(ur.keys, m[1])
	}

	np := keysExp.ReplaceAllString(path, "(.+)")
	exp, err := regexp.Compile(np)
	if err != nil {
		panic(err.Error())
	}
	ur.exp = exp

	//	log.Debugf("uriRegexp:%#v", ur)

	return ur
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
		case POST.String():
		case GET.String():
		case PUT.String():
		case DELETE.String():
		default:
			log.Warnf("ignore %v %v %v", method, path, rt)
			continue
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

			exp := newURIRegexp(path, rt.Elem())

			p.uriExps = append(p.uriExps, exp)

			server.prefix.ReplaceOrInsert(&p)

			log.Infof("add prefix %v %v %v %v", method, exp.path, exp.keys, rt)

			continue
		}

		ifc := iface{
			path:   fmt.Sprintf("%v%v", method, path),
			source: rt.Elem(),
		}
		//全路径匹配
		if _, ok := server.path[ifc.path]; ok {
			panic(fmt.Sprintf("exist url:%v %v", method, path))
		}

		server.path[ifc.path] = ifc
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

func parseRequestValues(path string, ur uriRegexp) context.Context {
	ctx := context.Background()
	for i, v := range ur.exp.FindAllStringSubmatch(path, -1) {
		//log.Debugf("i:%v, v:%#v, keys:%#v", i, v, ur.keys)
		ctx = context.WithValue(ctx, ur.keys[i], v[1])
	}
	return ctx
}

func getInterface(method, path string) (reflect.Type, context.Context) {
	server.RLock()
	defer server.RUnlock()

	path = method + path

	if i, ok := server.path[path]; ok {
		//log.Debugf("find path:%v", path)
		return i.source, nil
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

	for _, ue := range p.uriExps {
		if ue.exp.MatchString(path) {
			//	log.Debugf("uri exp:%v, path:%v", ue, path)
			return ue.source, parseRequestValues(path, ue)
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

	tp, ctx := getInterface(r.Method, r.URL.Path)
	if tp == nil {
		log.Errorf("%v %v %v not found.", r.RemoteAddr, r.Method, r.URL)
		SendResponse(w, http.StatusNotFound, "invalid request")
		return
	}

	if ctx != nil {
		nr = nr.WithContext(ctx)
	}

	log.Debugf("%v %v %v %v", r.RemoteAddr, r.Method, r.URL, tp)

	callback := reflect.New(tp).MethodByName(r.Method).Interface().(func(http.ResponseWriter, *http.Request))
	callback(w, nr)

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

//Stop 停止httpServer监听, 进行中的任务并不会因此而停止.
func Stop() error {
	server.Lock()
	defer server.Unlock()

	return server.listener.Close()
}
