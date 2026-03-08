package app

import (
	"testing"

	"github.com/sengokyu/kusabase/cli/internal/domain"
	"github.com/sengokyu/kusabase/cli/internal/ports"
)

func TestBuildConfiguredTools(t *testing.T) {
	tools := []domain.Tool{
		{UUID: "u1", Name: "web_search"},
		{UUID: "u2", Name: "file_read"},
	}

	tests := []struct {
		name         string
		enabledNames []string
		want         []ports.ConfiguredTool
	}{
		{
			name:         "有効化なし: 全ツールが disabled",
			enabledNames: []string{},
			want: []ports.ConfiguredTool{
				{UUID: "u1", Enabled: false},
				{UUID: "u2", Enabled: false},
			},
		},
		{
			name:         "一部ツールを有効化",
			enabledNames: []string{"web_search"},
			want: []ports.ConfiguredTool{
				{UUID: "u1", Enabled: true},
				{UUID: "u2", Enabled: false},
			},
		},
		{
			name:         "全ツールを有効化",
			enabledNames: []string{"web_search", "file_read"},
			want: []ports.ConfiguredTool{
				{UUID: "u1", Enabled: true},
				{UUID: "u2", Enabled: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildConfiguredTools(tools, tt.enabledNames)
			if len(got) != len(tt.want) {
				t.Fatalf("len = %d, want %d", len(got), len(tt.want))
			}
			for i, g := range got {
				if g != tt.want[i] {
					t.Errorf("[%d] = %+v, want %+v", i, g, tt.want[i])
				}
			}
		})
	}
}

func TestBuildConfiguredToolsEmpty(t *testing.T) {
	got := buildConfiguredTools([]domain.Tool{}, []string{"web_search"})
	if len(got) != 0 {
		t.Errorf("len = %d, want 0", len(got))
	}
}

func TestBuildToolMap(t *testing.T) {
	tools := []domain.Tool{
		{UUID: "u1", Name: "web_search"},
		{UUID: "u2", Name: "file_read"},
	}

	got := buildToolMap(tools)

	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got["web_search"] != "u1" {
		t.Errorf("web_search = %q, want u1", got["web_search"])
	}
	if got["file_read"] != "u2" {
		t.Errorf("file_read = %q, want u2", got["file_read"])
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name  string
		slice []string
		s     string
		want  bool
	}{
		{"空スライス", []string{}, "a", false},
		{"含まれている", []string{"a", "b", "c"}, "b", true},
		{"含まれていない", []string{"a", "b", "c"}, "d", false},
		{"部分一致は false", []string{"web_search"}, "web", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := contains(tt.slice, tt.s); got != tt.want {
				t.Errorf("contains(%v, %q) = %v, want %v", tt.slice, tt.s, got, tt.want)
			}
		})
	}
}
