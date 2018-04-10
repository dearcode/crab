package log

import (
	"fmt"
	"os"
)

var mlog = &Logger{out: os.Stdout, level: LogDebug, color: true}

func SetLevel(level LogLevel) {
	mlog.SetLevel(level)
}

func GetLogLevel() LogLevel {
	return mlog.level
}

func Info(v ...interface{}) {
	mlog.write(LogInfo, fmt.Sprint(v...))
}

func Infof(format string, v ...interface{}) {
	mlog.write(LogInfo, format, v...)
}

func Debug(v ...interface{}) {
	mlog.write(LogDebug,fmt.Sprint( v...))
}

func Debugf(format string, v ...interface{}) {
	mlog.write(LogDebug, format, v...)
}

func Warning(v ...interface{}) {
	mlog.write(LogWarning, fmt.Sprint( v...))
}

func Warningf(format string, v ...interface{}) {
	mlog.write(LogWarning, format, v...)
}

func Error(v ...interface{}) {
	mlog.write(LogError, fmt.Sprint( v...))
}

func Errorf(format string, v ...interface{}) {
	mlog.write(LogError, format, v...)
}

func Fatal(v ...interface{}) {
	mlog.write(LogFatal, fmt.Sprint( v...))
}

func Fatalf(format string, v ...interface{}) {
	mlog.write(LogFatal, format, v...)
}

func SetLevelByString(level string) {
	mlog.SetLevelByString(level)
}

func SetColor(color bool) {
	mlog.SetColor(color)
}
