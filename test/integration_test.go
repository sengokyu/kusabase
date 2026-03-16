package test

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/joho/godotenv"
	"github.com/sengokyu/kusabase/httpclient"
)

// TestMain はテスト実行前に .env ファイルを読み込む。
// .env が存在しない場合は無視する。
func TestMain(m *testing.M) {
	_ = godotenv.Load("../.env")
	os.Exit(m.Run())
}

// memStore はテスト用スレッドセーフなインメモリ Store。
type memStore struct {
	mu   sync.RWMutex
	data map[string]string
}

func (s *memStore) Save(_ context.Context, key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	return nil
}

func (s *memStore) Load(_ context.Context, key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[key], nil
}

// integrationClient は環境変数 KUSA_BASE_URL を使って Client を作成する。
// 未設定の場合はテストをスキップする。
func integrationClient(t *testing.T) *httpclient.Client {
	t.Helper()
	baseURL := os.Getenv("KUSA_BASE_URL")
	if baseURL == "" {
		t.Skip("KUSA_BASE_URL が未設定のためスキップ")
	}
	return httpclient.New(httpclient.Config{
		BaseURL: baseURL,
		Store:   &memStore{},
	})
}

// mustEnv は必須環境変数を取得し、未設定の場合はスキップする。
func mustEnv(t *testing.T, key string) string {
	t.Helper()
	v := os.Getenv(key)
	if v == "" {
		t.Skipf("環境変数 %s が未設定のためスキップ", key)
	}
	return v
}

// TestIntegration_LoginAndListPresets はログインしてプリセット一覧を取得する。
func TestIntegration_LoginAndListPresets(t *testing.T) {
	email := mustEnv(t, "KUSA_EMAIL")
	password := mustEnv(t, "KUSA_PASSWORD")

	client := integrationClient(t)
	ctx := context.Background()

	if err := client.Auth.LoginWithPassword(ctx, email, password); err != nil {
		t.Fatalf("ログイン失敗: %v", err)
	}
	t.Log("ログイン成功")

	presetsResponse, err := client.Presets.List(ctx)
	if err != nil {
		t.Fatalf("Presets.List 失敗: %v", err)
	}
	if len(presetsResponse.Presets) == 0 {
		t.Error("プリセットが0件です")
	}
	t.Logf("プリセット %d 件取得: 先頭 UUID=%s", len(presetsResponse.Presets), presetsResponse.Presets[0].UUID)
}

// TestIntegration_ListConversations はログインして会話履歴を取得する。
func TestIntegration_ListConversations(t *testing.T) {
	email := mustEnv(t, "KUSA_EMAIL")
	password := mustEnv(t, "KUSA_PASSWORD")

	client := integrationClient(t)
	ctx := context.Background()

	if err := client.Auth.LoginWithPassword(ctx, email, password); err != nil {
		t.Fatalf("ログイン失敗: %v", err)
	}

	convs, err := client.Conversation.List(ctx)
	if err != nil {
		t.Fatalf("Conversation.List 失敗: %v", err)
	}
	t.Logf("会話履歴 %d 件取得", len(convs))
}

// TestIntegration_ListTools はログインしてツール一覧を取得する。
func TestIntegration_ListTools(t *testing.T) {
	email := mustEnv(t, "KUSA_EMAIL")
	password := mustEnv(t, "KUSA_PASSWORD")

	client := integrationClient(t)
	ctx := context.Background()

	if err := client.Auth.LoginWithPassword(ctx, email, password); err != nil {
		t.Fatalf("ログイン失敗: %v", err)
	}

	tools, err := client.Tools.List(ctx)
	if err != nil {
		t.Fatalf("Tools.List 失敗: %v", err)
	}
	t.Logf("ツール %d 件取得", len(tools))
}

// TestIntegration_Chat はログインしてチャットを実行する。
func TestIntegration_Chat(t *testing.T) {
	email := mustEnv(t, "KUSA_EMAIL")
	password := mustEnv(t, "KUSA_PASSWORD")

	client := integrationClient(t)
	ctx := context.Background()

	if err := client.Auth.LoginWithPassword(ctx, email, password); err != nil {
		t.Fatalf("ログイン失敗: %v", err)
	}

	presetsResponse, err := client.Presets.List(ctx)
	if err != nil || len(presetsResponse.Presets) == 0 {
		t.Fatalf("プリセット取得失敗: %v (件数=%d)", err, len(presetsResponse.Presets))
	}

	chat, err := client.Chat.New(presetsResponse.Presets[0], httpclient.ChatRequest{
		Content: "こんにちは。一言だけ返してください。",
	})
	if err != nil {
		t.Fatalf("Chat.New 失敗: %v", err)
	}
	t.Logf("1回目の返答: %q", chat.LastResponse)

	reply, err := chat.Send(httpclient.ChatRequest{
		Content: "ありがとう。",
	})
	if err != nil {
		t.Fatalf("Chat.Send 失敗: %v", err)
	}
	t.Logf("2回目の返答: %q", reply)
}
