package botclient

import (
	"errors"
	"fmt"
)

// 基础错误定义
var (
	ErrInvalidWebhookKey = errors.New("invalid webhook key")
	ErrEmptyMessage      = errors.New("message cannot be empty")
	ErrUnsupportedType   = errors.New("unsupported message type")
	ErrUploadFailed      = errors.New("media upload failed")
	ErrRequestFailed     = errors.New("request failed")
	ErrRateLimited       = errors.New("rate limited")
	ErrTimeout           = errors.New("request timeout")
	ErrInvalidMsgType    = errors.New("invalid message type")
)

// ErrorCode 错误码类型
// 企业微信 API 错误码参考: https://developer.work.weixin.qq.com/document/path/91770
type ErrorCode int

const (
	// 本地错误码 (50000+)
	CodeSuccess     ErrorCode = 0      // 成功
	CodeSystemError ErrorCode = 50001  // 系统错误（本地定义）

	// 企业微信 API 错误码
	CodeInvalidKey     ErrorCode = 40001  // 无效 webhook key
	CodeInvalidMsgType ErrorCode = 40002  // 无效消息类型/内容
	CodeMsgTooLong     ErrorCode = 40013  // 消息内容过长
	CodeRateLimited    ErrorCode = 45009  // 接口频率限制
)

// Error 自定义错误类型
type Error struct {
	Code    ErrorCode // 错误码
	Message string    // 错误消息
	Err     error     // 原始错误
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("botclient: [%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("botclient: [%d] %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Err
}

// NewError 创建错误
func NewError(code ErrorCode, message string, err error) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// IsErrorType 判断错误类型
func IsErrorType(err error, code ErrorCode) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.Code == code
	}
	return false
}

// IsRetryable 判断是否可重试
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	var e *Error
	if errors.As(err, &e) {
		switch e.Code {
		case CodeRateLimited, CodeSystemError:
			return true
		}
	}
	// 网络错误可重试
	return errors.Is(err, ErrRequestFailed) || errors.Is(err, ErrTimeout)
}

// GetErrorCode 获取错误码
func GetErrorCode(err error) ErrorCode {
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return CodeSystemError
}