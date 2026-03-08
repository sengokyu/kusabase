package cli

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/sengokyu/kusabase/cli/internal/ports"
)

func newChatCmd() *cobra.Command {
	var modelFlag string
	var toolFlags []string

	chatCmd := &cobra.Command{
		Use:   "chat",
		Short: "チャットを操作する",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newContainer()
			if err != nil {
				return err
			}
			defer c.close()

			if err := c.chat.Continue(cmd.Context()); err != nil {
				if errors.Is(err, ports.ErrNotLoggedIn) {
					printNotLoggedIn()
					return nil
				}
				if errors.Is(err, ports.ErrNoActiveSession) {
					printNoActiveSession()
					return nil
				}
				return err
			}
			return nil
		},
	}

	chatNewCmd := &cobra.Command{
		Use:   "new",
		Short: "新しいチャットセッションを開始する",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newContainer()
			if err != nil {
				return err
			}
			defer c.close()

			if err := c.chat.StartNew(cmd.Context(), modelFlag, toolFlags); err != nil {
				if errors.Is(err, ports.ErrNotLoggedIn) {
					printNotLoggedIn()
					return nil
				}
				return err
			}
			return nil
		},
	}
	chatNewCmd.Flags().StringVar(&modelFlag, "model", "", "使用するモデル")
	chatNewCmd.Flags().StringArrayVar(&toolFlags, "tool", nil, "使用するツール（複数指定可）")

	chatListCmd := &cobra.Command{
		Use:   "list",
		Short: "チャットセッション一覧を表示する",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newContainer()
			if err != nil {
				return err
			}
			defer c.close()

			convs, err := c.chat.List(cmd.Context())
			if err != nil {
				if errors.Is(err, ports.ErrNotLoggedIn) {
					printNotLoggedIn()
					return nil
				}
				return err
			}

			if len(convs) == 0 {
				fmt.Println("チャット履歴がありません。")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tTITLE\tMODEL\tTOOLS\tUPDATED")
			for i, conv := range convs {
				tools := "-"
				if len(conv.ToolNames) > 0 {
					tools = strings.Join(conv.ToolNames, ",")
				}
				model := conv.ModelName
				if model == "" {
					model = "-"
				}
				updated := "-"
				if !conv.LastMessageAt.IsZero() {
					updated = conv.LastMessageAt.Local().Format("2006-01-02 15:04")
				}
				title := conv.Title
				if title == "" {
					title = "(無題)"
				}
				fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", i+1, title, model, tools, updated)
			}
			w.Flush()
			return nil
		},
	}

	chatDeleteCmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "チャットセッションを削除する",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil || id < 1 {
				return fmt.Errorf("無効なID: %s", args[0])
			}

			c, err := newContainer()
			if err != nil {
				return err
			}
			defer c.close()

			convs, err := c.chat.List(cmd.Context())
			if err != nil {
				if errors.Is(err, ports.ErrNotLoggedIn) {
					printNotLoggedIn()
					return nil
				}
				return err
			}

			if id > len(convs) {
				return fmt.Errorf("ID %d は存在しません（最大: %d）", id, len(convs))
			}

			conv := convs[id-1]
			if err := c.chat.Delete(cmd.Context(), conv.UUID); err != nil {
				return fmt.Errorf("削除に失敗しました: %w", err)
			}

			title := conv.Title
			if title == "" {
				title = conv.UUID
			}
			fmt.Printf("チャット \"%s\" をローカルから削除しました。\n", title)
			fmt.Println("注意: サーバー上の会話はまだ削除されていません（API 未対応）。")
			return nil
		},
	}

	chatCmd.AddCommand(chatNewCmd, chatListCmd, chatDeleteCmd)
	return chatCmd
}
