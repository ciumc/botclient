// Package botclient provides a client for sending messages to enterprise WeChat robots.
//
// Example:
//
//	client := botclient.New("your-webhook-key")
//	err := client.Send(context.Background(), botclient.Text("Hello, World!"))
package botclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"golang.org/x/sync/errgroup"
)

const (
	defaultBaseURL      = "https://qyapi.weixin.qq.com/cgi-bin/webhook"
	defaultTimeout      = 30 * time.Second
	defaultUserAgent    = "botclient/v2"
	defaultIdleTimeout  = 90 * time.Second // 连接池空闲超时
	defaultMaxIdleConn  = 10               // 最大空闲连接数
	defaultMaxIdleConnPerHost = 2          // 每个 host 最大空闲连接数
)

// sharedTransport 共享的 HTTP Transport（减少 goroutine 数量）
var sharedTransport = &http.Transport{
	IdleConnTimeout:       defaultIdleTimeout,
	MaxIdleConns:          defaultMaxIdleConn,
	MaxIdleConnsPerHost:   defaultMaxIdleConnPerHost,
	DisableKeepAlives:     false, // 启用 Keep-Alive
	ForceAttemptHTTP2:     true,
}

// Sender 客户端接口，用于测试和抽象
type Sender interface {
	Send(ctx context.Context, msg Message) error
	SendWithKey(ctx context.Context, webhookKey string, msg Message) error
}

// Client 企业微信机器人客户端
// Client 在创建后可安全并发使用
type Client struct {
	httpClient  *resty.Client
	webhookKey  string
	baseURL     string
	userAgent   string
	timeout     time.Duration
	middlewares []Middleware
	handler     Handler // 缓存的处理器链，避免每次 Send 重新构建
	logger      Logger
	metrics     MetricsCollector
	maxRetries  int
	retryDelay  time.Duration
}

// New 创建客户端实例
func New(webhookKey string, opts ...Option) *Client {
	// 使用共享 Transport 创建 resty Client，减少 goroutine
	httpClient := resty.NewWithClient(&http.Client{
		Transport: sharedTransport,
		Timeout:   defaultTimeout,
	})

	c := &Client{
		webhookKey: webhookKey,
		baseURL:    defaultBaseURL,
		userAgent:  defaultUserAgent,
		timeout:    defaultTimeout,
		httpClient: httpClient,
		logger:     &noopLogger{},
		metrics:    &noopMetrics{},
		middlewares: make([]Middleware, 0),
	}

	// 应用选项
	for _, opt := range opts {
		opt(c)
	}

	// 确保 HTTP 客户端配置
	if c.timeout > 0 {
		c.httpClient.SetTimeout(c.timeout)
	}

	// 设置 User-Agent
	c.httpClient.SetHeader("User-Agent", c.userAgent)

	// 自动应用内置中间件
	// 1. 重试中间件
	if c.maxRetries > 0 {
		c.middlewares = append(c.middlewares, RetryMiddleware(c.maxRetries, c.retryDelay))
	}

	// 2. 日志中间件 (如果不是 noopLogger)
	if _, ok := c.logger.(*noopLogger); !ok {
		c.middlewares = append(c.middlewares, LoggingMiddleware(c.logger))
	}

	// 3. 指标中间件 (如果不是 noopMetrics)
	if _, ok := c.metrics.(*noopMetrics); !ok {
		c.middlewares = append(c.middlewares, MetricsMiddleware(c.metrics))
	}

	// 构建并缓存处理器链（避免每次 Send 重新构建）
	c.handler = c.buildHandler()

	return c
}

// buildHandler 构建处理器链
// 使用 chain 函数组合中间件，返回最终的处理器
func (c *Client) buildHandler() Handler {
	if len(c.middlewares) == 0 {
		return c.doSend
	}
	return chain(c.middlewares...)(c.doSend)
}

// Send 发送消息
func (c *Client) Send(ctx context.Context, msg Message) error {
	if msg == nil {
		return NewError(CodeInvalidMsgType, "message is nil", ErrEmptyMessage)
	}

	// 验证消息
	if err := msg.Validate(); err != nil {
		return err
	}

	// 使用缓存的处理器链
	_, err := c.handler(ctx, &Request{
		WebhookKey: c.webhookKey,
		Message:    msg,
	})
	return err
}

// doSend 中间件处理函数
func (c *Client) doSend(ctx context.Context, req *Request) (*Response, error) {
	if req.Message == nil {
		return nil, NewError(CodeInvalidMsgType, "message is nil", ErrEmptyMessage)
	}

	err := c.doSendDirect(ctx, req.Message)
	if err != nil {
		return nil, err
	}

	return &Response{StatusCode: 200, ErrCode: 0}, nil
}

