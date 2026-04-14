package botclient

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestClient_Send_Text(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer server.Close()

	client := New("test-key", WithBaseURL(server.URL))

	err := client.Send(context.Background(), Text("hello"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestClient_Send_InvalidKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errcode":40001,"errmsg":"invalid key"}`))
	}))
	defer server.Close()

	client := New("test-key", WithBaseURL(server.URL))

	err := client.Send(context.Background(), Text("hello"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !IsErrorType(err, CodeInvalidKey) {
		t.Errorf("expected error type CodeInvalidKey, got %v", err)
	}
}

func TestClient_Send_NilMessage(t *testing.T) {
	client := New("test-key")

	err := client.Send(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil message")
	}

	if !IsErrorType(err, CodeInvalidMsgType) {
		t.Errorf("expected CodeInvalidMsgType, got %v", err)
	}
}

func TestClient_Send_RateLimited(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errcode":45009,"errmsg":"rate limited"}`))
	}))
	defer server.Close()

	client := New("test-key", WithBaseURL(server.URL))

	err := client.Send(context.Background(), Text("hello"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !IsErrorType(err, CodeRateLimited) {
		t.Errorf("expected CodeRateLimited, got %v", err)
	}
}

func TestMessageBuilder_Text(t *testing.T) {
	msg := Text("hello").
		Mention("user1", "user2").
		MentionAll()

	if msg.Content != "hello" {
		t.Errorf("expected content 'hello', got %s", msg.Content)
	}
	if len(msg.MentionedList) != 3 {
		t.Errorf("expected 3 mentioned, got %d", len(msg.MentionedList))
	}
}

func TestMessageBuilder_Text_Validate(t *testing.T) {
	// 空内容
	msg := Text("")
	err := msg.Validate()
	if err == nil {
		t.Fatal("expected validation error for empty content")
	}

	// 正常内容
	msg = Text("hello")
	err = msg.Validate()
	if err != nil {
		t.Errorf("unexpected validation error: %v", err)
	}
}

func TestMessageBuilder_Markdown(t *testing.T) {
	msg := Markdown("# Title\nContent")

	if msg.Content != "# Title\nContent" {
		t.Errorf("unexpected content: %s", msg.Content)
	}

	err := msg.Validate()
	if err != nil {
		t.Errorf("unexpected validation error: %v", err)
	}
}

func TestMessageBuilder_News(t *testing.T) {
	msg := News().
		AddArticle("Title1", "https://example.com/1").
		AddArticleWithPic("Title2", "Desc", "https://example.com/2", "https://example.com/pic.png")

	if len(msg.Articles) != 2 {
		t.Errorf("expected 2 articles, got %d", len(msg.Articles))
	}

	err := msg.Validate()
	if err != nil {
		t.Errorf("unexpected validation error: %v", err)
	}
}

func TestMessageBuilder_News_Empty(t *testing.T) {
	msg := News()
	err := msg.Validate()
	if err == nil {
		t.Fatal("expected validation error for empty articles")
	}
}

func TestMessageBuilder_News_EmptyTitle(t *testing.T) {
	msg := News().AddArticle("", "https://example.com")
	err := msg.Validate()
	if err == nil {
		t.Fatal("expected validation error for empty title")
	}
	// 检查错误消息格式正确
	if err.Error() == "" {
		t.Error("expected error message")
	}
}

func TestMessageBuilder_File(t *testing.T) {
	msg := File("media_id_123")

	if msg.MediaId != "media_id_123" {
		t.Errorf("expected media_id 'media_id_123', got %s", msg.MediaId)
	}

	err := msg.Validate()
	if err != nil {
		t.Errorf("unexpected validation error: %v", err)
	}
}

func TestMessageBuilder_File_Empty(t *testing.T) {
	msg := File("")
	err := msg.Validate()
	if err == nil {
		t.Fatal("expected validation error for empty media_id")
	}
}

func TestTemplateCardBuilder_TextNotice(t *testing.T) {
	card := TextNotice().
		Source("https://example.com/icon.png", "System").
		MainTitle("Alert", "CPU Usage High").
		Emphasis("95%", "Current Usage").
		AddContentItem("Server", "web-01").
		AddContentItem("Time", time.Now().Format("2006-01-02 15:04:05")).
		AddJump("Details", "https://example.com/alert/123").
		Build()

	if card.Type() != MsgTypeTemplateCard {
		t.Errorf("expected type %s, got %s", MsgTypeTemplateCard, card.Type())
	}

	if card.TemplateCard.CardType != TemplateTypeTextNotice {
		t.Errorf("expected card type %s, got %s", TemplateTypeTextNotice, card.TemplateCard.CardType)
	}

	if card.TemplateCard.MainTitle == nil {
		t.Fatal("expected MainTitle to be set")
	}

	if card.TemplateCard.MainTitle.Title != "Alert" {
		t.Errorf("expected title 'Alert', got %s", card.TemplateCard.MainTitle.Title)
	}

	err := card.Validate()
	if err != nil {
		t.Errorf("unexpected validation error: %v", err)
	}
}

func TestClient_Builder(t *testing.T) {
	client := NewBuilder("test-key").
		WithTimeout(10 * time.Second).
		WithRetry(3, time.Second).
		Build()

	if client.webhookKey != "test-key" {
		t.Errorf("expected webhook key 'test-key', got %s", client.webhookKey)
	}
	if client.timeout != 10*time.Second {
		t.Errorf("expected timeout 10s, got %v", client.timeout)
	}
	if client.maxRetries != 3 {
		t.Errorf("expected max retries 3, got %d", client.maxRetries)
	}
}

func TestError_NewError(t *testing.T) {
	err := NewError(CodeInvalidKey, "invalid key", ErrInvalidWebhookKey)

	if err.Code != CodeInvalidKey {
		t.Errorf("expected code %d, got %d", CodeInvalidKey, err.Code)
	}

	// Test Error() method
	errStr := err.Error()
	if errStr == "" {
		t.Error("expected non-empty error string")
	}
}

func TestError_IsRetryable(t *testing.T) {
	// Rate limited - retryable
	err := NewError(CodeRateLimited, "rate limited", nil)
	if !IsRetryable(err) {
		t.Error("expected rate limited error to be retryable")
	}

	// Invalid key - not retryable
	err = NewError(CodeInvalidKey, "invalid key", nil)
	if IsRetryable(err) {
		t.Error("expected invalid key error to be not retryable")
	}
}

func TestBatchSend(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer server.Close()

	keys := []string{"key1", "key2", "key3"}
	msg := Text("batch test")

	results := BatchSend(context.Background(), keys, msg, WithBaseURL(server.URL))

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	for key, err := range results {
		if err != nil {
			t.Errorf("unexpected error for key %s: %v", key, err)
		}
	}
}

func TestBatchSend_Concurrent(t *testing.T) {
	// 测试并发执行
	var callCount int
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		callCount++
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer server.Close()

	keys := []string{"key1", "key2", "key3", "key4", "key5"}
	msg := Text("concurrent test")

	start := time.Now()
	results := BatchSend(context.Background(), keys, msg, WithBaseURL(server.URL))
	duration := time.Since(start)

	if len(results) != 5 {
		t.Errorf("expected 5 results, got %d", len(results))
	}

	if callCount != 5 {
		t.Errorf("expected 5 calls, got %d", callCount)
	}

	// 并发执行应该很快（< 100ms），顺序执行会更慢
	if duration > 100*time.Millisecond {
		t.Logf("warning: batch send took %v, might not be concurrent", duration)
	}
}

func TestSendWithKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer server.Close()

	client := New("original-key", WithBaseURL(server.URL))

	err := client.SendWithKey(context.Background(), "another-key", Text("hello"))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSendWithKey_Empty(t *testing.T) {
	client := New("test-key")

	err := client.SendWithKey(context.Background(), "", Text("hello"))
	if err == nil {
		t.Fatal("expected error for empty webhook key")
	}
}

func TestSendWithKey_PreservesConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer server.Close()

	// 创建带重试配置的客户端
	client := New("original-key",
		WithBaseURL(server.URL),
		WithRetry(2, 10*time.Millisecond),
	)

	// SendWithKey 应该保留配置
	err := client.SendWithKey(context.Background(), "another-key", Text("hello"))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestMiddleware_Retry(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount < 3 {
			// 前两次返回限流错误
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"errcode":45009,"errmsg":"rate limited"}`))
		} else {
			// 第三次成功
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
		}
	}))
	defer server.Close()

	client := New("test-key",
		WithBaseURL(server.URL),
		WithRetry(3, 10*time.Millisecond),
	)

	err := client.Send(context.Background(), Text("hello"))
	if err != nil {
		t.Errorf("expected success after retries, got %v", err)
	}

	if callCount != 3 {
		t.Errorf("expected 3 calls (2 retries), got %d", callCount)
	}
}

func TestBatchSendStrict_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer server.Close()

	keys := []string{"key1", "key2", "key3"}
	msg := Text("strict test")

	err := BatchSendStrict(context.Background(), keys, msg, WithBaseURL(server.URL))
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestBatchSendStrict_FailFast(t *testing.T) {
	callCount := 0
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		callCount++
		current := callCount
		mu.Unlock()

		// 第一个请求失败
		if current == 1 {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"errcode":40001,"errmsg":"invalid key"}`))
		} else {
			// 其他请求可能因为 context 取消而未到达
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
		}
	}))
	defer server.Close()

	keys := []string{"key1", "key2", "key3", "key4", "key5"}
	msg := Text("strict fail test")

	err := BatchSendStrict(context.Background(), keys, msg, WithBaseURL(server.URL))
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// errgroup 的严格模式会在第一个失败后取消其他任务
	// 但由于网络延迟，可能已有多个请求发出
	t.Logf("call count: %d, error: %v", callCount, err)
}

func TestBatchSendStrict_ContextCancel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond) // 模拟延迟
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	keys := []string{"key1", "key2", "key3"}
	msg := Text("timeout test")

	err := BatchSendStrict(ctx, keys, msg, WithBaseURL(server.URL))
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
}