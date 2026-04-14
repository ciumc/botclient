package botclient

// 模板卡片消息类型
const (
	TemplateTypeTextNotice         = "text_notice"
	TemplateTypeNewsNotice         = "news_notice"
	TemplateTypeButtonInteraction  = "button_interaction"
	TemplateTypeVoteInteraction    = "vote_interaction"
	TemplateTypeMultipleInteraction = "multiple_interaction"
)

// JumpType 跳转类型常量
const (
	JumpTypeNone    = 0 // 不跳转
	JumpTypeURL     = 1 // 跳转 URL
	JumpTypeMiniApp = 2 // 跳转小程序
)

// HorizontalContentType 水平内容类型常量
const (
	HorizontalContentText = 0 // 文本
	HorizontalContentLink = 1 // 链接
)

// ImageTextAreaType 图片文本区域类型常量
const (
	ImageTextAreaTypeImage = 0 // 图片
	ImageTextAreaTypeURL   = 1 // 网页
)

// CardActionType 卡片动作类型常量
const (
	CardActionNone    = 0 // 不跳转
	CardActionURL     = 1 // 跳转 URL
	CardActionMiniApp = 2 // 跳转小程序
)

// TemplateCardMessage 模板卡片消息
type TemplateCardMessage struct {
	MsgType      string           `json:"msgtype"`
	TemplateCard TemplateCardBody `json:"template_card"`
}

// TemplateCardBody 模板卡片主体
type TemplateCardBody struct {
	CardType             string                     `json:"card_type"`
	Source               *TemplateSource            `json:"source,omitempty"`
	MainTitle            *TemplateMainTitle         `json:"main_title,omitempty"`
	SubTitleText         string                     `json:"sub_title_text,omitempty"`
	EmphasisContent      *TemplateEmphasis          `json:"emphasis_content,omitempty"`
	ContentItems         []TemplateContentItem      `json:"content_items,omitempty"`
	ImageTextArea        *TemplateImageTextArea     `json:"image_text_area,omitempty"`
	VerticalContentList  []TemplateVerticalContent  `json:"vertical_content_list,omitempty"`
	HorizontalContentList []TemplateHorizontalContent `json:"horizontal_content_list,omitempty"`
	JumpList             []TemplateJump             `json:"jump_list,omitempty"`
	CardAction           *TemplateCardAction        `json:"card_action,omitempty"`
	ButtonSelection      *TemplateButtonSelection   `json:"button_selection,omitempty"`
	TaskID               string                     `json:"task_id,omitempty"`
	CallbackID           string                     `json:"callback_id,omitempty"`
}

// TemplateSource 来源信息
type TemplateSource struct {
	IconURL   string `json:"icon_url,omitempty"`
	Desc      string `json:"desc,omitempty"`
	DescColor int    `json:"desc_color,omitempty"` // 0 灰色 1 黑色
}

// TemplateMainTitle 主标题
type TemplateMainTitle struct {
	Title string `json:"title"`
	Desc  string `json:"desc,omitempty"`
}

// TemplateEmphasis 强调内容
type TemplateEmphasis struct {
	Title string `json:"title"`
	Desc  string `json:"desc,omitempty"`
}

