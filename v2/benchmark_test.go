package botclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// 基准测试：发送文本消息
func BenchmarkClient_Send_Text(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer server.Close()

	client := New("test-key", WithBaseURL(server.URL))
	msg := Text("benchmark test message")

	b.ResetTimer()
	for b.Loop() {
		client.Send(context.Background(), msg)
	}
}

// 基准测试：发送 Markdown 消息
func BenchmarkClient_Send_Markdown(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer server.Close()

	client := New("test-key", WithBaseURL(server.URL))
	msg := Markdown("# Title\n**bold** *italic*\n> quote")

	b.ResetTimer()
	for b.Loop() {
		client.Send(context.Background(), msg)
	}
}

// 基准测试：发送模板卡片消息
func BenchmarkClient_Send_TemplateCard(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer server.Close()

	client := New("test-key", WithBaseURL(server.URL))
	msg := TextNotice().
		MainTitle("Alert", "CPU High").
		Emphasis("95%", "Usage").
		AddContentItem("Server", "web-01").
		Build()

	b.ResetTimer()
	for b.Loop() {
		client.Send(context.Background(), msg)
	}
}

// 基准测试：消息构建
func BenchmarkMessageBuilder_Text(b *testing.B) {
	for b.Loop() {
		Text("hello world").MentionAll().Mention("user1", "user2")
	}
}

func BenchmarkMessageBuilder_News(b *testing.B) {
	for b.Loop() {
		News().
			AddArticle("Title1", "url1").
			AddArticleWithPic("Title2", "desc", "url2", "pic")
	}
}

func BenchmarkMessageBuilder_TemplateCard(b *testing.B) {
	for b.Loop() {
		TextNotice().
			Source("icon", "System").
			MainTitle("Title", "Desc").
			AddContentItem("Key", "Value").
			AddJump("Link", "url").
			Build()
	}
}

// 基准测试：客户端创建
func BenchmarkClient_New(b *testing.B) {
	for b.Loop() {
		New("test-key")
	}
}

func BenchmarkClient_NewBuilder(b *testing.B) {
	for b.Loop() {
		NewBuilder("test-key").
			WithTimeout(10 * time.Second).
			WithRetry(3, time.Second).
			Build()
	}
}

// 基准测试：批量发送并发
func BenchmarkBatchSend_Parallel(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer server.Close()

	keys := []string{"key1", "key2", "key3", "key4", "key5"}
	msg := Text("parallel benchmark")

	b.ResetTimer()
	for b.Loop() {
		BatchSend(context.Background(), keys, msg, WithBaseURL(server.URL))
	}
}

// 基准测试：并发发送
func BenchmarkClient_Send_Parallel(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer server.Close()

	client := New("test-key", WithBaseURL(server.URL))
	msg := Text("parallel test")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			client.Send(context.Background(), msg)
		}
	})
}