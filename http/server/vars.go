package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/juju/errors"

	"github.com/dearcode/crab/log"
	"github.com/dearcode/crab/meta"
	"github.com/dearcode/crab/validation"
)

// UnmarshalForm 解析form中或者url中参数, 只支持int和string.
func UnmarshalForm(req *http.Request, postion VariablePostion, result interface{}) error {
	if postion == FORM {
		if err := req.ParseForm(); err != nil {
			return errors.Trace(err)
		}
	}

	rt := reflect.TypeOf(result)
	rv := reflect.ValueOf(result)

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
		key := f.Tag.Get("json")
		if key == "" {
			key = f.Name
		}
		var val string

		switch postion {
		case FORM, URI:
			val = req.FormValue(key)
		case HEADER:
			val = req.Header.Get(key)
		}

		switch f.Type.Kind() {
		case reflect.Bool:
			if val != "" {
				rv.Field(i).SetBool(true)
			}
		case reflect.Int, reflect.Int64:
			vi, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				//不需要验证的key就不返回错误了
				if f.Tag.Get("valid") == "" {
					break
				}
				return fmt.Errorf("key:%v value:%v format error", key, val)
			}
			rv.Field(i).SetInt(vi)
		case reflect.Uint, reflect.Uint64:
			vi, err := strconv.ParseUint(val, 10, 64)
			if err != nil {
				//不需要验证的key就不返回错误了
				if f.Tag.Get("valid") == "" {
					break
				}
				return fmt.Errorf("key:%v value:%v format error", key, val)
			}
			rv.Field(i).SetUint(vi)

		case reflect.String:
			rv.Field(i).SetString(val)
		}
	}
	return nil
}

//UnmarshalJSON 解析body中的json数据.
func UnmarshalJSON(req *http.Request, result interface{}) error {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return errors.Trace(err)
	}
	//空body不解析，不报错
	if len(body) == 0 {
		return nil
	}

	return json.Unmarshal(body, result)
}

// UnmarshalBody 解析body中的json, form数据.
func UnmarshalBody(req *http.Request, result interface{}) error {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return errors.Trace(err)
	}

	//空body不解析，不报错
	if len(body) == 0 {
		return nil
	}

	if err := json.Unmarshal(body, result); err != nil {
		//如果指定类型为json的，解析出错要抛出错误信息, 但大多人使用时不指定content-type
		if ct := req.Header.Get("Content-Type"); strings.Contains(strings.ToLower(ct), "json") {
			return errors.Trace(err)
		}
	}

	values, _ := url.ParseQuery(string(body))

	return reflectStruct(func(key string) (string, bool) {
		vals, exist := values[key]
		if !exist {
			return "", false
		}
		return vals[0], true
	}, result)
}

// UnmarshalValidate 解析并检证参数.
func UnmarshalValidate(req *http.Request, postion VariablePostion, result interface{}) error {
	var err error
	if result == nil {
		return meta.ErrArgIsNil
	}

	if postion == JSON {
		err = UnmarshalJSON(req, result)
	} else {
		err = UnmarshalForm(req, postion, result)
	}

	if err != nil {
		return errors.Trace(err)
	}

	log.Debugf("request %s vars:%#v", postion, result)
	valid := validation.Validation{}
	_, err = valid.Valid(result)
	return errors.Trace(err)
}

//ParseHeaderVars 解析并验证头中参数.
func ParseHeaderVars(req *http.Request, result interface{}) error {
	return UnmarshalValidate(req, HEADER, result)
}

//ParseFormVars 解析并验证Form表单中参数.
func ParseFormVars(req *http.Request, result interface{}) error {
	return UnmarshalValidate(req, FORM, result)
}

//ParseJSONVars 解析并验证Body中的Json参数.
func ParseJSONVars(req *http.Request, result interface{}) error {
	return UnmarshalValidate(req, JSON, result)
}

//ParseVars 通用解析，先解析url,再解析body,最后验证结果
func ParseVars(req *http.Request, result interface{}) error {
	if result == nil {
		return meta.ErrArgIsNil
	}

	if err := ParseURLVars(req, result); err != nil {
		return errors.Trace(err)
	}

	if err := UnmarshalBody(req, result); err != nil {
		return errors.Trace(err)
	}

	valid := validation.Validation{}
	_, err := valid.Valid(result)
	return errors.Trace(err)
}

//ParseURLVars 解析url中参数.
func ParseURLVars(req *http.Request, result interface{}) error {
	values := req.URL.Query()
	return reflectStruct(func(key string) (string, bool) {
		vals, exist := values[key]
		if !exist {
			return "", false
		}
		return vals[0], true
	}, result)
}
