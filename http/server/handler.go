package server

import (
	"fmt"
	"net"
	"net/http"
	"reflect"
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

// iface 对外服务接口, path格式：Method/URI
type iface struct {
	path   string
	source reflect.Type
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

func (i *iface) Less(bi btree.Item) bool {
	return strings.Compare(i.path, bi.(*iface).path) == 1
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

//RegisterPath 注册url完全匹配.
func RegisterPath(obj interface{}, path string) error {
	return register(obj, path, true)
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

		ifc := iface{
			path:   fmt.Sprintf("%v%v", method, path),
			source: rt.Elem(),
		}

		//前缀匹配
		if isPrefix {
			if server.prefix.Has(&ifc) {
				panic(fmt.Sprintf("exist url:%v %v", method, path))
			}
			server.prefix.ReplaceOrInsert(&ifc)
			log.Infof("add prefix %v %v %v", method, path, rt)
			continue
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

func getInterface(w http.ResponseWriter, r *http.Request) (i iface, ok bool) {
	server.RLock()
	defer server.RUnlock()

	path := r.Method + r.URL.Path

	if i, ok = server.path[path]; ok {
		//log.Debugf("find path:%v", path)
		return
	}

	//如果完全匹配没找到，再找前缀的
	server.prefix.AscendGreaterOrEqual(&iface{path: path}, func(item btree.Item) bool {
		i = *(item.(*iface))
		ok = strings.HasPrefix(path, i.path)
		// log.Debugf("path:%v, ipath:%v, ok:%v", path, i.path, ok)
		return !ok
	})

	//	log.Debugf("find prefix:%v, ok:%v", path, ok)
	return
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

	i, ok := getInterface(w, r)
	if !ok {
		log.Errorf("%v %v %v not found.", r.RemoteAddr, r.Method, r.URL)
		SendResponse(w, http.StatusNotFound, "invalid request")
		return
	}

	log.Debugf("%v %v %v %v", r.RemoteAddr, r.Method, r.URL, i.source)

	callback := reflect.New(i.source).MethodByName(r.Method).Interface().(func(http.ResponseWriter, *http.Request))
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
