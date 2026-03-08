package app

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sengokyu/kusabase/cli/internal/domain"
	"github.com/sengokyu/kusabase/cli/internal/ports"
)

// ChatUsecase handles chat session operations.
type ChatUsecase struct {
	api     ports.ExternalAPIClient
	storage ports.Storage
	in      io.Reader // 対話入力ソース（通常は os.Stdin）
}

// NewChatUsecase creates a new ChatUsecase.
func NewChatUsecase(api ports.ExternalAPIClient, storage ports.Storage, in io.Reader) *ChatUsecase {
	return &ChatUsecase{api: api, storage: storage, in: in}
}

// StartNew starts a new chat session and enters interactive mode.
func (u *ChatUsecase) StartNew(ctx context.Context, modelName string, toolNames []string) error {
	tools, err := u.api.ListTools(ctx)
	if err != nil {
		if errors.Is(err, ports.ErrNotLoggedIn) {
			return ports.ErrNotLoggedIn
		}
		return fmt.Errorf("ツール一覧の取得に失敗しました: %w", err)
	}

	// Validate specified tool names
	toolMap := buildToolMap(tools)
	for _, name := range toolNames {
		if _, ok := toolMap[name]; !ok {
			return fmt.Errorf("不明なツール: %s", name)
		}
	}

	// Resolve preset ID from model name
	var presetID *int
	if modelName != "" {
		models, err := u.api.ListPresets(ctx)
		if err != nil {
			if errors.Is(err, ports.ErrNotLoggedIn) {
				return ports.ErrNotLoggedIn
			}
			return fmt.Errorf("プリセット一覧の取得に失敗しました: %w", err)
		}
		for _, m := range models {
			if m.ModelID == modelName || m.Name == modelName {
				id := m.ID
				presetID = &id
				break
			}
		}
		if presetID == nil {
			return fmt.Errorf("不明なモデル: %s", modelName)
		}
	}

	// Build configured tools list (all tools, enable only the specified ones)
	configuredTools := buildConfiguredTools(tools, toolNames)

	if modelName != "" {
		fmt.Printf("モデル: %s\n", modelName)
	}
	if len(toolNames) > 0 {
		fmt.Printf("ツール: %s\n", strings.Join(toolNames, ", "))
	}
	fmt.Println("新しいチャットを開始します。Ctrl+D で終了します。")
	fmt.Println()

	return u.runREPL(ctx, "", modelName, toolNames, configuredTools, presetID, true)
}

// Continue resumes the active chat session.
func (u *ChatUsecase) Continue(ctx context.Context) error {
	session, err := u.storage.LoadActiveSession()
	if err != nil {
		return fmt.Errorf("セッションの読み込みに失敗しました: %w", err)
	}
	if session == nil {
		return ports.ErrNoActiveSession
	}

	tools, err := u.api.ListTools(ctx)
	if err != nil {
		if errors.Is(err, ports.ErrNotLoggedIn) {
			return ports.ErrNotLoggedIn
		}
		return fmt.Errorf("ツール一覧の取得に失敗しました: %w", err)
	}

	configuredTools := buildConfiguredTools(tools, session.ToolNames)

	fmt.Println("チャットを再開します。Ctrl+D で終了します。")
	fmt.Println()

	return u.runREPL(ctx, session.ConversationUUID, session.ModelName, session.ToolNames, configuredTools, nil, false)
}

// List returns conversations enriched with local metadata.
func (u *ChatUsecase) List(ctx context.Context) ([]domain.Conversation, error) {
	convs, err := u.api.GetConversations(ctx)
	if err != nil {
		return nil, err
	}

	metas, _ := u.storage.LoadSessionsMeta()
	metaMap := make(map[string]ports.SessionMeta)
	for _, m := range metas {
		metaMap[m.UUID] = m
	}

	for i, c := range convs {
		if meta, ok := metaMap[c.UUID]; ok {
			convs[i].ModelName = meta.ModelName
			convs[i].ToolNames = meta.ToolNames
		}
	}
	return convs, nil
}

// Delete removes a conversation from local metadata.
// NOTE: Server-side deletion is not yet supported (API未対応).
func (u *ChatUsecase) Delete(ctx context.Context, uuid string) error {
	if err := u.storage.DeleteSessionMeta(uuid); err != nil {
		return fmt.Errorf("ローカルセッション削除に失敗しました: %w", err)
	}
	session, _ := u.storage.LoadActiveSession()
	if session != nil && session.ConversationUUID == uuid {
		_ = u.storage.ClearActiveSession()
	}
	return nil
}

func (u *ChatUsecase) runREPL(
	ctx context.Context,
	conversationUUID, modelName string,
	toolNames []string,
	configuredTools []ports.ConfiguredTool,
	presetID *int,
	isNew bool,
) error {
	scanner := bufio.NewScanner(u.in)
	isFirst := isNew

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			fmt.Println()
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		req := ports.ChatRequest{
			Content:          line,
			ConversationUUID: conversationUUID,
			ConfiguredTools:  configuredTools,
			FastHeaders:      isFirst,
			PresetID:         presetID,
		}

		response, err := u.api.SendChat(ctx, req)
		if err != nil {
			if errors.Is(err, ports.ErrNotLoggedIn) {
				return ports.ErrNotLoggedIn
			}
			fmt.Fprintf(os.Stderr, "エラー: %v\n", err)
			continue
		}

		fmt.Println(response)

		if isFirst {
			isFirst = false
			// Retrieve the new conversation UUID from the overview API
			convs, err := u.api.GetConversations(ctx)
			if err == nil && len(convs) > 0 {
				conversationUUID = convs[0].UUID
				session := domain.ActiveSession{
					ConversationUUID: conversationUUID,
					ModelName:        modelName,
					ToolNames:        toolNames,
				}
				_ = u.storage.SaveActiveSession(session)
				_ = u.storage.SaveSessionMeta(ports.SessionMeta{
					UUID:      conversationUUID,
					ModelName: modelName,
					ToolNames: toolNames,
				})
			}
		}
	}

	return scanner.Err()
}

func buildToolMap(tools []domain.Tool) map[string]string {
	m := make(map[string]string, len(tools))
	for _, t := range tools {
		m[t.Name] = t.UUID
	}
	return m
}

func buildConfiguredTools(tools []domain.Tool, enabledNames []string) []ports.ConfiguredTool {
	result := make([]ports.ConfiguredTool, 0, len(tools))
	for _, t := range tools {
		result = append(result, ports.ConfiguredTool{
			UUID:    t.UUID,
			Enabled: contains(enabledNames, t.Name),
		})
	}
	return result
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