// TemplateContentItem 内容项
type TemplateContentItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// TemplateImageTextArea 图片文本区域
type TemplateImageTextArea struct {
	Type     int    `json:"type"` // 0 图片 1 网页
	URL      string `json:"url,omitempty"`
	Title    string `json:"title,omitempty"`
	Desc     string `json:"desc,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
}

// TemplateVerticalContent 垂直内容
type TemplateVerticalContent struct {
	Title   string `json:"title"`
	Desc    string `json:"desc,omitempty"`
	URL     string `json:"url,omitempty"`
	MediaID string `json:"media_id,omitempty"`
}

// TemplateHorizontalContent 水平内容
type TemplateHorizontalContent struct {
	Key     int    `json:"key"` // 0 文本 1 链接
	Title   string `json:"title"`
	URL     string `json:"url,omitempty"`
	MediaID string `json:"media_id,omitempty"`
}

// TemplateJump 跳转链接
type TemplateJump struct {
	Type     int    `json:"type"` // 0 不跳转 1 跳转URL 2 跳转小程序
	Title    string `json:"title"`
	URL      string `json:"url,omitempty"`
	AppID    string `json:"appid,omitempty"`
	PagePath string `json:"pagepath,omitempty"`
}

// TemplateCardAction 卡片点击动作
type TemplateCardAction struct {
	Type     int    `json:"type"` // 0 不跳转 1 跳转URL 2 跳转小程序
	URL      string `json:"url,omitempty"`
	AppID    string `json:"appid,omitempty"`
	PagePath string `json:"pagepath,omitempty"`
}

// TemplateButtonSelection 按钮选择
type TemplateButtonSelection struct {
	TaskID     string              `json:"task_id"`
	Question   string              `json:"question"`
	Selection  []TemplateSelection `json:"selection"`
	SelectedID string              `json:"selected_id,omitempty"`
}

// TemplateSelection 选择项
type TemplateSelection struct {
	ID    string `json:"id"`
	Text  string `json:"text"`
	Query string `json:"query,omitempty"`
}

func (m *TemplateCardMessage) Type() string { return MsgTypeTemplateCard }

func (m *TemplateCardMessage) Validate() error {
	if m.TemplateCard.CardType == "" {
		return NewError(CodeInvalidMsgType, "template card type is empty", ErrEmptyMessage)
	}
	return nil
}

// --- Builder 模式 ---

// TemplateCardBuilder 模板卡片构建器
type TemplateCardBuilder struct {
	card *TemplateCardBody
}

// TextNotice 文本通知模板卡片
func TextNotice() *TemplateCardBuilder {
	return &TemplateCardBuilder{
		card: &TemplateCardBody{
			CardType: TemplateTypeTextNotice,
		},
	}
}

// NewsNotice 新闻通知模板卡片
func NewsNotice() *TemplateCardBuilder {
	return &TemplateCardBuilder{
		card: &TemplateCardBody{
			CardType: TemplateTypeNewsNotice,
		},
	}
}

// ButtonInteraction 按钮交互模板卡片
func ButtonInteraction() *TemplateCardBuilder {
	return &TemplateCardBuilder{
		card: &TemplateCardBody{
			CardType: TemplateTypeButtonInteraction,
		},
	}
}

// Source 设置来源
func (b *TemplateCardBuilder) Source(iconURL, desc string) *TemplateCardBuilder {
	b.card.Source = &TemplateSource{
		IconURL: iconURL,
		Desc:    desc,
	}
	return b
}

// SourceWithColor 设置来源（带颜色）
func (b *TemplateCardBuilder) SourceWithColor(iconURL, desc string, color int) *TemplateCardBuilder {
	b.card.Source = &TemplateSource{
		IconURL:   iconURL,
		Desc:      desc,
		DescColor: color,
	}
	return b
}

// MainTitle 设置主标题
func (b *TemplateCardBuilder) MainTitle(title, desc string) *TemplateCardBuilder {
	b.card.MainTitle = &TemplateMainTitle{
		Title: title,
		Desc:  desc,
	}
	return b
}

// SubTitle 设置副标题
func (b *TemplateCardBuilder) SubTitle(text string) *TemplateCardBuilder {
	b.card.SubTitleText = text
	return b
}

// Emphasis 设置强调内容
func (b *TemplateCardBuilder) Emphasis(title, desc string) *TemplateCardBuilder {
	b.card.EmphasisContent = &TemplateEmphasis{
		Title: title,
		Desc:  desc,
	}
	return b
}

// AddContentItem 添加内容项
func (b *TemplateCardBuilder) AddContentItem(key, value string) *TemplateCardBuilder {
	b.card.ContentItems = append(b.card.ContentItems, TemplateContentItem{
		Key:   key,
		Value: value,
	})
	return b
}

// ImageTextArea 设置图片文本区域
func (b *TemplateCardBuilder) ImageTextArea(imgURL, title, desc string) *TemplateCardBuilder {
	b.card.ImageTextArea = &TemplateImageTextArea{
		Type:     ImageTextAreaTypeImage,
		ImageURL: imgURL,
		Title:    title,
		Desc:     desc,
	}
	return b
}

// ImageTextAreaWithURL 设置图片文本区域（带链接）
func (b *TemplateCardBuilder) ImageTextAreaWithURL(url, title, desc string) *TemplateCardBuilder {
	b.card.ImageTextArea = &TemplateImageTextArea{
		Type:  ImageTextAreaTypeURL,
		URL:   url,
		Title: title,
		Desc:  desc,
	}
	return b
}

// AddVerticalContent 添加垂直内容
func (b *TemplateCardBuilder) AddVerticalContent(title, desc string) *TemplateCardBuilder {
	b.card.VerticalContentList = append(b.card.VerticalContentList, TemplateVerticalContent{
		Title: title,
		Desc:  desc,
	})
	return b
}

// AddHorizontalContentText 添加水平文本内容
func (b *TemplateCardBuilder) AddHorizontalContentText(title string) *TemplateCardBuilder {
	b.card.HorizontalContentList = append(b.card.HorizontalContentList, TemplateHorizontalContent{
		Key:   HorizontalContentText,
		Title: title,
	})
	return b
}

// AddHorizontalContentLink 添加水平链接内容
func (b *TemplateCardBuilder) AddHorizontalContentLink(title, url string) *TemplateCardBuilder {
	b.card.HorizontalContentList = append(b.card.HorizontalContentList, TemplateHorizontalContent{
		Key:   HorizontalContentLink,
		Title: title,
		URL:   url,
	})
	return b
}

// AddJump 添加跳转链接
func (b *TemplateCardBuilder) AddJump(title, url string) *TemplateCardBuilder {
	b.card.JumpList = append(b.card.JumpList, TemplateJump{
		Type:  JumpTypeURL,
		Title: title,
		URL:   url,
	})
	return b
}

// AddJumpMiniApp 添加小程序跳转
func (b *TemplateCardBuilder) AddJumpMiniApp(title, appID, pagePath string) *TemplateCardBuilder {
	b.card.JumpList = append(b.card.JumpList, TemplateJump{
		Type:     JumpTypeMiniApp,
		Title:    title,
		AppID:    appID,
		PagePath: pagePath,
	})
	return b
}

// CardAction 设置卡片点击动作
func (b *TemplateCardBuilder) CardAction(url string) *TemplateCardBuilder {
	b.card.CardAction = &TemplateCardAction{
		Type: CardActionURL,
		URL:  url,
	}
	return b
}

// CardActionMiniApp 设置小程序卡片点击动作
func (b *TemplateCardBuilder) CardActionMiniApp(appID, pagePath string) *TemplateCardBuilder {
	b.card.CardAction = &TemplateCardAction{
		Type:     CardActionMiniApp,
		AppID:    appID,
		PagePath: pagePath,
	}
	return b
}

// TaskID 设置任务 ID
func (b *TemplateCardBuilder) TaskID(id string) *TemplateCardBuilder {
	b.card.TaskID = id
	return b
}

// CallbackID 设置回调 ID
func (b *TemplateCardBuilder) CallbackID(id string) *TemplateCardBuilder {
	b.card.CallbackID = id
	return b
}

// Build 构建模板卡片消息
func (b *TemplateCardBuilder) Build() *TemplateCardMessage {
	return &TemplateCardMessage{
		MsgType:      MsgTypeTemplateCard,
		TemplateCard: *b.card,
	}
}