package validation

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	//"github.com/juju/errors"
)

// ValidFormer valid interface
type ValidFormer interface {
	Valid(*Validation)
}

// Error show the error
type Error struct {
	Message, Key, Name, Field, Tmpl string
	Value                           interface{}
	LimitValue                      interface{}
}

// String Returns the Message.
func (e *Error) String() string {
	if e == nil {
		return ""
	}
	return e.Message
}

// Implement Error interface.
// Return e.String()
func (e *Error) Error() string { return e.String() }

// Result is returned from every validation method.
// It provides an indication of success, and a pointer to the Error (if any).
type Result struct {
	Error *Error
	Ok    bool
}

// Key Get Result by given key string.
func (r *Result) Key(key string) *Result {
	if r.Error != nil {
		r.Error.Key = key
	}
	return r
}

// Message Set Result message by string or format string with args
func (r *Result) Message(message string, args ...interface{}) *Result {
	if r.Error != nil {
		if len(args) == 0 {
			r.Error.Message = message
		} else {
			r.Error.Message = fmt.Sprintf(message, args...)
		}
	}
	return r
}

// A Validation context manages data validation and error messages.
type Validation struct {
	Errors    []*Error
	ErrorsMap map[string]*Error
}

// Clear Clean all ValidationError.
func (v *Validation) Clear() {
	v.Errors = []*Error{}
	v.ErrorsMap = nil
}

// HasErrors Has ValidationError nor not.
func (v *Validation) HasErrors() bool {
	return len(v.Errors) > 0
}

// ErrorMap Return the errors mapped by key.
// If there are multiple validation errors associated with a single key, the
// first one "wins".  (Typically the first validation will be the more basic).
func (v *Validation) ErrorMap() map[string]*Error {
	return v.ErrorsMap
}

// Error Add an error to the validation context.
func (v *Validation) Error(message string, args ...interface{}) *Result {
	result := (&Result{
		Ok:    false,
		Error: &Error{},
	}).Message(message, args...)
	v.Errors = append(v.Errors, result.Error)
	return result
}

// Required Test that the argument is non-nil and non-empty (if string or list)
func (v *Validation) Required(obj interface{}, key string) *Result {
	return v.apply(Required{key}, obj)
}

// Min Test that the obj is greater than min if obj's type is int
func (v *Validation) Min(obj interface{}, min int, key string) *Result {
	return v.apply(Min{min, key}, obj)
}

// Max Test that the obj is less than max if obj's type is int
func (v *Validation) Max(obj interface{}, max int, key string) *Result {
	return v.apply(Max{max, key}, obj)
}

// Range Test that the obj is between mni and max if obj's type is int
func (v *Validation) Range(obj interface{}, min, max int, key string) *Result {
	return v.apply(Range{Min{Min: min}, Max{Max: max}, key}, obj)
}

// MinSize Test that the obj is longer than min size if type is string or slice
func (v *Validation) MinSize(obj interface{}, min int, key string) *Result {
	return v.apply(MinSize{min, key}, obj)
}

// MaxSize Test that the obj is shorter than max size if type is string or slice
func (v *Validation) MaxSize(obj interface{}, max int, key string) *Result {
	return v.apply(MaxSize{max, key}, obj)
}

// Length Test that the obj is same length to n if type is string or slice
func (v *Validation) Length(obj interface{}, n int, key string) *Result {
	return v.apply(Length{n, key}, obj)
}

// Alpha Test that the obj is [a-zA-Z] if type is string
func (v *Validation) Alpha(obj interface{}, key string) *Result {
	return v.apply(Alpha{key}, obj)
}

// Numeric Test that the obj is [0-9] if type is string
func (v *Validation) Numeric(obj interface{}, key string) *Result {
	return v.apply(Numeric{key}, obj)
}

// AlphaNumeric Test that the obj is [0-9a-zA-Z] if type is string
func (v *Validation) AlphaNumeric(obj interface{}, key string) *Result {
	return v.apply(AlphaNumeric{key}, obj)
}

// Match Test that the obj matches regexp if type is string
func (v *Validation) Match(obj interface{}, regex *regexp.Regexp, key string) *Result {
	return v.apply(Match{regex, key}, obj)
}

// NoMatch Test that the obj doesn't match regexp if type is string
func (v *Validation) NoMatch(obj interface{}, regex *regexp.Regexp, key string) *Result {
	return v.apply(NoMatch{Match{Regexp: regex}, key}, obj)
}

// AlphaDash Test that the obj is [0-9a-zA-Z_-] if type is string
func (v *Validation) AlphaDash(obj interface{}, key string) *Result {
	return v.apply(AlphaDash{NoMatch{Match: Match{Regexp: alphaDashPattern}}, key}, obj)
}

// Email Test that the obj is email address if type is string
func (v *Validation) Email(obj interface{}, key string) *Result {
	return v.apply(Email{Match{Regexp: emailPattern}, key}, obj)
}

