package main

import (
	"time"

	"dearcode.net/crab/log"
)

func main() {
	log.Debug("default log begin")
	log.Infof("%v test log", time.Now())

	l := log.NewLogger()
	l.Debug("logger 1111111111")
	l.Info("logger 2222222222")
	l.Warningf("logger 33333 %v", time.Now())
	l.Errorf("logger color %v xxxxxx", time.Now().UnixNano())

	//关闭颜色显示
	l.SetColor(false)

	l.Errorf("logger no color %v yyyyyy", time.Now().UnixNano())
	log.Infof("%v default has color test log", time.Now())

	//指定输出文件
	l.SetOutputFile("./vvv.log").SetRolling(true)
	l.Info(time.Now())

}
