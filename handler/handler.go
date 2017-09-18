package handler

import (
	"fmt"
	"net/http"
	"reflect"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/google/btree"
	"github.com/zssky/log"
)

var (
	//Server 默认服务
	Server = newHTTPServer()
)

type server struct {
	path   map[string]iface
	prefix *btree.BTree
	filter Filter
	mu     sync.RWMutex
}

func newHTTPServer() *server {
	return &server{
		path:   make(map[string]iface),
		prefix: btree.New(3),
		filter: defaultFilter,
	}
}

// Filter 请求过滤， 如果返回结果为nil,直接返回，不再进行后续处理.
type Filter func(http.ResponseWriter, *http.Request) *http.Request

// iface 对外服务接口, path格式：Method/URI
type iface struct {
	path   string
	source reflect.Type
}

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

//AddInterface 自动注册接口
//只要struct实现了Get(),Post(),Delete(),Put()接口就可以自动注册
func (s *server) AddInterface(obj interface{}, path string, isPrefix bool) error {
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

	s.mu.Lock()
	defer s.mu.Unlock()

	rv := reflect.ValueOf(obj)
	for i := 0; i < rv.NumMethod(); i++ {
		method := rt.Method(i).Name
        log.Debugf("rt:%v, %d, method:%v", rt, i, method)
		switch method {
		case POST.String():
		case GET.String():
		case PUT.String():
		case DELETE.String():
		default:
			log.Debugf("ignore method:%v path:%v", method, path)
			continue
		}

		ifc := iface{
			path:   fmt.Sprintf("%v%v", method, path),
			source: rt.Elem(),
		}

		//前缀匹配
		if isPrefix {
			if s.prefix.Has(&ifc) {
				panic(fmt.Sprintf("exist url:%v %v", method, path))
			}
			s.prefix.ReplaceOrInsert(&ifc)
            log.Debugf("add prefix:%v", path)
			continue
		}

		//全路径匹配
		if _, ok := s.path[ifc.path]; ok {
			panic(fmt.Sprintf("exist url:%v %v", method, path))
		}

		s.path[ifc.path] = ifc
	}

	return nil
}

//AddFilter 添加过滤函数.
func (s *server) AddFilter(filter Filter) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.filter = filter
}

func (s *server) iface(w http.ResponseWriter, r *http.Request) (i iface, ok bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := r.Method + r.URL.Path

	if i, ok = s.path[path]; ok {
		//log.Debugf("find path:%v", path)
		return
	}

	//如果完全匹配没找到，再找前缀的
	s.prefix.AscendGreaterOrEqual(&iface{path: path}, func(item btree.Item) bool {
		i = *(item.(*iface))
		ok = strings.HasPrefix(path, i.path)
       // log.Debugf("path:%v, ipath:%v, ok:%v", path, i.path, ok)
		return !ok
	})

//	log.Debugf("find prefix:%v, ok:%v", path, ok)
	return
}

//ServeHTTP 真正对外服务接口
func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	i, ok := s.iface(w, r)
	if !ok {
		log.Errorf("%v %v %v not found.", r.RemoteAddr, r.Method, r.URL)
		SendResponse(w, http.StatusNotFound, "invalid request")
		return
	}

	log.Debugf("%v %v %v %v", r.RemoteAddr, r.Method, r.URL, i.path)

    callback := reflect.New(i.source).MethodByName(r.Method).Interface().(func(http.ResponseWriter, *http.Request))
	callback(w, nr)

	return
}

func defaultFilter(_ http.ResponseWriter, r *http.Request) *http.Request {
	return r
}
