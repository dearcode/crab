package config

import (
	"fmt"
	"io/ioutil"
	"reflect"
	//"strconv"
	"strings"
	"sync"
	"unicode"
)

var (
	ErrExpectPtr = fmt.Errorf("expect ptr data")
)

// Config 读ini配置文件.
type Config struct {
	kv map[string]string
	mu sync.RWMutex
}

//TrimSpace 删除头尾的空字符，空格，换行之类的东西, 具体在unicode.IsSpac.
func TrimSpace(raw string) string {
	s := strings.TrimLeftFunc(raw, unicode.IsSpace)
	if s != "" {
		s = strings.TrimRightFunc(s, unicode.IsSpace)
	}
	return s
}

// TrimSplit 按sep拆分，并去掉空字符.
func TrimSplit(raw, sep string) []string {
	var ss []string

	s := TrimSpace(raw)
	if s == "" {
		return ss
	}

	ss = strings.Split(s, sep)
	i := 0
	for _, s := range ss {
		s = TrimSpace(s)
		if s != "" {
			ss[i] = s
			i++
		}
	}

	return ss[:i]
}

//trim 清理注释, 空格.
func trim(line string) string {
	b := '\n'
	for i, v := range line {
		switch v {
		case '/':
			if b == '/' {
				return line[:i]
			}
		case '#':
			return line[:i]
		}
		b = v
	}
	return TrimSpace(line)
}

//NewConfig 加载配置文件.
func NewConfig(f string) (c *Config, err error) {
	dat, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}

	c = &Config{kv: make(map[string]string)}
	s := ""

	for _, line := range TrimSplit(string(dat), "\n") {
		line = trim(line)
		if line[0] == '[' && line[len(line)-1] == ']' {
			s = line[1 : len(line)-1]
			continue
		}

		kv := TrimSplit(line, "=")
		if len(kv) == 2 {
			key := makeKey(s, kv[0])
			c.kv[key] = kv[1]
		}
	}

	return c, nil
}

func makeKey(s, k string) string {
	return fmt.Sprintf("%s_%s", s, k)
}

//GetData 获取指定段的指定key的值, 支持int,string.
func (c *Config) GetData(s, k string, result interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	rt := reflect.TypeOf(result)
	if rt.Kind() != reflect.Ptr {
		return ErrExpectPtr
	}
	rt = rt.Elem()
	rv := reflect.ValueOf(result).Elem()

	key := makeKey(s, k)

	v, ok := c.kv[key]
	if !ok {
		return nil
	}

	//var data interface{}
    var err error
	switch rt.Kind() {
	case reflect.Int:
		//data, err = strconv.Atoi(v)
    case reflect.String:
	rv.SetString(v)
//        data = &v
	}

//	rv.Set(reflect.ValueOf(data))

	return err
}
