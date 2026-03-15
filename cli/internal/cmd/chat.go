package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/sengokyu/kusabase/cli/internal/app"
	httpclient "github.com/sengokyu/kusabase/httpclient"
)

var (
	chatModel string
	chatTools []string
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Chat with AI",
	RunE:  runChat,
}

var chatNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Start a new chat session",
	RunE:  runChatNew,
}

var chatListCmd = &cobra.Command{
	Use:   "list",
	Short: "List chat sessions",
	RunE:  runChatList,
}

var chatDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a chat session",
	Args:  cobra.ExactArgs(1),
	RunE:  runChatDelete,
}

func init() {
	chatNewCmd.Flags().StringVar(&chatModel, "model", "", "model to use (default: server default)")
	chatNewCmd.Flags().StringArrayVar(&chatTools, "tool", nil, "tool to enable (repeatable)")
	chatCmd.AddCommand(chatNewCmd)
	chatCmd.AddCommand(chatListCmd)
	chatCmd.AddCommand(chatDeleteCmd)
}

// runChat continues the active chat session.
func runChat(_ *cobra.Command, _ []string) error {
	a, err := app.New()
	if err != nil {
		return err
	}

	if !a.IsLoggedIn() {
		printNotLoggedIn()
		return nil
	}

	conversationUUID, err := a.Session.Load()
	if err != nil || conversationUUID == "" {
		fmt.Println("No active chat session.")
		fmt.Println("Run `kusa chat new` to start a new chat,")
		fmt.Println("or `kusa chat list` to resume an existing one.")
		return nil
	}

	debugLog("resuming conversation %s", conversationUUID)
	return interactiveLoop(a, conversationUUID, nil, nil)
}

// runChatNew starts a new chat session.
func runChatNew(_ *cobra.Command, _ []string) error {
	a, err := app.New()
	if err != nil {
		return err
	}

	if !a.IsLoggedIn() {
		printNotLoggedIn()
		return nil
	}

	ctx := context.Background()

	// Resolve preset.
	presetsResp, err := a.Client.Presets.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch presets: %w", err)
	}

	preset, err := selectPreset(presetsResp, chatModel)
	if err != nil {
		return err
	}

	// Resolve tools.
	configuredTools, err := resolveTools(ctx, a, chatTools)
	if err != nil {
		return err
	}

	// Read and send the first message.
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("You: ")
	firstLine, err := reader.ReadString('\n')
	firstLine = strings.TrimSpace(firstLine)
	if err != nil && err != io.EOF {
		return err
	}
	if firstLine == "" {
		return nil
	}

	presetID := preset.ID
	reply, err := a.Client.Chat.Send(ctx, httpclient.ChatRequest{
		Content:         firstLine,
		ConfiguredTools: configuredTools,
		FastHeaders:     true,
		PresetID:        &presetID,
	})
	if err != nil {
		return fmt.Errorf("chat error: %w", err)
	}
	fmt.Printf("\nAI: %s\n\n", reply)

	// Fetch the UUID of the conversation just created.
	var conversationUUID string
	convs, listErr := a.Client.Conversation.List(ctx)
	if listErr != nil {
		debugLog("failed to fetch conversations: %v", listErr)
	} else if len(convs) > 0 {
		conversationUUID = convs[0].UUID
		if saveErr := a.Session.Save(conversationUUID); saveErr != nil {
			debugLog("failed to save session: %v", saveErr)
		}
		debugLog("started conversation %s", conversationUUID)
	}

	// Exit here if EOF was already reached on the first read.
	if err == io.EOF {
		return nil
	}

	return interactiveLoopWithReader(a, conversationUUID, configuredTools, &presetID, reader)
}

// runChatList lists saved chat sessions.
func runChatList(_ *cobra.Command, _ []string) error {
	a, err := app.New()
	if err != nil {
		return err
	}

	if !a.IsLoggedIn() {
		printNotLoggedIn()
		return nil
	}

	ctx := context.Background()
	convs, err := a.Client.Conversation.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list chats: %w", err)
	}

	if len(convs) == 0 {
		fmt.Println("No chat sessions found.")
		return nil
	}

	fmt.Printf("%-4s %-30s %s\n", "ID", "TITLE", "UPDATED")
	for i, c := range convs {
		title := c.Title
		if len(title) > 28 {
			title = title[:25] + "..."
		}
		fmt.Printf("%-4d %-30s %s\n",
			i+1,
			title,
			c.LastMessageAt.Format("2006-01-02 15:04"),
		)
	}
	return nil
}

// runChatDelete is a stub – no delete endpoint exists in the current API.
func runChatDelete(_ *cobra.Command, _ []string) error {
	fmt.Fprintln(os.Stderr, "Error: chat delete is not supported by the current API.")
	return nil
}

// selectPreset finds the preset matching modelName, or the server default.
func selectPreset(resp httpclient.PresetListResponse, modelName string) (*httpclient.Preset, error) {
	if modelName != "" {
		for i := range resp.Presets {
			if resp.Presets[i].ModelParameters.Model == modelName {
				return &resp.Presets[i], nil
			}
		}
		return nil, fmt.Errorf("model %q not found", modelName)
	}
	// Use the default preset.
	for i := range resp.Presets {
		if resp.Presets[i].UUID == resp.DefaultPreset {
			return &resp.Presets[i], nil
		}
	}
	if len(resp.Presets) > 0 {
		return &resp.Presets[0], nil
	}
	return nil, fmt.Errorf("no presets available")
}

// resolveTools maps tool names to ConfiguredTool values with enabled=true.
func resolveTools(ctx context.Context, a *app.App, names []string) ([]httpclient.ConfiguredTool, error) {
	if len(names) == 0 {
		return nil, nil
	}
	allTools, err := a.Client.Tools.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tools: %w", err)
	}
	byName := make(map[string]string, len(allTools))
	for _, t := range allTools {
		byName[t.Name] = t.UUID
	}
	result := make([]httpclient.ConfiguredTool, 0, len(names))
	for _, name := range names {
		uuid, ok := byName[name]
		if !ok {
			return nil, fmt.Errorf("tool %q not found", name)
		}
		result = append(result, httpclient.ConfiguredTool{
			UUID:     uuid,
			Settings: httpclient.ConfiguredToolSettings{Enabled: true},
		})
	}
	return result, nil
}

// interactiveLoop reads lines from a new stdin reader and chats.
func interactiveLoop(a *app.App, conversationUUID string, tools []httpclient.ConfiguredTool, presetID *int) error {
	return interactiveLoopWithReader(a, conversationUUID, tools, presetID, bufio.NewReader(os.Stdin))
}

func interactiveLoopWithReader(a *app.App, conversationUUID string, tools []httpclient.ConfiguredTool, presetID *int, reader *bufio.Reader) error {
	ctx := context.Background()
	for {
		fmt.Print("You: ")
		line, err := reader.ReadString('\n')
		line = strings.TrimSpace(line)

		if line != "" {
			reply, sendErr := a.Client.Chat.Send(ctx, httpclient.ChatRequest{
				Content:          line,
				ConversationUUID: conversationUUID,
				ConfiguredTools:  tools,
				PresetID:         presetID,
			})
			if sendErr != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", sendErr)
			} else {
				fmt.Printf("\nAI: %s\n\n", reply)
			}
		}

		if err != nil {
			if err == io.EOF {
				fmt.Println()
				return nil
			}
			return err
		}
	}
}
