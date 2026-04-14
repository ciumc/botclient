package botclient

import (
	"context"
	"time"
)

// Logger 日志接口
type Logger interface {
	Info(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
	Debug(ctx context.Context, msg string, args ...any)
}

// MetricsCollector 指标收集接口
type MetricsCollector interface {
	Record(webhookKey string, msgType string, duration time.Duration, success bool)
}

// noopLogger 空日志实现
type noopLogger struct{}

func (l *noopLogger) Info(ctx context.Context, msg string, args ...any)  {}
func (l *noopLogger) Error(ctx context.Context, msg string, args ...any) {}
func (l *noopLogger) Debug(ctx context.Context, msg string, args ...any) {}

// noopMetrics 空指标实现
type noopMetrics struct{}

func (m *noopMetrics) Record(webhookKey string, msgType string, duration time.Duration, success bool) {}