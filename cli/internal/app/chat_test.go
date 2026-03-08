package app

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/sengokyu/kusabase/cli/internal/domain"
	"github.com/sengokyu/kusabase/cli/internal/ports"
)

func TestChatUsecaseList(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		apiConvs  []domain.Conversation
		apiErr    error
		metas     []ports.SessionMeta
		wantLen   int
		wantErr   error
		checkFunc func(t *testing.T, convs []domain.Conversation)
	}{
		{
			name:    "未ログイン時は ErrNotLoggedIn を返す",
			apiErr:  ports.ErrNotLoggedIn,
			wantErr: ports.ErrNotLoggedIn,
		},
		{
			name:    "会話なし",
			wantLen: 0,
		},
		{
			name:     "ローカルメタがない場合は ModelName/ToolNames が空",
			apiConvs: []domain.Conversation{{UUID: "uuid-1", Title: "テスト", LastMessageAt: now}},
			wantLen:  1,
			checkFunc: func(t *testing.T, convs []domain.Conversation) {
				if convs[0].ModelName != "" {
					t.Errorf("ModelName = %q, want empty", convs[0].ModelName)
				}
				if len(convs[0].ToolNames) != 0 {
					t.Errorf("ToolNames = %v, want empty", convs[0].ToolNames)
				}
			},
		},
		{
			name:     "UUID が一致するメタで ModelName/ToolNames が補完される",
			apiConvs: []domain.Conversation{{UUID: "uuid-1", Title: "テスト", LastMessageAt: now}},
			metas:    []ports.SessionMeta{{UUID: "uuid-1", ModelName: "gpt-4.1", ToolNames: []string{"web_search"}}},
			wantLen:  1,
			checkFunc: func(t *testing.T, convs []domain.Conversation) {
				if convs[0].ModelName != "gpt-4.1" {
					t.Errorf("ModelName = %q, want gpt-4.1", convs[0].ModelName)
				}
				if len(convs[0].ToolNames) != 1 || convs[0].ToolNames[0] != "web_search" {
					t.Errorf("ToolNames = %v, want [web_search]", convs[0].ToolNames)
				}
			},
		},
		{
			name:     "UUID が一致しないメタは無視される",
			apiConvs: []domain.Conversation{{UUID: "uuid-1", Title: "テスト", LastMessageAt: now}},
			metas:    []ports.SessionMeta{{UUID: "uuid-999", ModelName: "gpt-4.1"}},
			wantLen:  1,
			checkFunc: func(t *testing.T, convs []domain.Conversation) {
				if convs[0].ModelName != "" {
					t.Errorf("ModelName = %q, want empty", convs[0].ModelName)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &mockAPI{
				getConversationsFn: func(_ context.Context) ([]domain.Conversation, error) {
					return tt.apiConvs, tt.apiErr
				},
			}
			storage := &mockStorage{
				loadSessionsMetaFn: func() ([]ports.SessionMeta, error) {
					return tt.metas, nil
				},
			}
			u := NewChatUsecase(api, storage, strings.NewReader(""))

			convs, err := u.List(context.Background())

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if len(convs) != tt.wantLen {
				t.Fatalf("len(convs) = %d, want %d", len(convs), tt.wantLen)
			}
			if tt.checkFunc != nil {
				tt.checkFunc(t, convs)
			}
		})
	}
}

func TestChatUsecaseDelete(t *testing.T) {
	const targetUUID = "uuid-target"

	tests := []struct {
		name            string
		activeSession   *domain.ActiveSession
		deleteMetaErr   error
		wantClearCalled bool
		wantErr         bool
	}{
		{
			name:            "非アクティブセッションの削除では ClearActiveSession は呼ばれない",
			activeSession:   &domain.ActiveSession{ConversationUUID: "uuid-other"},
			wantClearCalled: false,
		},
		{
			name:            "アクティブセッションの削除では ClearActiveSession が呼ばれる",
			activeSession:   &domain.ActiveSession{ConversationUUID: targetUUID},
			wantClearCalled: true,
		},
		{
			name:            "アクティブセッションがない場合でも削除は成功する",
			activeSession:   nil,
			wantClearCalled: false,
		},
		{
			name:          "DeleteSessionMeta 失敗時はエラーを返す",
			deleteMetaErr: errors.New("IO error"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearCalled := false
			storage := &mockStorage{
				loadActiveSessionFn: func() (*domain.ActiveSession, error) {
					return tt.activeSession, nil
				},
				deleteSessionMetaFn: func(_ string) error {
					return tt.deleteMetaErr
				},
				clearActiveSessionFn: func() error {
					clearCalled = true
					return nil
				},
			}
			u := NewChatUsecase(&mockAPI{}, storage, strings.NewReader(""))

			err := u.Delete(context.Background(), targetUUID)

			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if clearCalled != tt.wantClearCalled {
				t.Errorf("clearCalled = %v, want %v", clearCalled, tt.wantClearCalled)
			}
		})
	}
}

