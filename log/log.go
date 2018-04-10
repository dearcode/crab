package log

import (
	"fmt"
)

var mlog = NewLogger()

//SetLevel 设置日志级别.
func SetLevel(level Level) {
	mlog.SetLevel(level)
}

//GetLogLevel 获取日志级别.
func GetLogLevel() Level {
	return mlog.level
}

//Info .
func Info(v ...interface{}) {
	mlog.write(LogInfo, fmt.Sprint(v...))
}

//Infof .
func Infof(format string, v ...interface{}) {
	mlog.write(LogInfo, format, v...)
}

//Debug .
func Debug(v ...interface{}) {
	mlog.write(LogDebug, fmt.Sprint(v...))
}

//Debugf .
func Debugf(format string, v ...interface{}) {
	mlog.write(LogDebug, format, v...)
}

//Warning .
func Warning(v ...interface{}) {
	mlog.write(LogWarning, fmt.Sprint(v...))
}

//Warningf .
func Warningf(format string, v ...interface{}) {
	mlog.write(LogWarning, format, v...)
}

//Error .
func Error(v ...interface{}) {
	mlog.write(LogError, fmt.Sprint(v...))
}

//Errorf .
func Errorf(format string, v ...interface{}) {
	mlog.write(LogError, format, v...)
}

//Fatal .
func Fatal(v ...interface{}) {
	mlog.write(LogFatal, fmt.Sprint(v...))
}

//Fatalf .
func Fatalf(format string, v ...interface{}) {
	mlog.write(LogFatal, format, v...)
}

//SetLevelByString 设置日志级别.
func SetLevelByString(level string) {
	mlog.SetLevelByString(level)
}

//SetColor 设置是否显示颜色.
func SetColor(color bool) {
	mlog.SetColor(color)
}
