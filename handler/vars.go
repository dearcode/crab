package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"

	"github.com/davygeek/log"
	"github.com/juju/errors"

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

		log.Debugf("field:%v, val:%v", f, val)
		switch f.Type.Kind() {
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
		case reflect.String:
			rv.Field(i).SetString(val)
		}
	}
	return nil
}

// UnmarshalJSON 解析body中的json数据.
func UnmarshalJSON(req *http.Request, result interface{}) error {
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return errors.Trace(err)
	}
	return json.Unmarshal(data, result)
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

	log.Debugf("request %s vars:%v", postion, result)
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

//ParseURLVars 解析并验证头中参数.
func ParseURLVars(req *http.Request, result interface{}) error {
	return UnmarshalValidate(req, URI, result)
}
