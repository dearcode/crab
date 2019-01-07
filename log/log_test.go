package log

import (
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