func TestChatUsecaseStartNew(t *testing.T) {
	availableTools := []domain.Tool{
		{UUID: "tool-uuid-1", Name: "web_search"},
		{UUID: "tool-uuid-2", Name: "file_read"},
	}

	tests := []struct {
		name          string
		toolNames     []string
		listToolsErr  error
		wantErr       error   // errors.Is で照合するセンチネルエラー
		wantErrString string  // エラーメッセージに含まれるべき文字列
		wantModelName string  // StartNew の modelName 引数
	}{
		{
			name:         "未ログイン時は ErrNotLoggedIn を返す",
			listToolsErr: ports.ErrNotLoggedIn,
			wantErr:      ports.ErrNotLoggedIn,
		},
		{
			name:          "不明なツール名を指定するとエラーを返す",
			toolNames:     []string{"unknown_tool"},
			wantErrString: "不明なツール",
		},
		{
			name:      "ツールなしで正常終了（即 EOF）",
			toolNames: []string{},
		},
		{
			name:      "有効なツールを指定して正常終了（即 EOF）",
			toolNames: []string{"web_search"},
		},
		{
			name:          "有効なモデル名を指定して正常終了（即 EOF）",
			wantModelName: "gpt-4.1",
		},
		{
			name:          "不明なモデル名を指定するとエラーを返す",
			wantModelName: "unknown-model",
			wantErrString: "不明なモデル",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &mockAPI{
				listToolsFn: func(_ context.Context) ([]domain.Tool, error) {
					if tt.listToolsErr != nil {
						return nil, tt.listToolsErr
					}
					return availableTools, nil
				},
				listPresetsFn: func(_ context.Context) ([]domain.Model, error) {
					return []domain.Model{
						{ID: 1, UUID: "preset-uuid-1", Name: "GPT-4.1", ModelID: "gpt-4.1"},
					}, nil
				},
			}
			u := NewChatUsecase(api, &mockStorage{}, strings.NewReader(""))

			err := u.StartNew(context.Background(), tt.wantModelName, tt.toolNames)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if tt.wantErrString != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErrString) {
					t.Errorf("err = %v, want error containing %q", err, tt.wantErrString)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestChatUsecaseContinue(t *testing.T) {
	tests := []struct {
		name          string
		activeSession *domain.ActiveSession
		listToolsErr  error
		wantErr       error
	}{
		{
			name:    "セッションなしの場合は ErrNoActiveSession を返す",
			wantErr: ports.ErrNoActiveSession,
		},
		{
			name:          "未ログイン時は ErrNotLoggedIn を返す",
			activeSession: &domain.ActiveSession{ConversationUUID: "uuid-1"},
			listToolsErr:  ports.ErrNotLoggedIn,
			wantErr:       ports.ErrNotLoggedIn,
		},
		{
			name:          "セッションありで正常終了（即 EOF）",
			activeSession: &domain.ActiveSession{ConversationUUID: "uuid-1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &mockAPI{
				listToolsFn: func(_ context.Context) ([]domain.Tool, error) {
					return nil, tt.listToolsErr
				},
			}
			storage := &mockStorage{
				loadActiveSessionFn: func() (*domain.ActiveSession, error) {
					return tt.activeSession, nil
				},
			}
			u := NewChatUsecase(api, storage, strings.NewReader(""))

			err := u.Continue(context.Background())

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
