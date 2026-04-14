package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	botclient "github.com/ciumc/botclient/v2"
)

func main() {
	// 从环境变量获取 webhook key
	webhookKey := os.Getenv("WECOM_WEBHOOK_KEY")
	if webhookKey == "" {
		log.Fatal("请设置环境变量 WECOM_WEBHOOK_KEY")
	}

	ctx := context.Background()

	// 方式1: 简单初始化
	client := botclient.New(webhookKey)

	// 发送文本消息
	err := client.Send(ctx, botclient.Text("Hello, World!"))
	if err != nil {
		log.Printf("发送文本消息失败: %v", err)
	}

	// 发送带提及的文本消息
	err = client.Send(ctx, botclient.Text("@all 请注意").
		MentionAll().
		Mention("user1", "user2"))
	if err != nil {
		log.Printf("发送提及消息失败: %v", err)
	}

	// 发送 Markdown 消息
	err = client.Send(ctx, botclient.Markdown(`
# 标题
## 二级标题
> 引用文本
**粗体** *斜体*
[链接](https://example.com)
`))
	if err != nil {
		log.Printf("发送 Markdown 消息失败: %v", err)
	}

	// 发送图文消息
	err = client.Send(ctx, botclient.News().
		AddArticleWithPic("标题", "描述", "https://example.com", "https://example.com/pic.png").
		AddArticle("标题2", "https://example.com/2"))
	if err != nil {
		log.Printf("发送图文消息失败: %v", err)
	}

	// 发送模板卡片消息
	card := botclient.TextNotice().
		Source("https://example.com/icon.png", "系统通知").
		MainTitle("服务器告警", "CPU 使用率过高").
		Emphasis("95%", "当前使用率").
		AddContentItem("服务器", "web-01").
		AddContentItem("时间", time.Now().Format("2006-01-02 15:04:05")).
		AddJump("查看详情", "https://monitor.example.com/alert/123").
		Build()

	err = client.Send(ctx, card)
	if err != nil {
		log.Printf("发送模板卡片消息失败: %v", err)
	}

	fmt.Println("消息发送完成")
}