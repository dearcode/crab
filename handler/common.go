package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
)

//VariablePostion 变量位置.
type VariablePostion int

//Method 请求方式.
type Method int

const (
	//URI 参数在uri里.
	URI VariablePostion = iota
	//HEADER 参数在头里.
	HEADER
	//BODY 参数在body里.
	BODY
)

//String 类型转字符串
func (p VariablePostion) String() string {
	switch p {
	case URI:
		return "URI"
	case HEADER:
		return "HEADER"
	case BODY:
		return "BODY"
	}
	return "NIL"
}

const (
	//GET http method.
	GET Method = iota
	//POST http method.
	POST
	//PUT http method.
	PUT
	//DELETE http method.
	DELETE
	//RESTFul any method, may be get,post,put or delete.
	RESTFul
)

//String 类型转字符串
func (m Method) String() string {
	switch m {
	case GET:
		return "GET"
	case POST:
		return "POST"
	case PUT:
		return "PUT"
	case DELETE:
		return "DELETE"
	case RESTFul:
		return "RESTFul"
	}
	return "NIL"
}

//Response 通用返回结果
type Response struct {
	Status  int
	Message string      `json:",omitempty"`
	Data    interface{} `json:",omitempty"`
}

//SendResponse 返回结果，支持json
func SendResponse(w http.ResponseWriter, status int, f string, args ...interface{}) {
	w.Header().Add("Content-Type", "application/json")
	r := Response{Status: status, Message: f}
	if len(args) > 0 {
		r.Message = fmt.Sprintf(f, args)
	}

	buf, _ := json.Marshal(&r)
	w.Write(buf)
}
