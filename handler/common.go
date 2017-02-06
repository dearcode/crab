package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type VariablePostion int
type InterfaceMethod int

const (
	URI VariablePostion = iota
	HEADER
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
	GET InterfaceMethod = iota
	POST
	PUT
	DELETE
	RESTful
)

//String 类型转字符串
func (m InterfaceMethod) String() string {
	switch m {
	case GET:
		return "GET"
	case POST:
		return "POST"
	case PUT:
		return "PUT"
	case DELETE:
		return "DELETE"
	case RESTful:
		return "RESTful"
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
