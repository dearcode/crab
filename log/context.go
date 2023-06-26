package log

import "context"

type logKey struct{}

// ToContext 保存log对象到ctx中
func ToContext(ctx context.Context, log *Logger) context.Context {
	return context.WithValue(ctx, logKey{}, log)
}

// FromContext 从ctx中查找日志对象，找不到返回nil
func FromContext(ctx context.Context) *Logger {
	if v := ctx.Value(logKey{}); v != nil {
		return v.(*Logger)
	}
	return nil
}

// FromContextOrDefault 从ctx中查找日志对象，找不到返回默认log对象
func FromContextOrDefault(ctx context.Context) *Logger {
	if v := ctx.Value(logKey{}); v != nil {
		return v.(*Logger)
	}
	return NewLogger()
}
