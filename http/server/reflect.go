package server

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/juju/errors"
)

//fieldName 解析field中的json,db等标签，提取别名
func fieldName(f reflect.StructField) string {
	if name := f.Tag.Get("json"); name != "" {
		return name
	}

	if name := f.Tag.Get("db"); name != "" {
		return name
	}

	if name := f.Tag.Get("cfg"); name != "" {
		return name
	}

	return f.Name
}

func setValue(val string, rv reflect.Value) error {
	val = strings.TrimSpace(val)
	switch rv.Type().Kind() {
	case reflect.Int, reflect.Int64:
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return errors.Trace(err)
		}
		rv.SetInt(i)

	case reflect.Uint, reflect.Uint64:
		u, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return errors.Trace(err)
		}
		rv.SetUint(u)

	case reflect.String:
		rv.SetString(val)

	case reflect.Bool:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return errors.Trace(err)
		}
		rv.SetBool(b)
	}

	return nil
}

func setValueOrPtr(f reflect.StructField, val string, rv reflect.Value) error {
	var pv reflect.Value
	v := rv
	if f.Type.Kind() == reflect.Ptr {
		pv = reflect.New(f.Type.Elem())
		v = pv.Elem()
	}

	if err := setValue(val, v); err != nil {
		return errors.Trace(err)
	}

	if f.Type.Kind() == reflect.Ptr {
		rv.Set(pv)
	}

	return nil
}

type getValueFunc func(string) (string, bool)

func reflectKeyValue(getVal getValueFunc, rt reflect.Type, rv reflect.Value, level int) error {
	if rv.Type().Kind() == reflect.Ptr && rv.IsNil() {
		pv := reflect.New(rt.Elem())
		rv.Set(pv)
	}

	if rt.Kind() == reflect.Ptr && rt.Elem().Kind() == reflect.Struct {
		rt = rt.Elem()
		rv = rv.Elem()
	}

	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if f.PkgPath != "" && !f.Anonymous {
			continue
		}

		if (f.Type.Kind() == reflect.Struct) || (f.Type.Kind() == reflect.Ptr && f.Type.Elem().Kind() == reflect.Struct) {
			if level < 1 {
				continue
			}
			if err := reflectKeyValue(getVal, f.Type, rv.Field(i), level-1); err != nil {
				return errors.Trace(err)
			}
			continue
		}

		key := fieldName(f)
		val, ok := getVal(key)
		if !ok {
			continue
		}

		if err := setValueOrPtr(f, val, rv.Field(i)); err != nil {
			return errors.Trace(err)
		}
	}

	return nil
}

//reflectStruct 反射结构体中字段，根据字段名或标签取对应结果，并赋值返回
func reflectStruct(getVal getValueFunc, obj interface{}) error {
	rt := reflect.TypeOf(obj)
	rv := reflect.ValueOf(obj)
	return reflectKeyValue(getVal, rt, rv, 1)
}
