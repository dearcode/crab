package server

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hokaccha/go-prettyjson"
)

type TestSub struct {
	Subint int
	SubStr []string
}

type TestSubSecond struct {
	Secondint int
	SecondStr string
}

type TestPtrSub struct {
	PSubint *int
	PSubStr *string
	Second  *TestSubSecond
}

type TestStruct struct {
	*TestPtrSub
	TestSub
	Key  string
	PVal *string
}

func TestReflectStruct(t *testing.T) {
	ts := TestStruct{
		TestPtrSub: &TestPtrSub{
			Second: &TestSubSecond{
				SecondStr: "second str",
				Secondint: 333,
			},
		},
		TestSub: TestSub{
			SubStr: []string{"s1", "s2"},
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
				return strings.Join(ts.SubStr, "\x00"), true
			case "Key":
				return ts.Key, true
			case "PVal":
				return *ts.PVal, true
			case "Secondint":
				return fmt.Sprintf("%v", ts.Second.Secondint), true
			case "SecondStr":
				return ts.Second.SecondStr, true
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
