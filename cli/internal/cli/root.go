package cli

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/sengokyu/kusabase/cli/internal/app"
	"github.com/sengokyu/kusabase/cli/internal/infra/httpclient/kusaapi"
	"github.com/sengokyu/kusabase/cli/internal/infra/storage/file"
)

// defaultBaseURL は KUSA_BASE_URL 未設定時に使う API のベース URL。
const defaultBaseURL = "https://gai.exabase.ai"

// globalOptions はルートコマンドのグローバルオプションをまとめた構造体。
type globalOptions struct {
	Debug bool // --debug: デバッグログを STDERR に出力する
}

// globalOpts はグローバルオプションの値を保持する。
var globalOpts globalOptions

// rootCmd はアプリケーションのルートコマンド。
var rootCmd = &cobra.Command{
	Use:   "kusa",
	Short: "kusa - AI チャット CLI クライアント",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&globalOpts.Debug, "debug", false, "デバッグログを STDERR に出力する")
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(newChatCmd())
	rootCmd.AddCommand(modelCmd)
	rootCmd.AddCommand(toolCmd)
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

// container holds all initialized use cases.
type container struct {
	login     *app.LoginUsecase // ログインユースケース
	chat      *app.ChatUsecase  // チャットユースケース
	tool      *app.ToolUsecase  // ツールユースケース
	model     *app.ModelUsecase // モデルユースケース
	apiClient *kusaapi.Client   // API クライアント（close 時に Cookie を保存）
}

// close saves cookies and releases resources.
func (c *container) close() {
	if err := c.apiClient.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "Cookie の保存に失敗しました: %v\n", err)
	}
}

// newContainer initializes all dependencies.
func newContainer() (*container, error) {
	baseURL := os.Getenv("KUSA_BASE_URL")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, fmt.Errorf("キャッシュディレクトリの取得に失敗しました: %w", err)
	}

	kusaDir := filepath.Join(cacheDir, "kusa")
	httpCacheDir := filepath.Join(kusaDir, "http")

	storage, err := file.NewStorage(kusaDir)
	if err != nil {
		return nil, fmt.Errorf("ストレージの初期化に失敗しました: %w", err)
	}

	apiClient, err := kusaapi.NewClient(baseURL, httpCacheDir, makeDebugFn())
	if err != nil {
		return nil, fmt.Errorf("API クライアントの初期化に失敗しました: %w", err)
	}

	return &container{
		login:     app.NewLoginUsecase(apiClient),
		chat:      app.NewChatUsecase(apiClient, storage, os.Stdin),
		tool:      app.NewToolUsecase(apiClient),
		model:     app.NewModelUsecase(apiClient),
		apiClient: apiClient,
	}, nil
}

func makeDebugFn() func(string, ...interface{}) {
	if globalOpts.Debug {
		logger := log.New(os.Stderr, "[DEBUG] ", log.LstdFlags)
		return func(format string, args ...interface{}) {
			logger.Printf(format, args...)
		}
	}
	return func(string, ...interface{}) {}
}

// printNotLoggedIn prints the standard "not logged in" message.
func printNotLoggedIn() {
	fmt.Println("You are not logged in.")
	fmt.Println("Run `kusa login` to continue.")
}

// printNoActiveSession prints the standard "no active chat session" message.
func printNoActiveSession() {
	fmt.Println("No active chat session.")
	fmt.Println("Run `kusa chat new` to start a new chat,")
	fmt.Println("or `kusa chat list` to resume an existing one.")
}

