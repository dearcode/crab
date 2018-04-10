package log

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"
)

type Logger struct {
	out   *os.File
	level LogLevel
	color bool
	mu    sync.Mutex
}

//NewLogger 创建日志对象.
func NewLogger() *Logger {
	return &Logger{
		out:   os.Stdout,
		level: LogDebug,
		color: true,
	}
}

//SetColor 开启/关闭颜色.
func (l *Logger) SetColor(on bool) *Logger {
	l.color = on
	return l
}

//SetLevel 设置日志级别.
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

//SetLevelByString 设置字符串格式的日志级别.
func (l *Logger) SetLevelByString(level string) {
	l.level = StringToLogLevel(level)
}

type pos struct {
	file string
	line int
	name string
}

type posCache struct {
	ps map[uintptr]pos
	sync.RWMutex
}

var mpc = posCache{ps: make(map[uintptr]pos)}

func (l *Logger) caller() (string, int, string) {
	pc, file, line, _ := runtime.Caller(3)

	mpc.RLock()
	p, ok := mpc.ps[pc]
	mpc.RUnlock()

	if ok {
		return p.file, p.line, p.name
	}

	name := runtime.FuncForPC(pc).Name()
	if i := bytes.LastIndexAny([]byte(name), "."); i != -1 {
		name = name[i+1:]
	}
	if i := bytes.LastIndexAny([]byte(file), "/"); i != -1 {
		file = file[i+1:]
	}

	mpc.Lock()
	mpc.ps[pc] = pos{file: file, line: line, name: name}
	mpc.Unlock()

	return file, line, name
}

func (l *Logger) write(t LogLevel, format string, argv ...interface{}) {
	if t > l.level {
		return
	}

	date := time.Now().Format("2006/01/02 15:04:05")

	file, line, name := l.caller()

	//时间，源码文件，源码列
	fmt.Fprintf(l.out, "%s %s:%d ", date, file, line)
	if l.color {
		//颜色开始
		fmt.Fprint(l.out, t.Color())
	}

	//函数名
	fmt.Fprintf(l.out, "%s %s ", t.String(), name)

	fmt.Fprintf(l.out, format, argv...)

	if l.color {
		//颜色结束
		fmt.Fprint(l.out, "\033[0m")
	}

	l.out.WriteString("\n")
}

func (l *Logger) Info(v ...interface{}) {
	l.write(LogInfo, fmt.Sprint(v...))
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.write(LogInfo, format, v...)
}

func (l *Logger) Debug(v ...interface{}) {
	l.write(LogDebug, fmt.Sprint(v...))
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	l.write(LogDebug, format, v...)
}

func (l *Logger) Warning(v ...interface{}) {
	l.write(LogWarning, fmt.Sprint(v...))
}

func (l *Logger) Warningf(format string, v ...interface{}) {
	l.write(LogWarning, format, v...)
}

func (l *Logger) Error(v ...interface{}) {
	l.write(LogError, fmt.Sprint(v...))
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.write(LogError, format, v...)
}

func (l *Logger) Fatal(v ...interface{}) {
	l.write(LogFatal, fmt.Sprint(v...))
	os.Exit(-1)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.write(LogFatal, format, v...)
	os.Exit(-1)
}
