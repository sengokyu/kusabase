package client_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/joho/godotenv"
	kusaclient "github.com/sengokyu/kusabase/client"
)

// TestMain はテスト実行前に .env ファイルを読み込む。
// .env が存在しない場合は無視する。
func TestMain(m *testing.M) {
	_ = godotenv.Load()
	os.Exit(m.Run())
}

// ── テスト用ヘルパー ──────────────────────────────────────────

// memStore はテスト用スレッドセーフなインメモリ Store。
type memStore struct {
	mu   sync.RWMutex
	data map[string]string
}

// newMemStore は memStore を作成する。
// 引数は key, value の順で初期値を設定できる。
func newMemStore(initial ...string) *memStore {
	s := &memStore{data: make(map[string]string)}
	for i := 0; i+1 < len(initial); i += 2 {
		s.data[initial[i]] = initial[i+1]
	}
	return s
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

// newTestClient はテスト用サーバーを向く Client を返す。
func newTestClient(t *testing.T, srv *httptest.Server, storeValues ...string) (*kusaclient.Client, *memStore) {
	t.Helper()
	store := newMemStore(storeValues...)
	client := kusaclient.New(kusaclient.Config{
		BaseURL: srv.URL,
		Store:   store,
	})
	return client, store
}

// respondJSON は JSON レスポンスを書き込む。
func respondJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// respondText はプレーンテキストのレスポンスを書き込む。
func respondText(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "text/plain")
	_, _ = io.WriteString(w, body)
}

// ── ユニットテスト ────────────────────────────────────────────

// TestNew_RestoresSession は Store に保存済みのセッションが
// リクエスト時に Cookie として送信されることを確認する。
func TestNew_RestoresSession(t *testing.T) {
	var gotCookie string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c, err := r.Cookie("next-session"); err == nil {
			gotCookie = c.Value
		}
		respondJSON(w, http.StatusOK, kusaclient.PresetListResponse{})
	}))
	defer srv.Close()

	client, _ := newTestClient(t, srv, "next-session", "saved-session-token")
	_, _ = client.Presets.List(context.Background())

	if gotCookie != "saved-session-token" {
		t.Errorf("restored session: want %q, got %q", "saved-session-token", gotCookie)
	}
}

// TestAuth_LoginWithPassword_Success はログイン成功時にセッション Cookie が
// Store に保存されることを確認する。
func TestAuth_LoginWithPassword_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/auth/password/login" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
			return
		}
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if body.Email != "user@example.com" || body.Password != "secret" {
			t.Errorf("unexpected credentials: email=%q password=%q", body.Email, body.Password)
		}
		http.SetCookie(w, &http.Cookie{Name: "next-session", Value: "new-session-token"})
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client, store := newTestClient(t, srv)
	ctx := context.Background()

	if err := client.Auth.LoginWithPassword(ctx, "user@example.com", "secret"); err != nil {
		t.Fatalf("LoginWithPassword: %v", err)
	}

	if sess, _ := store.Load(ctx, "next-session"); sess != "new-session-token" {
		t.Errorf("session saved: want %q, got %q", "new-session-token", sess)
	}
}

// TestAuth_LoginWithPassword_Error はログイン失敗時に *APIError が返ることを確認する。
func TestAuth_LoginWithPassword_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusBadRequest, map[string]any{
			"error": map[string]string{"ns": "auth", "key": "invalid_credentials"},
		})
	}))
	defer srv.Close()

	client, _ := newTestClient(t, srv)
	err := client.Auth.LoginWithPassword(context.Background(), "bad@example.com", "wrong")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*kusaclient.APIError)
	if !ok {
		t.Fatalf("want *kusaclient.APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode: want 400, got %d", apiErr.StatusCode)
	}
	if apiErr.Namespace != "auth" {
		t.Errorf("Namespace: want %q, got %q", "auth", apiErr.Namespace)
	}
	if apiErr.Key != "invalid_credentials" {
		t.Errorf("Key: want %q, got %q", "invalid_credentials", apiErr.Key)
	}
}

// TestAuth_Probe は認証方式問い合わせのレスポンスが正しくパースされることを確認する。
func TestAuth_Probe(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/auth/probe" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
			return
		}
		respondJSON(w, http.StatusOK, kusaclient.AuthProbeResponse{
			AllowPassword:     true,
			ExternalProviders: []string{"google"},
		})
	}))
	defer srv.Close()

	client, _ := newTestClient(t, srv)
	resp, err := client.Auth.Probe(context.Background(), kusaclient.AuthProbeRequest{
		Email:       "user@example.com",
		RedirectURL: "https://example.com/cb",
	})
	if err != nil {
		t.Fatalf("Probe: %v", err)
	}
	if !resp.AllowPassword {
		t.Error("AllowPassword: want true")
	}
	if len(resp.ExternalProviders) != 1 || resp.ExternalProviders[0] != "google" {
		t.Errorf("ExternalProviders: want [google], got %v", resp.ExternalProviders)
	}
}

