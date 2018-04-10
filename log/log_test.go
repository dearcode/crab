package log

import (
	"testing"
	"time"
)

func TestLog(t *testing.T) {
	Debug("default log begin 11111111")
	Debugf("%s, %+v", "abc", mlog)
	Infof("%v test log", time.Now())

	l := NewLogger()
	l.Info("logger 2222222222")
	l.Errorf("logger color %v xxxxxx", time.Now().UnixNano())
	l.SetColor(false)
	l.Errorf("logger no color %v yyyyyy", time.Now().UnixNano())
	Infof("%v default has color test log", time.Now())

	l.SetOutputFile("./vvv.log").SetRolling(true)
	l.Info(time.Now())
}
