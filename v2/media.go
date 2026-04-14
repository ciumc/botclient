package botclient

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// MediaType 媒体类型
type MediaType string

const (
	MediaTypeFile  MediaType = "file"
	MediaTypeImage MediaType = "image"
	MediaTypeVoice MediaType = "voice"
	MediaTypeVideo MediaType = "video"
)

// UploadRequest 上传请求
type UploadRequest struct {
	WebhookKey string        // Webhook Key (可选，如果客户端已设置则使用默认值)
	Type       MediaType     // 媒体类型: file, image, voice, video
	FilePath   string        // 文件路径 (与 Reader 二选一)
	Reader     io.ReadCloser // 文件 Reader (与 FilePath 二选一)，默认由 UploadMedia 关闭
	Filename   string        // 文件名 (使用 Reader 时需要指定)
	CloseReader bool         // 是否由 UploadMedia 关闭 Reader，默认 true
}

// Validate 验证上传请求
func (r *UploadRequest) Validate() error {
	if r.FilePath == "" && r.Reader == nil {
		return NewError(CodeInvalidMsgType, "file path or reader is required", ErrEmptyMessage)
	}
	if r.FilePath != "" && r.Reader != nil {
		return NewError(CodeInvalidMsgType, "only one of file path or reader should be provided", nil)
	}

	// 验证文件路径（防御性清理）
	if r.FilePath != "" {
		// 清理路径防止目录遍历攻击
		cleanPath := filepath.Clean(r.FilePath)

		// 检查文件是否存在（TOCTOU: 仍然存在风险，但大部分场景可接受）
		// 更安全的做法是直接打开文件并处理错误
		if _, err := os.Stat(cleanPath); err != nil {
			if os.IsNotExist(err) {
				return NewError(CodeInvalidMsgType, "file not found: "+cleanPath, err)
			}
			return NewError(CodeSystemError, "cannot access file: "+cleanPath, err)
		}

		// 更新为清理后的路径
		r.FilePath = cleanPath
	}

	return nil
}

// UploadResult 上传结果
type UploadResult struct {
	MediaId   string    `json:"media_id"`
	Type      MediaType `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

// UploadMedia 上传媒体文件
//
// 注意：如果提供了 Reader 且 CloseReader 为 true（默认），Reader 将在上传完成后被关闭
func (c *Client) UploadMedia(ctx context.Context, req *UploadRequest) (*UploadResult, error) {
	if req == nil {
		return nil, NewError(CodeInvalidMsgType, "upload request is nil", ErrEmptyMessage)
	}

	// 设置默认关闭行为
	if !req.CloseReader && req.Reader != nil {
		req.CloseReader = true
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	// 确保 Reader 在函数结束时关闭
	if req.Reader != nil && req.CloseReader {
		defer req.Reader.Close()
	}

	// 确定 webhook key（优先使用请求中的 key，否则使用客户端默认值）
	webhookKey := req.WebhookKey
	if webhookKey == "" {
		webhookKey = c.webhookKey
	}
	if webhookKey == "" {
		return nil, NewError(CodeInvalidKey, "webhook key is empty", ErrInvalidWebhookKey)
	}

	// 创建请求
	httpReq := c.httpClient.NewRequest()
	httpReq.SetContext(ctx)

	// 设置文件
	if req.FilePath != "" {
		httpReq.SetFile("media", req.FilePath)
	} else if req.Reader != nil {
		filename := req.Filename
		if filename == "" {
			filename = "file"
		}
		// 注意：resty 的 SetFileReader 不会关闭 Reader，所以我们用 defer 关闭
		httpReq.SetFileReader("media", filename, req.Reader)
	}

	// 设置响应结构
	var result struct {
		Type      string `json:"type"`
		MediaId   string `json:"media_id"`
		CreatedAt string `json:"created_at"`
		ErrCode   int    `json:"errcode"`
		ErrMsg    string `json:"errmsg"`
	}
	httpReq.SetResult(&result)

	// 发送请求
	url := fmt.Sprintf("%s/upload_media?key=%s&type=%s", c.baseURL, webhookKey, req.Type)
	resp, err := httpReq.Post(url)
	if err != nil {
		return nil, NewError(CodeSystemError, "http request failed", err)
	}

	if !resp.IsSuccess() {
		return nil, NewError(CodeSystemError, fmt.Sprintf("status code: %s", resp.Status()), ErrUploadFailed)
	}

	if result.ErrCode != 0 {
		return nil, NewError(ErrorCode(result.ErrCode), result.ErrMsg, ErrUploadFailed)
	}

	return &UploadResult{
		MediaId:   result.MediaId,
		Type:      MediaType(result.Type),
		CreatedAt: time.Now(),
	}, nil
}

// UploadFromPath 从路径上传文件
func (c *Client) UploadFromPath(ctx context.Context, filePath string, mediaType MediaType) (*UploadResult, error) {
	return c.UploadMedia(ctx, &UploadRequest{
		FilePath: filePath,
		Type:     mediaType,
	})
}

// UploadFromReader 从 Reader 上传
//
// 注意：reader 将在上传完成后被关闭
func (c *Client) UploadFromReader(ctx context.Context, filename string, reader io.ReadCloser, mediaType MediaType) (*UploadResult, error) {
	return c.UploadMedia(ctx, &UploadRequest{
		Reader:     reader,
		Filename:   filename,
		Type:       mediaType,
		CloseReader: true,
	})
}

// ImageFromFile 从文件创建图片消息
// 注意：此方法将整个文件读入内存，不适合处理大文件（建议限制在 2MB 以内）
func (c *Client) ImageFromFile(ctx context.Context, filePath string) (*ImageMessage, error) {
	// 防御性清理路径（调用方可能传入未清理的路径）
	cleanPath := filepath.Clean(filePath)

	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, NewError(CodeSystemError, "read file failed", err)
	}

	hash := md5.Sum(data)
	md5Str := hex.EncodeToString(hash[:])
	base64Str := base64.StdEncoding.EncodeToString(data)

	return &ImageMessage{
		Base64: base64Str,
		Md5:    md5Str,
	}, nil
}

// SendFile 发送文件消息（上传并发送）
func (c *Client) SendFile(ctx context.Context, filePath string) error {
	result, err := c.UploadFromPath(ctx, filePath, MediaTypeFile)
	if err != nil {
		return err
	}
	return c.Send(ctx, File(result.MediaId))
}

// SendImage 发送图片消息（上传并发送）
func (c *Client) SendImage(ctx context.Context, filePath string) error {
	msg, err := c.ImageFromFile(ctx, filePath)
	if err != nil {
		return err
	}
	return c.Send(ctx, msg)
}