package log

import ()

type LogLevel int

const (
	LogFatal LogLevel = iota
	LogError
	LogWarning
	LogInfo
	LogDebug
)

//StringToLogLevel 字符串转LogLevel.
func StringToLogLevel(level string) LogLevel {
	switch level {
	case "fatal":
		return LogFatal
	case "error":
		return LogError
	case "warn":
		return LogWarning
	case "warning":
		return LogWarning
	case "debug":
		return LogDebug
	case "info":
		return LogInfo
	}
	return LogDebug
}

//LogLevel loglevel 转字符串.
func (l LogLevel) String() string {
	switch l {
	case LogFatal:
		return "fatal"
	case LogError:
		return "error"
	case LogWarning:
		return "warning"
	case LogDebug:
		return "debug"
	case LogInfo:
		return "info"
	}
	return "unknown"
}

//LogLevel loglevel转颜色.
func (l LogLevel) Color() string {
	switch l {
	case LogFatal:
		return "\033[0;31m"
	case LogError:
		return "\033[0;31m"
	case LogWarning:
		return "\033[0;33m"
	case LogDebug:
		return "\033[0;36m"
	case LogInfo:
		return "\033[0;32m"
	}
	return "\033[0;37m"
}
