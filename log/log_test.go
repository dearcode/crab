package log

import (
	"context"
	"os"
	"testing"
	"time"
)

const (
	testLogFile = "./test.log"
)

func TestLog(t *testing.T) {
	Debug("default log begin")
	Infof("%v test log", time.Now())

	l := NewLogger()
	l.Debug("logger 1111111111")
	l.Info("logger 2222222222")
	l.Warningf("logger 33333 %v", time.Now())
	l.Errorf("logger color %v xxxxxx", time.Now().UnixNano())
	l.SetColor(false)
	l.Errorf("logger no color %v yyyyyy", time.Now().UnixNano())
	Infof("%v default has color test log", time.Now())

	l.SetOutputFile(testLogFile).SetRolling(true)
	l.Info(time.Now())
	os.Remove(testLogFile)

}

func TestLogContext(t *testing.T) {
	l := NewLogger()
	ctx := context.Background()

	ctx = ToContext(ctx, l)

	l2 := FromContext(ctx)
	if l2 == nil {
		t.Fatalf("expect l2 != nil, FromContext recv:%v", l2)
	}

	l3 := FromContext(context.Background())
	if l3 != nil {
		t.Fatalf("expect l3 == nil, FromContext recv:%v", l3)
	}

	l2.Infof("ok")
}
