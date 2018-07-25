package config

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"

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

func configSplit(raw, sep string) []string {
	if i := strings.Index(raw, sep); i > 0 {
		return []string{strings.TrimSpace(raw[:i]), strings.TrimSpace(raw[i+1:])}
	}
	return []string{raw}
}

//NewConfig 加载配置文件.
func NewConfig(path, body string) (c *Config, err error) {
	if path != "" {
		dat, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}
		body = string(dat)
	}

	c = &Config{kv: make(map[string]string)}
	s := ""

	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if len(line) < 3 || line[0] == ';' || line[0] == '#' {
			continue
		}

		if line[0] == '[' && line[len(line)-1] == ']' {
			s = line[1 : len(line)-1]
			continue
		}

		kv := configSplit(line, "=")
		if len(kv) == 2 {
			key := makeKey(s, kv[0])
			c.kv[key] = kv[1]
		}
	}

	return c, nil
}

func makeKey(s, k string) string {
	return strings.ToLower(strings.TrimSpace(s) + "/" + strings.TrimSpace(k))
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
			return errors.Annotatef(ErrNotFound, "%v->%v", s, k)
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

//ParseConfig 解析内存中配置文件.
func ParseConfig(body string, result interface{}) error {
	c, err := NewConfig("", body)
	if err != nil {
		return errors.Trace(err)
	}
	return c.Parse(result)
}

//LoadConfig 加载文件形式配置文件, 并解析成指定结构.
func LoadConfig(path string, result interface{}) error {
	c, err := NewConfig(path, "")
	if err != nil {
		return errors.Trace(err)
	}
	return c.Parse(result)
}

func (c Config) getDefault(k reflect.Kind, v string) (interface{}, error) {
	switch k {
	case reflect.Uint:
		d, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return nil, errors.Trace(err)
		}
		return uint(d), nil
	case reflect.Uint8:
		d, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return nil, errors.Trace(err)
		}
		return uint8(d), nil
	case reflect.Uint16:
		d, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return nil, errors.Trace(err)
		}
		return uint16(d), nil
	case reflect.Uint32:
		d, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return nil, errors.Trace(err)
		}
		return uint32(d), nil
	case reflect.Uint64:
		return strconv.ParseUint(v, 10, 64)
	case reflect.Int:
		d, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, errors.Trace(err)
		}
		return int(d), nil
	case reflect.Int8:
		d, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, errors.Trace(err)
		}
		return uint8(d), nil
	case reflect.Int16:
		d, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, errors.Trace(err)
		}
		return uint16(d), nil
	case reflect.Int32:
		d, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, errors.Trace(err)
		}
		return uint32(d), nil
	case reflect.Int64:
		return strconv.ParseInt(v, 10, 64)
	case reflect.String:
		return v, nil
	case reflect.Bool:
		return strconv.ParseBool(v)
	case reflect.Float32:
		d, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return nil, errors.Trace(err)
		}
		return float32(d), nil
	case reflect.Float64:
		return strconv.ParseFloat(v, 32)

	default:
		return nil, Errunsupported
	}
}

//Parse 根据result结构读配置文件.
func (c *Config) Parse(result interface{}) error {
	rt := reflect.TypeOf(result)
	rv := reflect.ValueOf(result)

	if rt.Kind() != reflect.Ptr {
		return errors.Trace(ErrInvalidType)
	}

	//去指针
	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
		rv = rv.Elem()
	}

	//只有两层
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if f.PkgPath != "" && !f.Anonymous { // unexported
			continue
		}
		ft := rt.Field(i).Type
		fv := rv.Field(i)
		if f.Type.Kind() == reflect.Ptr {
			fv = fv.Elem()
			ft = ft.Elem()
		}

		if ft.Kind() != reflect.Struct {
			continue
		}

		segment := f.Name
		if name := f.Tag.Get("cfg_key"); name != "" {
			segment = name
		}

		for j := 0; j < ft.NumField(); j++ {
			sf := ft.Field(j)
			if f.PkgPath != "" && !f.Anonymous { // unexported
				continue
			}
			sfv := fv.Field(j)
			if sf.Type.Kind() == reflect.Ptr {
				sfv = reflect.New(sfv.Elem().Type())
				fv.Field(j).Set(sfv)
			}

			var cfgDefault interface{}
			if v := sf.Tag.Get("cfg_default"); v != "" {
				d, err := c.getDefault(sf.Type.Kind(), v)
				if err != nil {
					return errors.Trace(err)
				}
				cfgDefault = d
			}

			key := sf.Name
			if name := sf.Tag.Get("cfg_key"); name != "" {
				key = name
			}

			if err := c.GetData(segment, key, sfv.Addr().Interface(), cfgDefault); err != nil {
				return errors.Trace(err)
			}
		}
	}

	return nil
}
