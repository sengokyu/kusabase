package app

import (
	"context"
	"errors"
	"testing"

	"github.com/sengokyu/kusabase/cli/internal/domain"
	"github.com/sengokyu/kusabase/cli/internal/ports"
)

func TestToolUsecaseList(t *testing.T) {
	tests := []struct {
		name    string
		tools   []domain.Tool
		apiErr  error
		wantErr error
	}{
		{
			name:    "未ログイン時は ErrNotLoggedIn を返す",
			apiErr:  ports.ErrNotLoggedIn,
			wantErr: ports.ErrNotLoggedIn,
		},
		{
			name:  "ツールが空のときは空スライスを返す",
			tools: []domain.Tool{},
		},
		{
			name: "複数ツールをそのまま返す",
			tools: []domain.Tool{
				{UUID: "u1", Name: "web_search", DisplayName: "Web 検索", Description: "Web を検索する"},
				{UUID: "u2", Name: "file_read", DisplayName: "ファイル読み込み", Description: "ファイルを読み込む"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &mockAPI{
				listToolsFn: func(_ context.Context) ([]domain.Tool, error) {
					return tt.tools, tt.apiErr
				},
			}
			u := NewToolUsecase(api)

			tools, err := u.List(context.Background())

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if len(tools) != len(tt.tools) {
				t.Errorf("len(tools) = %d, want %d", len(tools), len(tt.tools))
			}
		})
	}
}
