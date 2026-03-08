# kusa 内部設計（Go 言語）

本ドキュメントは、kusa CLI アプリケーションの内部設計を定義する。対象 OS は Linux / Windows / macOS、実装言語は Go とする。

## アーキテクチャ概要

- CLI 層: 引数解釈、標準入出力、REPL 制御
- App（Usecase）層: ユースケース実装、状態遷移、エラー整形
- Ports 層: App が依存する抽象インタフェース
- Infra 層: HTTP クライアント、永続ストレージ実装
- Domain 層: エンティティ、値オブジェクト

依存方向: CLI -> App -> Ports <- Infra

## ディレクトリ構成

cmd/kusa/main.go  
internal/cli  
internal/app  
internal/ports  
internal/domain  
internal/infra/httpclient/kusaapi  
internal/infra/storage/file

## CLI 設計

- ライブラリ: cobra
- コマンド:
  - kusa login
  - kusa chat new --model <name> --tool <tool> [--tool <tool>]
  - kusa chat list
  - kusa chat delete <id>
  - kusa model
  - kusa tool
- コマンド未指定時: ヘルプ表示して終了
- 未ログイン時: メッセージ出力して終了
- kusa chat 実行時: ログイン済みかつチャット開始済みなら標準入力を受け付け

## HTTP クライアント設計

- net/http + cookiejar を使用
- cookiejar を XDG Cache に永続化
- OpenAPI 仕様に基づくクライアント実装（未実装 API は将来対応）
- 会話新規作成は /api/chat の引数違いで実現

## Cookie 永続化

- 保存先: XDG Cache Home / kusa/http/cookies.bin
- cookies を保存
- 起動時に cookies.bin から復元
- 保存タイミング: プログラム終了時
- 権限: ディレクトリ 0700 / ファイル 0600

## 会話セッション保存

- 保存先: XDG Cache Home / kusa/conversation.json
- 保存対象:
  - 現在の会話セッション ID
- 保存タイミング: 新規会話セッション開始時

## Ports 定義（要約）

- ExternalAPIClient

  - Login
  - SendChat
  - GetConversationOverview
  - ListTools
  - ListModels（将来）
  - DeleteConversation（将来）

- Storage
  - SaveActiveSession / LoadActiveSession / ClearActiveSession

## エラーハンドリング

- 401/403: 未ログイン扱い
- Cookie 破損: Cookie 初期化 + 再ログイン要求
- Cookie 期限切れ: API 失敗時に未ログイン扱い

## セキュリティ

- 権限管理の徹底
- HTTP Client は単一インスタンスとして管理

## 拡張方針

- OpenAPI に API が追加され次第 infra/httpclient に実装を追加
- CLI コマンド追加時も Ports/App の境界は維持
