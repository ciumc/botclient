package botclient

import (
	"fmt"
	"slices"
)

// 消息类型常量
const (
	MsgTypeText         = "text"
	MsgTypeMarkdown     = "markdown"
	MsgTypeImage        = "image"
	MsgTypeNews         = "news"
	MsgTypeFile         = "file"
	MsgTypeTemplateCard = "template_card"
)

// supportedMessageTypes 支持的消息类型列表
var supportedMessageTypes = []string{
	MsgTypeText, MsgTypeMarkdown, MsgTypeImage,
	MsgTypeNews, MsgTypeFile, MsgTypeTemplateCard,
}

// Message 消息接口
type Message interface {
	// Type 返回消息类型
	Type() string
	// Validate 验证消息有效性
	Validate() error
}

// --- 文本消息 ---

// TextMessage 文本消息
type TextMessage struct {
	Content             string   `json:"content"`
	MentionedList       []string `json:"mentioned_list,omitempty"`
	MentionedMobileList []string `json:"mentioned_mobile_list,omitempty"`
}

func (m *TextMessage) Type() string { return MsgTypeText }

func (m *TextMessage) Validate() error {
	if m.Content == "" {
		return NewError(CodeInvalidMsgType, "text content is empty", ErrEmptyMessage)
	}
	if len(m.Content) > 2048 {
		return NewError(CodeMsgTooLong, "text content exceeds 2048 bytes", nil)
	}
	return nil
}

// Text 创建文本消息 (Builder 辅助函数)
func Text(content string) *TextMessage {
	return &TextMessage{Content: content}
}

// Mention 添加提及用户
func (m *TextMessage) Mention(users ...string) *TextMessage {
	m.MentionedList = append(m.MentionedList, users...)
	return m
}

// MentionMobile 添加提及手机号
func (m *TextMessage) MentionMobile(mobiles ...string) *TextMessage {
	m.MentionedMobileList = append(m.MentionedMobileList, mobiles...)
	return m
}

// MentionAll 提及所有人
func (m *TextMessage) MentionAll() *TextMessage {
	return m.Mention("@all")
}

// --- Markdown 消息 ---

// MarkdownMessage Markdown 消息
type MarkdownMessage struct {
	Content string `json:"content"`
}

func (m *MarkdownMessage) Type() string { return MsgTypeMarkdown }

func (m *MarkdownMessage) Validate() error {
	if m.Content == "" {
		return NewError(CodeInvalidMsgType, "markdown content is empty", ErrEmptyMessage)
	}
	if len(m.Content) > 4096 {
		return NewError(CodeMsgTooLong, "markdown content exceeds 4096 bytes", nil)
	}
	return nil
}

// Markdown 创建 Markdown 消息
func Markdown(content string) *MarkdownMessage {
	return &MarkdownMessage{Content: content}
}

// --- 图片消息 ---

// ImageMessage 图片消息
type ImageMessage struct {
	Base64 string `json:"base64"`
	Md5    string `json:"md5"`
}

func (m *ImageMessage) Type() string { return MsgTypeImage }

func (m *ImageMessage) Validate() error {
	if m.Base64 == "" || m.Md5 == "" {
		return NewError(CodeInvalidMsgType, "image base64 or md5 is empty", ErrEmptyMessage)
	}
	return nil
}

// Image 创建图片消息
func Image(base64, md5 string) *ImageMessage {
	return &ImageMessage{Base64: base64, Md5: md5}
}

// --- 图文消息 ---

// NewsMessage 图文消息
type NewsMessage struct {
	Articles []NewsArticle `json:"articles"`
}

// NewsArticle 图文文章
type NewsArticle struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url"`
	PicURL      string `json:"picurl,omitempty"`
}

func (m *NewsMessage) Type() string { return MsgTypeNews }

func (m *NewsMessage) Validate() error {
	if len(m.Articles) == 0 {
		return NewError(CodeInvalidMsgType, "news articles is empty", ErrEmptyMessage)
	}
	if len(m.Articles) > 8 {
		return NewError(CodeMsgTooLong, "news articles exceeds 8", nil)
	}
	for i, a := range m.Articles {
		if a.Title == "" || a.URL == "" {
			return NewError(CodeInvalidMsgType,
				fmt.Sprintf("article %d: title or url is empty", i), nil)
		}
	}
	return nil
}

// News 创建图文消息
func News() *NewsMessage {
	return &NewsMessage{Articles: make([]NewsArticle, 0)}
}

// AddArticle 添加文章
func (m *NewsMessage) AddArticle(title, url string) *NewsMessage {
	m.Articles = append(m.Articles, NewsArticle{Title: title, URL: url})
	return m
}

// AddArticleWithDesc 添加带描述的文章
func (m *NewsMessage) AddArticleWithDesc(title, desc, url string) *NewsMessage {
	m.Articles = append(m.Articles, NewsArticle{Title: title, Description: desc, URL: url})
	return m
}

// AddArticleWithPic 添加带图片的文章
func (m *NewsMessage) AddArticleWithPic(title, desc, url, picURL string) *NewsMessage {
	m.Articles = append(m.Articles, NewsArticle{
		Title: title, Description: desc, URL: url, PicURL: picURL,
	})
	return m
}

// --- 文件消息 ---

// FileMessage 文件消息
type FileMessage struct {
	MediaId string `json:"media_id"`
}

func (m *FileMessage) Type() string { return MsgTypeFile }

func (m *FileMessage) Validate() error {
	if m.MediaId == "" {
		return NewError(CodeInvalidMsgType, "file media_id is empty", ErrEmptyMessage)
	}
	return nil
}

// File 创建文件消息
func File(mediaId string) *FileMessage {
	return &FileMessage{MediaId: mediaId}
}

// --- 辅助函数 ---

// ValidateMessage 验证消息 (用于外部调用)
func ValidateMessage(msg Message) error {
	if msg == nil {
		return NewError(CodeInvalidMsgType, "message is nil", ErrEmptyMessage)
	}
	return msg.Validate()
}

// IsMessageTypeSupported 判断消息类型是否支持
func IsMessageTypeSupported(msgType string) bool {
	return slices.Contains(supportedMessageTypes, msgType)
}