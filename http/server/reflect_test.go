package server

import (
	"fmt"
	"testing"

	"github.com/hokaccha/go-prettyjson"
)

type TestSub struct {
	Subint int
	SubStr string
}

type TestPtrSub struct {
	PSubint *int
	PSubStr *string
}

type TestStruct struct {
	*TestPtrSub
	TestSub
	Key  string
	PVal *string
}

func TestReflectStruct(t *testing.T) {
	ts := TestStruct{
		TestPtrSub: &TestPtrSub{},
		TestSub: TestSub{
			SubStr: "sub string values",
			Subint: 111,
		},
		Key: "main string key",
	}

	psv := "ptr sub string val"
	ts.PSubStr = &psv
	psi := 222
	ts.PSubint = &psi

	pv := "ptr string values"
	ts.PVal = &pv

	nt := &TestStruct{}

	err := reflectStruct(
		func(k string) (string, bool) {
			switch k {
			case "PSubint":
				return fmt.Sprintf(" %v ", *ts.PSubint), true
			case "PSubStr":
				return *ts.PSubStr, true
			case "Subint":
				return fmt.Sprintf("%v", ts.Subint), true
			case "SubStr":
				return ts.SubStr, true
			case "Key":
				return ts.Key, true
			case "PVal":
				return *ts.PVal, true
			default:
				return "", false
			}
		},
		nt,
	)

	if err != nil {
		t.Fatalf("%v", err)
	}

	b, _ := prettyjson.Marshal(ts)
	nb, _ := prettyjson.Marshal(nt)

	if string(b) != string(nb) {
		t.Fatalf("expect:%s, recv:%s", b, nb)
	}

}