// IP Test that the obj is IP address if type is string
func (v *Validation) IP(obj interface{}, key string) *Result {
	return v.apply(IP{Match{Regexp: ipPattern}, key}, obj)
}

// Base64 Test that the obj is base64 encoded if type is string
func (v *Validation) Base64(obj interface{}, key string) *Result {
	return v.apply(Base64{Match{Regexp: base64Pattern}, key}, obj)
}

// Mobile Test that the obj is chinese mobile number if type is string
func (v *Validation) Mobile(obj interface{}, key string) *Result {
	return v.apply(Mobile{Match{Regexp: mobilePattern}, key}, obj)
}

// Tel Test that the obj is chinese telephone number if type is string
func (v *Validation) Tel(obj interface{}, key string) *Result {
	return v.apply(Tel{Match{Regexp: telPattern}, key}, obj)
}

// Phone Test that the obj is chinese mobile or telephone number if type is string
func (v *Validation) Phone(obj interface{}, key string) *Result {
	return v.apply(Phone{Mobile{Match: Match{Regexp: mobilePattern}},
		Tel{Match: Match{Regexp: telPattern}}, key}, obj)
}

// ZipCode Test that the obj is chinese zip code if type is string
func (v *Validation) ZipCode(obj interface{}, key string) *Result {
	return v.apply(ZipCode{Match{Regexp: zipCodePattern}, key}, obj)
}

func (v *Validation) apply(chk Validator, obj interface{}) *Result {
	if chk.IsSatisfied(obj) {
		return &Result{Ok: true}
	}

	// Add the error to the validation context.
	key := chk.GetKey()
	Name := key
	Field := ""

	parts := strings.Split(key, ".")
	if len(parts) == 2 {
		Field = parts[0]
		Name = parts[1]
	}

	err := &Error{
		Message:    chk.DefaultMessage(),
		Key:        key,
		Name:       Name,
		Field:      Field,
		Value:      obj,
		Tmpl:       MessageTmpls[Name],
		LimitValue: chk.GetLimitValue(),
	}
	v.setError(err)

	// Also return it in the result.
	return &Result{
		Ok:    false,
		Error: err,
	}
}

func (v *Validation) setError(err *Error) {
	v.Errors = append(v.Errors, err)
	if v.ErrorsMap == nil {
		v.ErrorsMap = make(map[string]*Error)
	}
	if _, ok := v.ErrorsMap[err.Field]; !ok {
		v.ErrorsMap[err.Field] = err
	}
}

// SetError Set error message for one field in ValidationError
func (v *Validation) SetError(fieldName string, errMsg string) *Error {
	err := &Error{Key: fieldName, Field: fieldName, Tmpl: errMsg, Message: errMsg}
	v.setError(err)
	return err
}

// Check Apply a group of validators to a field, in order, and return the
// ValidationResult from the first one that fails, or the last one that
// succeeds.
func (v *Validation) Check(obj interface{}, checks ...Validator) *Result {
	var result *Result
	for _, check := range checks {
		result = v.apply(check, obj)
		if !result.Ok {
			return result
		}
	}
	return result
}

// Valid Validate a struct.
// the obj parameter must be a struct or a struct pointer
func (v *Validation) Valid(obj interface{}) (b bool, err error) {
	objT := reflect.TypeOf(obj)
	objV := reflect.ValueOf(obj)
	switch {
	case isStruct(objT):
	case isStructPtr(objT):
		objT = objT.Elem()
		objV = objV.Elem()
	default:
		err = fmt.Errorf("%v must be a struct or a struct pointer", obj)
		return
	}

	for i := 0; i < objT.NumField(); i++ {
		switch objT.Field(i).Type.Kind() {
		case reflect.Array, reflect.Slice:
			for si := 0; si < objV.Field(i).Len(); si++ {
				sv := objV.Field(i).Index(si)
				if sv.Kind() != reflect.Struct {
					continue
				}
				b, err = v.Valid(sv.Interface())
				if err != nil {
					return b, fmt.Errorf("[]%v.%v", sv.Type().Name(), err.Error())
				}
			}
			return
		case reflect.Struct:
			sv := objV.Field(i)
			b, err = v.Valid(sv.Interface())
			if err != nil {
				return b, fmt.Errorf("%v.%v", sv.Type().Name(), err.Error())
			}
			return
		}
		var vfs []ValidFunc
		if vfs, err = getValidFuncs(objT.Field(i)); err != nil {
			return
		}

		var rs []reflect.Value
		for _, vf := range vfs {
			rs, err = funcs.Call(vf.Name, mergeParam(v, objV.Field(i).Interface(), vf.Params)...)
			if err != nil {
				return
			}
			result, _ := rs[0].Interface().(*Result)
			if result.Error != nil {
				return false, fmt.Errorf("%v %v", vf.Aliase, result.Error.Error())
			}
		}
	}

	if !v.HasErrors() {
		if form, ok := obj.(ValidFormer); ok {
			form.Valid(v)
		}
	}

	return !v.HasErrors(), nil
}
