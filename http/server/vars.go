package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/juju/errors"

	"dearcode.net/crab/log"
	"dearcode.net/crab/meta"
	"dearcode.net/crab/validation"
)

// UnmarshalForm 解析form中或者url中参数, 只支持int和string.
func UnmarshalForm(req *http.Request, result interface{}) error {
	req.ParseForm()
	fvs := req.Form
	hvs := req.Header

	return reflectStruct(func(key string) (string, bool) {
		vals, exist := fvs[key]
		if !exist {
			if vals, exist = hvs[key]; !exist {
				return "", false
			}
		}
		return strings.Join(vals, "\x00"), true
	}, result)
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
		err = UnmarshalForm(req, result)
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
