package config

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/juju/errors"
)

var (
	Errunsupported = fmt.Errorf("expect ptr data")
	ErrInvalidType = fmt.Errorf("result must be ptr")
	ErrNotFound    = fmt.Errorf("key not found")
)

// Config 读ini配置文件.
type Config struct {
	kv map[string]string
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

//NewConfig 加载配置文件.
func NewConfig(path string) (c *Config, err error) {
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	c = &Config{kv: make(map[string]string)}
	s := ""

	for _, line := range TrimSplit(string(dat), "\n") {
		line = TrimSpace(line)
		if len(line) < 3 || line[0] == ';' || line[0] == '#' {
			continue
		}

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
	return strings.ToLower(fmt.Sprintf("%s_%s", s, k))
}

//GetData 获取指定段的指定key的值, 支持int,string.
func (c *Config) GetData(s, k string, result interface{}, d interface{}) error {
	rt := reflect.TypeOf(result)
	if rt.Kind() != reflect.Ptr {
		return errors.Trace(ErrInvalidType)
	}
	rt = rt.Elem()
	rv := reflect.ValueOf(result).Elem()

	key := makeKey(s, k)

	v, ok := c.kv[key]
	if !ok {
		//没有对应的key, 这时候要看看有没有default.
		if d == nil {
			return errors.Trace(ErrNotFound)
		}
		rv.Set(reflect.ValueOf(d))
		return nil
	}

	switch rt.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		data, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return errors.Trace(err)
		}
		rv.SetUint(data)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		data, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return errors.Trace(err)
		}
		rv.SetInt(data)

	case reflect.String:
		rv.SetString(v)

	case reflect.Bool:
		data, err := strconv.ParseBool(v)
		if err != nil {
			return errors.Trace(err)
		}
		rv.SetBool(data)
	case reflect.Float32, reflect.Float64:
		data, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return errors.Trace(err)
		}
		rv.SetFloat(data)
	default:
		return errors.Trace(Errunsupported)
	}

	return nil
}

func LoadConfig(path string, result interface{}) error {
	c, err := NewConfig(path)
	if err != nil {
		return errors.Trace(err)
	}

	rt := reflect.TypeOf(result)
	rv := reflect.ValueOf(result)

	if rt.Kind() != reflect.Ptr {
		return errors.Trace(ErrInvalidType)
	}

	//去指针
	if rt.Kind() == reflect.Ptr && rt.Elem().Kind() == reflect.Struct {
		rt = rt.Elem()
		rv = rv.Elem()
	}

	//只有两层
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if f.PkgPath != "" && !f.Anonymous { // unexported
			continue
		}
		fv := rv.Field(i)

		if f.Type.Kind() == reflect.Struct {
			for j := 0; j < f.Type.NumField(); j++ {
				sf := f.Type.Field(j)
				if f.PkgPath != "" && !f.Anonymous { // unexported
					continue
				}
				sfv := fv.Field(j)

				var d interface{}
				if v := sf.Tag.Get("default"); v != "" {
					switch sf.Type.Kind() {
					case reflect.Uint:
						d, err = strconv.ParseUint(v, 10, 64)
						d = uint(d.(uint64))
					case reflect.Uint8:
						d, err = strconv.ParseUint(v, 10, 64)
						d = uint8(d.(uint64))
					case reflect.Uint16:
						d, err = strconv.ParseUint(v, 10, 64)
						d = uint16(d.(uint64))
					case reflect.Uint32:
						d, err = strconv.ParseUint(v, 10, 64)
						d = uint32(d.(uint64))
					case reflect.Uint64:
						d, err = strconv.ParseUint(v, 10, 64)
					case reflect.Int:
						d, err = strconv.ParseInt(v, 10, 64)
						d = int(d.(int64))
					case reflect.Int8:
						d, err = strconv.ParseInt(v, 10, 64)
						d = uint8(d.(uint64))
					case reflect.Int16:
						d, err = strconv.ParseInt(v, 10, 64)
						d = uint16(d.(uint64))
					case reflect.Int32:
						d, err = strconv.ParseInt(v, 10, 64)
						d = uint32(d.(uint64))
					case reflect.Int64:
						d, err = strconv.ParseInt(v, 10, 64)
					case reflect.String:
						d = v
					case reflect.Bool:
						d, err = strconv.ParseBool(v)
					case reflect.Float32, reflect.Float64:
						d, err = strconv.ParseFloat(v, 32)

					default:
						return Errunsupported
					}
				}
				if err != nil {
					return errors.Trace(err)
				}

				if err = c.GetData(f.Name, sf.Name, sfv.Addr().Interface(), d); err != nil {
					return errors.Trace(err)
				}
			}
		}
	}

	return nil
}
