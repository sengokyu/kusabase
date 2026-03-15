# kusa 内部設計（Go 言語）

本ドキュメントは、kusa CLI アプリケーションの内部設計を定義する。対象 OS は Linux / Windows / macOS、実装言語は Go とする。

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
