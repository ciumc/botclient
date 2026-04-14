# botclient/v2

企业微信机器人客户端 v2 版本，支持 Context、Builder 模式、中间件、模板卡片消息等功能。

## 安装

```bash
go get github.com/ciumc/botclient/v2
```

## 特性

- ✅ **Context 支持** - 所有操作支持 context，可取消、超时
- ✅ **Builder 模式** - 消息构建更流畅，链式调用
- ✅ **Option 模式** - 客户端配置灵活
- ✅ **错误处理** - 自定义错误类型，支持错误码
- ✅ **模板卡片消息** - 完整实现文本通知、新闻通知、按钮交互等
- ✅ **中间件支持** - 日志、重试、指标等
- ✅ **批量发送** - 使用 errgroup 并发发送，支持严格模式

## 快速开始

### 基础使用

```go
package main

import (
    "context"
    "log"
    
    botclient "github.com/ciumc/botclient/v2"
)

func main() {
    ctx := context.Background()
    
    // 创建客户端
    client := botclient.New("your-webhook-key")
    
    // 发送文本消息
    err := client.Send(ctx, botclient.Text("Hello, World!"))
    if err != nil {
        log.Fatal(err)
    }
    
    // 发送带提及的消息
    err = client.Send(ctx, botclient.Text("@all 请注意").
        MentionAll().
        Mention("user1", "user2"))
    
    // 发送 Markdown 消息
    err = client.Send(ctx, botclient.Markdown("# 标题\n内容"))
    
    // 发送图文消息
    err = client.Send(ctx, botclient.News().
        AddArticle("标题", "https://example.com").
        AddArticleWithPic("标题", "描述", "url", "picUrl"))
}
```

### 高级配置

```go
import (
    "context"
    "time"
    
    botclient "github.com/ciumc/botclient/v2"
)

// Builder 模式配置客户端
client := botclient.NewBuilder("webhook-key").
    WithTimeout(10 * time.Second).
    WithRetry(3, time.Second).
    WithLogger(myLogger).
    Build()

// 使用不同的 webhook key 发送
client.SendWithKey(ctx, "another-key", botclient.Text("hello"))

// 批量发送（并发执行，收集所有结果）
keys := []string{"key1", "key2", "key3"}
results := botclient.BatchSend(ctx, keys, botclient.Text("批量消息"))
for key, err := range results {
    if err != nil {
        log.Printf("发送到 %s 失败: %v", key, err)
    }
}

// 严格模式批量发送（任一失败立即取消其他）
err := botclient.BatchSendStrict(ctx, keys, botclient.Text("严格批量"))
if err != nil {
    log.Printf("批量发送失败: %v", err)
}
```

## 消息类型

### 文本消息

```go
// 简单文本
botclient.Text("Hello")

// 带提及
botclient.Text("Hello").Mention("user1", "user2")
botclient.Text("Hello").MentionAll()
botclient.Text("Hello").MentionMobile("13800138000")
```

### Markdown 消息

```go
botclient.Markdown("# 标题\n**粗体** *斜体*\n[链接](url)")
```

### 图文消息

```go
botclient.News().
    AddArticle("标题", "url").
    AddArticleWithDesc("标题", "描述", "url").
    AddArticleWithPic("标题", "描述", "url", "picUrl")
```

### 文件消息

```go
// 上传文件
result, _ := client.UploadFromPath(ctx, "./file.pdf", botclient.MediaTypeFile)

// 发送文件消息
client.Send(ctx, botclient.File(result.MediaId))

// 或直接上传并发送
client.SendFile(ctx, "./file.pdf")
```

### 图片消息

```go
// 从文件创建
msg, _ := client.ImageFromFile(ctx, "./image.png")
client.Send(ctx, msg)

// 或直接发送
client.SendImage(ctx, "./image.png")
```

### 模板卡片消息

```go
// 文本通知
card := botclient.TextNotice().
    Source("iconUrl", "来源").
    MainTitle("标题", "描述").
    Emphasis("95%", "使用率").
    AddContentItem("服务器", "web-01").
    AddJump("详情", "url").
    Build()

// 新闻通知
card := botclient.NewsNotice().
    Source("iconUrl", "新闻来源").
    MainTitle("标题", "描述").
    AddJump("查看", "url").
    CardAction("url").
    Build()

// 按钮交互
card := botclient.ButtonInteraction().
    Source("iconUrl", "审批系统").
    MainTitle("审批请求", "张三提交了请假申请").
    TaskID("task-001").
    Build()

client.Send(ctx, card)
```

## 错误处理

```go
err := client.Send(ctx, msg)
if err != nil {
    // 检查错误类型
    if botclient.IsErrorType(err, botclient.CodeRateLimited) {
        // 限流，等待重试
        time.Sleep(time.Second * 10)
    }
    
    // 检查是否可重试
    if botclient.IsRetryable(err) {
        // 自动重试
    }
}
```

## 自定义中间件

```go
func MyMiddleware() botclient.Middleware {
    return func(next botclient.Handler) botclient.Handler {
        return func(ctx context.Context, req *botclient.Request) (*botclient.Response, error) {
            // 前置处理
            resp, err := next(ctx, req)
            // 后置处理
            return resp, err
        }
    }
}

client := botclient.NewBuilder("key").
    WithMiddleware(MyMiddleware()).
    Build()
```

## 迁移指南

从 v1 迁移到 v2：

| v1 | v2 |
|----|----|
| `bot.Send(msg)` | `bot.Send(ctx, msg)` |
| `&TextMessage{Content: "..."}` | `Text("...")` |
| `bot.UploadMedia(path)` | `bot.UploadFromPath(ctx, path, type)` |

## License

MIT