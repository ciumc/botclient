package botclient

import (
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

// Option 客户端配置选项
type Option func(*Client)

// WithHTTPClient 自定义 HTTP 客户端配置
//
// Deprecated: 使用 WithTimeout 或其他具体选项替代。
// 此选项会覆盖共享 Transport 配置，可能增加 goroutine 数量。
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		// 将 http.Client 配置转换为 resty.Client
		c.httpClient = resty.New()
		if client.Timeout > 0 {
			c.httpClient.SetTimeout(client.Timeout)
			c.timeout = client.Timeout
		}
	}
}

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// WithRetry 设置重试策略
func WithRetry(maxRetries int, delay time.Duration) Option {
	return func(c *Client) {
		c.maxRetries = maxRetries
		c.retryDelay = delay
	}
}

// WithLogger 设置日志记录器
func WithLogger(logger Logger) Option {
	return func(c *Client) {
		c.logger = logger
	}
}

// WithMetrics 设置指标收集器
func WithMetrics(metrics MetricsCollector) Option {
	return func(c *Client) {
		c.metrics = metrics
	}
}

// WithMiddleware 添加自定义中间件
func WithMiddleware(middleware Middleware) Option {
	return func(c *Client) {
		c.middlewares = append(c.middlewares, middleware)
	}
}

// WithBaseURL 自定义基础 URL (用于测试或自定义 endpoint)
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithUserAgent 自定义 User-Agent
func WithUserAgent(ua string) Option {
	return func(c *Client) {
		c.userAgent = ua
	}
}