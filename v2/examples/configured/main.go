package main

import (
	"context"
	"log"
	"os"
	"time"

	botclient "github.com/ciumc/botclient/v2"
)

// 自定义日志记录器
type MyLogger struct{}

func (l *MyLogger) Info(ctx context.Context, msg string, args ...any) {
	log.Printf("[INFO] %s %v", msg, args)
}

func (l *MyLogger) Error(ctx context.Context, msg string, args ...any) {
	log.Printf("[ERROR] %s %v", msg, args)
}

func (l *MyLogger) Debug(ctx context.Context, msg string, args ...any) {
	log.Printf("[DEBUG] %s %v", msg, args)
}

func main() {
	webhookKey := os.Getenv("WECOM_WEBHOOK_KEY")
	if webhookKey == "" {
		log.Fatal("请设置环境变量 WECOM_WEBHOOK_KEY")
	}

	ctx := context.Background()

	// 方式2: Builder 模式配置客户端
	client := botclient.NewBuilder(webhookKey).
		WithTimeout(10 * time.Second).
		WithRetry(3, time.Second).
		WithLogger(&MyLogger{}).
		Build()

	// 发送消息
	err := client.Send(ctx, botclient.Text("Hello with retry and logging!"))
	if err != nil {
		log.Printf("发送失败: %v", err)
	}

	// 使用不同的 webhook key 发送
	err = client.SendWithKey(ctx, "another-webhook-key", botclient.Text("Hello to another group!"))
	if err != nil {
		log.Printf("发送到其他群失败: %v", err)
	}

	// 批量发送到多个群
	keys := []string{"key1", "key2", "key3"}
	results := botclient.BatchSend(ctx, keys, botclient.Text("批量消息"))
	for key, err := range results {
		if err != nil {
			log.Printf("发送到 %s 失败: %v", key, err)
		} else {
			log.Printf("发送到 %s 成功", key)
		}
	}

	log.Println("消息发送完成")
}