// doSendDirect 直接发送消息
func (c *Client) doSendDirect(ctx context.Context, msg Message) error {
	// 构建请求
	httpReq := c.httpClient.NewRequest()
	httpReq.SetContext(ctx)
	httpReq.SetHeader("Content-Type", "application/json")

	// 构建消息体
	body := buildMessageBody(msg)
	httpReq.SetBody(body)

	// 发送请求
	url := fmt.Sprintf("%s/send?key=%s", c.baseURL, c.webhookKey)
	resp, err := httpReq.Post(url)
	if err != nil {
		return NewError(CodeSystemError, "http request failed", err)
	}

	if !resp.IsSuccess() {
		return NewError(CodeSystemError, fmt.Sprintf("status code: %s", resp.Status()), ErrRequestFailed)
	}

	// 解析响应
	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return NewError(CodeSystemError, "parse response failed", err)
	}

	if result.ErrCode != 0 {
		return NewError(ErrorCode(result.ErrCode), result.ErrMsg, nil)
	}

	return nil
}

// SendWithKey 使用指定 webhook key 发送消息
// 注意：此方法复用当前客户端的 HTTP 连接和中间件配置
func (c *Client) SendWithKey(ctx context.Context, webhookKey string, msg Message) error {
	key, err := c.getWebhookKey(webhookKey)
	if err != nil {
		return err
	}

	// 使用缓存的处理器链，仅替换 webhookKey
	_, err = c.handler(ctx, &Request{
		WebhookKey: key,
		Message:    msg,
	})
	return err
}

// getWebhookKey 获取有效的 webhook key
// 如果 override 为空，使用客户端默认 key
func (c *Client) getWebhookKey(override string) (string, error) {
	key := override
	if key == "" {
		key = c.webhookKey
	}
	if key == "" {
		return "", NewError(CodeInvalidKey, "webhook key is empty", ErrInvalidWebhookKey)
	}
	return key, nil
}

// buildMessageBody 构建消息体
func buildMessageBody(msg Message) map[string]any {
	body := map[string]any{
		"msgtype": msg.Type(),
	}

	switch msg.Type() {
	case MsgTypeText:
		body["text"] = msg
	case MsgTypeMarkdown:
		body["markdown"] = msg
	case MsgTypeImage:
		body["image"] = msg
	case MsgTypeNews:
		body["news"] = msg
	case MsgTypeFile:
		body["file"] = msg
	case MsgTypeTemplateCard:
		// 模板卡片需要提取内部结构
		if tc, ok := msg.(*TemplateCardMessage); ok {
			body["template_card"] = tc.TemplateCard
		} else {
			body["template_card"] = msg
		}
	}

	return body
}

// --- Builder 模式 ---

// Builder 客户端构建器
type Builder struct {
	webhookKey string
	opts       []Option
}

// NewBuilder 创建构建器
func NewBuilder(webhookKey string) *Builder {
	return &Builder{
		webhookKey: webhookKey,
		opts:       make([]Option, 0),
	}
}

// WithTimeout 设置超时
func (b *Builder) WithTimeout(timeout time.Duration) *Builder {
	b.opts = append(b.opts, WithTimeout(timeout))
	return b
}

// WithRetry 设置重试
func (b *Builder) WithRetry(maxRetries int, delay time.Duration) *Builder {
	b.opts = append(b.opts, WithRetry(maxRetries, delay))
	return b
}

// WithLogger 设置日志
func (b *Builder) WithLogger(logger Logger) *Builder {
	b.opts = append(b.opts, WithLogger(logger))
	return b
}

// WithMetrics 设置指标
func (b *Builder) WithMetrics(metrics MetricsCollector) *Builder {
	b.opts = append(b.opts, WithMetrics(metrics))
	return b
}

// WithMiddleware 添加中间件
func (b *Builder) WithMiddleware(m Middleware) *Builder {
	b.opts = append(b.opts, WithMiddleware(m))
	return b
}

// WithBaseURL 设置基础 URL
func (b *Builder) WithBaseURL(baseURL string) *Builder {
	b.opts = append(b.opts, WithBaseURL(baseURL))
	return b
}

// Build 构建客户端
func (b *Builder) Build() *Client {
	return New(b.webhookKey, b.opts...)
}

// --- 批量发送 ---

// BatchSend 批量发送消息到多个 webhook（并发执行）
// 使用 errgroup 实现并发控制，任一失败不影响其他发送
func BatchSend(ctx context.Context, webhookKeys []string, msg Message, opts ...Option) map[string]error {
	results := make(map[string]error)
	var mu sync.Mutex

	// 创建单个 Client 并复用（避免每次创建新 Client 产生 goroutine）
	baseClient := New("", opts...)

	// 使用 errgroup 进行并发控制
	g, gctx := errgroup.WithContext(ctx)

	for _, key := range webhookKeys {
		k := key

		g.Go(func() error {
			err := baseClient.SendWithKey(gctx, k, msg)
			mu.Lock()
			results[k] = err
			mu.Unlock()
			return nil // 不返回错误，收集所有结果
		})
	}

	g.Wait()
	return results
}

// BatchSendStrict 批量发送消息到多个 webhook（严格模式）
// 任一发送失败立即取消其他发送并返回第一个错误
func BatchSendStrict(ctx context.Context, webhookKeys []string, msg Message, opts ...Option) error {
	// 创建单个 Client 并复用
	baseClient := New("", opts...)

	g, gctx := errgroup.WithContext(ctx)

	for _, key := range webhookKeys {
		k := key

		g.Go(func() error {
			return baseClient.SendWithKey(gctx, k, msg)
		})
	}

	return g.Wait()
}