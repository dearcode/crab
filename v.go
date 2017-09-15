package main

import (
	"fmt"
	"reflect"
)

type testServer struct {
	A int
}

func test(t reflect.Type) {
	for i := 0; i < 10; i++ {
		fmt.Printf("new:%p\n", reflect.New(t).Interface())
	}
}

func main() {
    t := &testServer{}

    tt := reflect.TypeOf(t)

    test(tt.Elem())


}
