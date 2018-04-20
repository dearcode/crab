package log

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"
)

type pos struct {
	file     string
	function string
}

type Logger struct {
	rolling  bool
	fileName string
	fileTime time.Time
	file     *os.File
	out      *bufio.Writer
	level    LogLevel
	color    bool
	posCache map[uintptr]pos
	mu       sync.Mutex
}

//NewLogger 创建日志对象.
func NewLogger() *Logger {
	return &Logger{
		out:      bufio.NewWriter(os.Stdout),
		level:    LogDebug,
		color:    true,
		posCache: make(map[uintptr]pos),
	}
}

//SetColor 开启/关闭颜色.
func (l *Logger) SetColor(on bool) *Logger {
	l.color = on
	return l
}

//SetRolling 每天生成一个文件.
func (l *Logger) SetRolling(on bool) *Logger {
	l.rolling = on
	return l
}

//SetOutputFile 初始化时设置输出文件.
func (l *Logger) SetOutputFile(path string) *Logger {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(fmt.Sprintf("open %s error:%v", path, err.Error()))
	}

	now, _ := time.ParseInLocation("20060102", time.Now().Format("20060102"), time.Local)
	l.fileTime = now.Add(time.Hour * 24)
	l.file = f
	l.fileName = path
	l.out = bufio.NewWriter(f)

	return l
}

//SetLevel 设置日志级别.
func (l *Logger) SetLevel(level LogLevel) *Logger {
	l.level = level
	return l
}

//SetLevelByString 设置字符串格式的日志级别.
func (l *Logger) SetLevelByString(level string) *Logger {
	l.level = StringToLogLevel(level)
	return l
}

func (l *Logger) caller() (string, string) {
	pc, file, line, _ := runtime.Caller(3)

	p, ok := l.posCache[pc]
	if ok {
		return p.file, p.function
	}

	name := runtime.FuncForPC(pc).Name()
	if i := bytes.LastIndexAny([]byte(name), "."); i != -1 {
		name = name[i+1:]
	}
	if i := bytes.LastIndexAny([]byte(file), "/"); i != -1 {
		file = file[i+1:]
	}

	p = pos{file: fmt.Sprintf("%s:%d", file, line), function: name}
	l.posCache[pc] = p

	return p.file, p.function
}

func (l *Logger) rotate(now time.Time) {
	if !l.rolling || l.file == nil || now.Before(l.fileTime) {
		return
	}

	l.out.Flush()
	l.file.Close()

	oldFile := l.fileName + time.Now().Format("20060102")

	os.Rename(l.fileName, oldFile)

	l.SetOutputFile(l.fileName)
}

func (l *Logger) write(t LogLevel, format string, argv ...interface{}) {
	if t > l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	l.rotate(now)
	date := now.Format("2006/01/02 15:04:05")

	file, function := l.caller()

	//时间，源码文件和行号
	l.out.WriteString(date)
	l.out.WriteString(" ")
	l.out.WriteString(file)
	l.out.WriteString(" ")

	if l.color {
		//颜色开始
		l.out.WriteString(t.Color())
	}

	//日志级别
	l.out.WriteString(t.String())

	//函数名
	l.out.WriteString(function)

	//日志正文
	fmt.Fprintf(l.out, format, argv...)

	if l.color {
		//颜色结束
		l.out.WriteString("\033[0m")
	}

	l.out.WriteString("\n")

	l.out.Flush()
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