// TestPresets_List はプリセット一覧が正しくパースされることを確認する。
func TestPresets_List(t *testing.T) {
	fixture := kusaclient.PresetListResponse{
		Presets: []kusaclient.Preset{
			{ID: 1, UUID: "preset-uuid-1"},
			{ID: 2, UUID: "preset-uuid-2"},
		},
		DefaultPreset: "preset-uuid-1",
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/preset" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
			return
		}
		respondJSON(w, http.StatusOK, fixture)
	}))
	defer srv.Close()

	client, _ := newTestClient(t, srv)
	presets, err := client.Presets.List(context.Background())
	if err != nil {
		t.Fatalf("Presets.List: %v", err)
	}
	if len(presets) != 2 {
		t.Fatalf("len(presets): want 2, got %d", len(presets))
	}
	if presets[0].UUID != "preset-uuid-1" || presets[1].UUID != "preset-uuid-2" {
		t.Errorf("unexpected preset UUIDs: %v", presets)
	}
}

// TestConversation_List は会話履歴一覧が正しくパースされることを確認する。
func TestConversation_List(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/conversation/overview" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
			return
		}
		respondJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data": map[string]any{
				"latestConversations": []map[string]any{
					{"uuid": "conv-uuid-1", "title": "テスト会話"},
				},
			},
		})
	}))
	defer srv.Close()

	client, _ := newTestClient(t, srv)
	convs, err := client.Conversation.List(context.Background())
	if err != nil {
		t.Fatalf("Conversation.List: %v", err)
	}
	if len(convs) != 1 {
		t.Fatalf("len(convs): want 1, got %d", len(convs))
	}
	if convs[0].UUID != "conv-uuid-1" {
		t.Errorf("UUID: want %q, got %q", "conv-uuid-1", convs[0].UUID)
	}
	if convs[0].Title != "テスト会話" {
		t.Errorf("Title: want %q, got %q", "テスト会話", convs[0].Title)
	}
}

// TestTools_List はツール一覧が正しくパースされることを確認する。
func TestTools_List(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/tools" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
			return
		}
		respondJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data": []map[string]any{
				{"uuid": "tool-uuid-1", "name": "WebSearch"},
			},
		})
	}))
	defer srv.Close()

	client, _ := newTestClient(t, srv)
	tools, err := client.Tools.List(context.Background())
	if err != nil {
		t.Fatalf("Tools.List: %v", err)
	}
	if len(tools) != 1 {
		t.Fatalf("len(tools): want 1, got %d", len(tools))
	}
	if tools[0].UUID != "tool-uuid-1" || tools[0].Name != "WebSearch" {
		t.Errorf("unexpected tool: %+v", tools[0])
	}
}

// TestChat_New は新規会話開始時に fastHeaders=true が送信され、
// AI の返答が Chat.LastResponse に格納されることを確認する。
func TestChat_New(t *testing.T) {
	type chatBody struct {
		Content     string `json:"content"`
		FastHeaders bool   `json:"fastHeaders"`
		PresetID    int    `json:"presetId"`
	}
	var got chatBody

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/chat" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
			return
		}
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &got)
		respondText(w, "Hello from AI")
	}))
	defer srv.Close()

	client, _ := newTestClient(t, srv)
	chat, err := client.Chat.New(kusaclient.Preset{ID: 42}, kusaclient.ChatRequest{Content: "こんにちは"})
	if err != nil {
		t.Fatalf("Chat.New: %v", err)
	}

	if got.Content != "こんにちは" {
		t.Errorf("content: want %q, got %q", "こんにちは", got.Content)
	}
	if !got.FastHeaders {
		t.Error("fastHeaders: want true for first message")
	}
	if got.PresetID != 42 {
		t.Errorf("presetId: want 42, got %d", got.PresetID)
	}
	if chat.LastResponse != "Hello from AI" {
		t.Errorf("LastResponse: want %q, got %q", "Hello from AI", chat.LastResponse)
	}
}

// TestChat_Send はフォローアップメッセージで fastHeaders=false が送信され、
// 返答が LastResponse に反映されることを確認する。
func TestChat_Send(t *testing.T) {
	type chatBody struct {
		Content     string `json:"content"`
		FastHeaders bool   `json:"fastHeaders"`
	}
	callCount := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var body chatBody
		_ = json.NewDecoder(r.Body).Decode(&body)

		if callCount == 2 && body.FastHeaders {
			t.Error("fastHeaders: want false for follow-up message")
		}
		if callCount == 1 {
			respondText(w, "first reply")
		} else {
			respondText(w, "second reply")
		}
	}))
	defer srv.Close()

	client, _ := newTestClient(t, srv)
	chat, err := client.Chat.New(kusaclient.Preset{ID: 1}, kusaclient.ChatRequest{Content: "Hello"})
	if err != nil {
		t.Fatalf("Chat.New: %v", err)
	}

	reply, err := chat.Send(kusaclient.ChatRequest{Content: "How are you?"})
	if err != nil {
		t.Fatalf("Chat.Send: %v", err)
	}
	if reply != "second reply" {
		t.Errorf("reply: want %q, got %q", "second reply", reply)
	}
	if chat.LastResponse != "second reply" {
		t.Errorf("LastResponse after Send: want %q, got %q", "second reply", chat.LastResponse)
	}
}

// TestAPIError_Error はエラーメッセージのフォーマットを確認する。
func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		err  *kusaclient.APIError
		want string
	}{
		{
			err:  &kusaclient.APIError{StatusCode: 400, Namespace: "auth", Key: "invalid_credentials"},
			want: "api error 400: auth:invalid_credentials",
		},
		{
			err:  &kusaclient.APIError{StatusCode: 500},
			want: "api error 500",
		},
	}
	for _, tt := range tests {
		if got := tt.err.Error(); got != tt.want {
			t.Errorf("APIError.Error() = %q, want %q", got, tt.want)
		}
	}
}
