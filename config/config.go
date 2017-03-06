package config

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

var (
	Errunsupported = fmt.Errorf("expect ptr data")
	ErrInvalidType = fmt.Errorf("result must be ptr")
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
func NewConfig(f string) (c *Config, err error) {
	dat, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}

	c = &Config{kv: make(map[string]string)}
	s := ""

	for _, line := range TrimSplit(string(dat), "\n") {
		if e := strings.Index(line, "#"); e > 0 {
			line = line[:e]
		}
		line = TrimSpace(line)

		if len(line) > 2 && line[0] == '[' && line[len(line)-1] == ']' {
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
func (c *Config) GetData(s, k string, result interface{}) error {
	rt := reflect.TypeOf(result)
	if rt.Kind() != reflect.Ptr {
		return ErrInvalidType
	}
	rt = rt.Elem()
	rv := reflect.ValueOf(result).Elem()

	key := makeKey(s, k)

	v, ok := c.kv[key]
	if !ok {
		return nil
	}

	switch rt.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		data, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return err
		}
		rv.SetUint(data)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		data, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return err
		}
		rv.SetInt(data)

	case reflect.String:
		rv.SetString(v)

	case reflect.Bool:
		data, err := strconv.ParseBool(v)
		if err != nil {
			return err
		}
		rv.SetBool(data)
	case reflect.Float32, reflect.Float64:
		data, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return err
		}
		rv.SetFloat(data)
	default:
		return Errunsupported
	}

	return nil
}

func LoadConfig(f string, result interface{}) error {
	c, err := NewConfig(f)
	if err != nil {
		return err
	}

	rt := reflect.TypeOf(result)
	rv := reflect.ValueOf(result)

	if rt.Kind() != reflect.Ptr {
		return ErrInvalidType
	}

	//去指针
	if rt.Kind() == reflect.Ptr && rt.Elem().Kind() == reflect.Struct {
		rt = rt.Elem()
		rv = rv.Elem()
	}

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

				if err = c.GetData(f.Name, sf.Name, sfv.Addr().Interface()); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
