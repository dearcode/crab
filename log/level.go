package log

//Level 日志级别.
type Level int

const (
	//LogFatal fatal.
	LogFatal Level = iota
	//LogError error.
	LogError
	//LogWarning warning.
	LogWarning
	//LogInfo info.
	LogInfo
	//LogDebug debug.
	LogDebug
)

//stringToLevel 字符串转Level.
func stringToLevel(level string) Level {
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

//Level Level 转字符串.
func (l Level) String() string {
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

//color Level转颜色.
func (l Level) color() string {
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
