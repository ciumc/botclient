package botclient

import (
	"context"
	"io"
	"time"
)

// Middleware 中间件函数签名
type Middleware func(next Handler) Handler

// Handler 处理函数签名
type Handler func(ctx context.Context, req *Request) (*Response, error)

// Request 请求结构
type Request struct {
	WebhookKey string
	Message    Message
	MediaType  MediaType // 用于上传媒体
	FilePath   string    // 文件路径
	Reader     io.ReadCloser
}

// Response 响应结构
type Response struct {
	StatusCode int
	ErrCode    int
	ErrMsg     string
	MediaId    string // 上传媒体时返回
}

// chain 将多个中间件组合成一个中间件
// 中间件按从后向前的顺序执行（洋葱模型）
// 示例: chain(m1, m2, m3)(handler) 等价于 m1(m2(m3(handler)))
func chain(middlewares ...Middleware) Middleware {
	return func(next Handler) Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// --- 内置中间件 ---

// LoggingMiddleware 日志中间件
func LoggingMiddleware(logger Logger) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, req *Request) (*Response, error) {
			start := time.Now()
			resp, err := next(ctx, req)

			duration := time.Since(start)
			msgType := ""
			if req.Message != nil {
				msgType = req.Message.Type()
			}

			if err != nil {
				logger.Error(ctx, "request failed",
					"error", err,
					"duration", duration,
					"webhook_key", req.WebhookKey,
					"msg_type", msgType,
				)
			} else {
				logger.Info(ctx, "request succeeded",
					"duration", duration,
					"webhook_key", req.WebhookKey,
					"msg_type", msgType,
				)
			}

			return resp, err
		}
	}
}

// RetryMiddleware 重试中间件（指数退避）
// 参数:
//   - maxRetries: 最大重试次数（不含首次尝试），总尝试次数 = maxRetries + 1
//   - initialDelay: 初始延迟时间，每次重试后延迟翻倍
func RetryMiddleware(maxRetries int, initialDelay time.Duration) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, req *Request) (*Response, error) {
			var lastErr error
			delay := initialDelay

			for i := 0; i <= maxRetries; i++ {
				if i > 0 {
					select {
					case <-ctx.Done():
						return nil, ctx.Err()
					case <-time.After(delay):
					}
					// 指数退避
					delay = delay * 2
				}

				resp, err := next(ctx, req)
				if err == nil {
					return resp, nil
				}

				// 不可重试的错误直接返回
				if !IsRetryable(err) {
					return nil, err
				}

				lastErr = err
			}

			return nil, lastErr
		}
	}
}

// MetricsMiddleware 指标收集中间件
func MetricsMiddleware(metrics MetricsCollector) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, req *Request) (*Response, error) {
			start := time.Now()
			resp, err := next(ctx, req)

			msgType := ""
			if req.Message != nil {
				msgType = req.Message.Type()
			}

			metrics.Record(
				req.WebhookKey,
				msgType,
				time.Since(start),
				err == nil,
			)

			return resp, err
		}
	}
}

// TimeoutMiddleware 超时中间件
func TimeoutMiddleware(timeout time.Duration) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, req *Request) (*Response, error) {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			resp, err := next(ctx, req)
			if ctx.Err() == context.DeadlineExceeded {
				return nil, NewError(CodeSystemError, "request timeout", ErrTimeout)
			}
			return resp, err
		}
	}
}