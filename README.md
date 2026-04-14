# botclient

企业微信机器人客户端，用于通过 Webhook 发送消息到企业微信群。

> **注意**: 本版本 (v1) 是简化版本，推荐使用 [v2 版本](./v2) 获得更多功能。

## 安装

```bash
go get github.com/ciumc/botclient
```

## 功能

- ✅ 文本消息发送
- ✅ Markdown 消息发送
- ✅ 图片消息发送
- ✅ 图文消息发送
- ✅ 文件消息发送
- ✅ 媒体文件上传

## 快速开始

### 初始化客户端

```go
import "github.com/ciumc/botclient"

// 使用 Webhook Key 创建客户端
bot := botclient.New("your-webhook-key")
```

### 发送文本消息

```go
// 简单文本
msg := &botclient.TextMessage{
    Content: "Hello, World!",
}
bot.Send(msg)

// 带提及的文本
msg := &botclient.TextMessage{
    Content:       "@all 请注意",
    MentionedList: []string{botclient.All},
}
bot.Send(msg)
```

### 发送 Markdown 消息

```go
msg := &botclient.MarkdownMessage{
    Content: "# 标题\n**粗体** *斜体*\n> 引用文本",
}
bot.Send(msg)
```

### 发送图片消息

```go
// 读取图片文件
body, err := os.ReadFile("image.png")
if err != nil {
    panic(err)
}

// 计算 MD5 和 Base64
hash := md5.New()
hash.Write(body)
md5Str := hex.EncodeToString(hash.Sum(nil))
base64Str := base64.StdEncoding.EncodeToString(body)

msg := &botclient.ImageMessage{
    Base64: base64Str,
    Md5:    md5Str,
}
bot.Send(msg)
```

### 发送图文消息

```go
msg := &botclient.NewsMessage{}
msg.AddArticle(botclient.NewsMessageArticle{
    Title:       "文章标题",
    Description: "文章描述",
    URL:         "https://example.com",
    PicURL:      "https://example.com/pic.png",
})
bot.Send(msg)
```

### 发送文件消息

```go
// 先上传文件获取 media_id
mediaId, err := bot.UploadMedia("file.pdf")
if err != nil {
    panic(err)
}

// 发送文件消息
msg := &botclient.FileMessage{
    MediaId: mediaId,
}
bot.Send(msg)
```

## 消息类型

| 类型 | 结构体 | 说明 |
|------|--------|------|
| 文本 | `TextMessage` | 支持提及用户/手机号 |
| Markdown | `MarkdownMessage` | 支持基本 Markdown 语法 |
| 图片 | `ImageMessage` | 需要 Base64 和 MD5 |
| 图文 | `NewsMessage` | 最多 8 条文章 |
| 文件 | `FileMessage` | 需要先上传获取 media_id |

## 常量

```go
const (
    Text     = "text"
    Markdown = "markdown"
    Image    = "image"
    News     = "news"
    File     = "file"
    All      = "@all"  // 提及所有人
)
```

## 与 v2 版本的区别

| 特性 | v1 | v2 |
|------|----|----|
| Context 支持 | ❌ | ✅ |
| Builder 模式 | ❌ | ✅ |
| Option 配置 | ❌ | ✅ |
| 中间件支持 | ❌ | ✅ |
| 模板卡片 | ❌ | ✅ |
| 批量发送 | ❌ | ✅ |
| 错误码 | ❌ | ✅ |

推荐升级到 [v2 版本](./v2) 以获得更好的功能支持：

```bash
go get github.com/ciumc/botclient/v2
```

## 企业微信 API 文档

- [群机器人发送消息](https://developer.work.weixin.qq.com/document/path/91770)

## License

MIT