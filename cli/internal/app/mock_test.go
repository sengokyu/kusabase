package app

import (
	"context"

	"github.com/sengokyu/kusabase/cli/internal/domain"
	"github.com/sengokyu/kusabase/cli/internal/ports"
)

// mockAPI は ports.ExternalAPIClient のテスト用モック。
// 各フィールドに関数をセットすることで任意の挙動を注入できる。
type mockAPI struct {
	loginFn            func(ctx context.Context, email, password string) error
	sendChatFn         func(ctx context.Context, req ports.ChatRequest) (string, error)
	getConversationsFn func(ctx context.Context) ([]domain.Conversation, error)
	listToolsFn        func(ctx context.Context) ([]domain.Tool, error)
	listPresetsFn      func(ctx context.Context) ([]domain.Model, error)
}

func (m *mockAPI) Login(ctx context.Context, email, password string) error {
	if m.loginFn != nil {
		return m.loginFn(ctx, email, password)
	}
	return nil
}

func (m *mockAPI) SendChat(ctx context.Context, req ports.ChatRequest) (string, error) {
	if m.sendChatFn != nil {
		return m.sendChatFn(ctx, req)
	}
	return "", nil
}

func (m *mockAPI) GetConversations(ctx context.Context) ([]domain.Conversation, error) {
	if m.getConversationsFn != nil {
		return m.getConversationsFn(ctx)
	}
	return nil, nil
}

func (m *mockAPI) ListTools(ctx context.Context) ([]domain.Tool, error) {
	if m.listToolsFn != nil {
		return m.listToolsFn(ctx)
	}
	return nil, nil
}

func (m *mockAPI) ListPresets(ctx context.Context) ([]domain.Model, error) {
	if m.listPresetsFn != nil {
		return m.listPresetsFn(ctx)
	}
	return nil, nil
}

// mockStorage は ports.Storage のテスト用モック。
type mockStorage struct {
	saveActiveSessionFn  func(session domain.ActiveSession) error
	loadActiveSessionFn  func() (*domain.ActiveSession, error)
	clearActiveSessionFn func() error
	saveSessionMetaFn    func(meta ports.SessionMeta) error
	loadSessionsMetaFn   func() ([]ports.SessionMeta, error)
	deleteSessionMetaFn  func(uuid string) error
}

func (m *mockStorage) SaveActiveSession(session domain.ActiveSession) error {
	if m.saveActiveSessionFn != nil {
		return m.saveActiveSessionFn(session)
	}
	return nil
}

func (m *mockStorage) LoadActiveSession() (*domain.ActiveSession, error) {
	if m.loadActiveSessionFn != nil {
		return m.loadActiveSessionFn()
	}
	return nil, nil
}

func (m *mockStorage) ClearActiveSession() error {
	if m.clearActiveSessionFn != nil {
		return m.clearActiveSessionFn()
	}
	return nil
}

func (m *mockStorage) SaveSessionMeta(meta ports.SessionMeta) error {
	if m.saveSessionMetaFn != nil {
		return m.saveSessionMetaFn(meta)
	}
	return nil
}

func (m *mockStorage) LoadSessionsMeta() ([]ports.SessionMeta, error) {
	if m.loadSessionsMetaFn != nil {
		return m.loadSessionsMetaFn()
	}
	return nil, nil
}

func (m *mockStorage) DeleteSessionMeta(uuid string) error {
	if m.deleteSessionMetaFn != nil {
		return m.deleteSessionMetaFn(uuid)
	}
	return nil
}
