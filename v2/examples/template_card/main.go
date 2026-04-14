package main

import (
	"context"
	"log"
	"os"
	"time"

	botclient "github.com/ciumc/botclient/v2"
)

func main() {
	webhookKey := os.Getenv("WECOM_WEBHOOK_KEY")
	if webhookKey == "" {
		log.Fatal("请设置环境变量 WECOM_WEBHOOK_KEY")
	}

	ctx := context.Background()
	client := botclient.New(webhookKey)

	// 文本通知模板卡片
	textCard := botclient.TextNotice().
		Source("https://cdn.example.com/icon.png", "系统监控").
		MainTitle("服务器告警", "CPU 使用率异常").
		Emphasis("95%", "当前使用率").
		AddContentItem("服务器", "web-01").
		AddContentItem("IP 地址", "192.168.1.100").
		AddContentItem("时间", time.Now().Format("2006-01-02 15:04:05")).
		AddJump("查看详情", "https://monitor.example.com/alert/123").
		AddJump("处理告警", "https://monitor.example.com/handle").
		Build()

	err := client.Send(ctx, textCard)
	if err != nil {
		log.Printf("发送文本通知卡片失败: %v", err)
	}

	// 新闻通知模板卡片
	newsCard := botclient.NewsNotice().
		Source("https://cdn.example.com/news-icon.png", "公司新闻").
		MainTitle("重要通知", "2024年度总结").
		SubTitle("请各部门及时提交年度报告").
		AddJump("查看详情", "https://internal.example.com/news/2024").
		CardAction("https://internal.example.com/news").
		Build()

	err = client.Send(ctx, newsCard)
	if err != nil {
		log.Printf("发送新闻通知卡片失败: %v", err)
	}

	// 按钮交互模板卡片
	buttonCard := botclient.ButtonInteraction().
		Source("https://cdn.example.com/app-icon.png", "审批系统").
		MainTitle("审批请求", "张三提交了请假申请").
		SubTitle("请假类型：年假，请假天数：3天").
		AddContentItem("申请时间", time.Now().Format("2006-01-02 15:04:05")).
		TaskID("leave-approval-001").
		Build()

	err = client.Send(ctx, buttonCard)
	if err != nil {
		log.Printf("发送按钮交互卡片失败: %v", err)
	}

	log.Println("模板卡片发送完成")
}