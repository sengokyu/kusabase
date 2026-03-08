# kusa CLI – 外部設計（v1.0）

## 概要

`kusa` は、CLI 上から対話・モデル管理・ツール管理を行うための  
マルチコマンド型 CLI アプリケーションである。  
Linux / Windows 両対応とする。

## Usage

```sh
kusa [options] [command] [command options...]
```

### グローバルオプション

- `--debug`:
  すべての処理に対してデバッグログを STDERR に出力する。  
  通常出力（会話結果など）は STDOUT を使用する。  
  機微情報（Cookie 値、認証情報等）は出力しない。

---

## 基本動作仕様

### コマンド未指定時の挙動

```sh
kusa
```

- コマンドが指定されていない場合、ヘルプメッセージを表示して終了する。

---

### 未ログイン時の挙動

- ログインが必要なコマンドを、未ログイン状態で実行した場合:
  - エラーメッセージを表示して終了する。
  - 例:
    ```txt
    You are not logged in.
    Run `kusa login` to continue.
    ```

対象コマンド:

- `kusa chat`
- `kusa chat new`
- `kusa chat list`
- `kusa chat delete`
- `kusa model`
- `kusa tool`

---

### `kusa chat` 単体実行時の挙動

```sh
kusa chat
```

- 条件:

  - ログイン済み
  - アクティブなチャットセッションが存在する

- 挙動:

  - 標準入力を受け付け、対話モードを継続する。

- 条件を満たさない場合:
  - 未ログイン:
    ```txt
    You are not logged in.
    Run `kusa login` to continue.
    ```
  - チャット未開始:
    ```txt
    No active chat session.
    Run `kusa chat new` to start a new chat,
    or `kusa chat list` to resume an existing one.
    ```

---

## コマンド構成

### 1. login

```sh
kusa login
```

- 認証情報の設定・保存を行う。

---

### 2. chat

```sh
kusa chat [subcommand] [options]
```

#### サブコマンド

```sh
kusa chat new [options]
kusa chat list
kusa chat delete <id>
```

##### `kusa chat new`

- 新しいチャットセッションを開始する。
- 対話モードを起動する。

###### オプション

```sh
--model <model_name>
```

- 使用するモデルを指定する。
- 未指定時はデフォルトモデルを使用する。

```sh
--tool <tool_name>
```

- 使用するツールを指定する。
- 複数指定可能。
- 例:
  ```sh
  kusa chat new --tool web_search --tool file_read
  ```

---

##### `kusa chat list`

- 保存されているチャットセッションの一覧を表示する。

表示例（案）:

```txt
ID   MODEL        TOOLS                 UPDATED
1    gpt-4.1      web_search,file_read  2026-02-20 10:12
2    o4-mini      -                     2026-02-19 22:01
```

---

##### `kusa chat delete <id>`

- 指定したチャットセッションを削除する。
- `<id>` は `kusa chat list` で表示される識別子。

---

### 3. model

```sh
kusa model
```

- 利用可能なモデルの一覧を表示する。

表示例（案）:

```txt
Available Models:
- gpt-4.1
- gpt-4.1-mini
- o4-mini
```

---

### 4. tool

```sh
kusa tool
```

- 利用可能なツールの一覧を表示する。

表示例（案）:

```txt
Available Tools:
- web_search   : Search the web
- file_read    : Read local files
- shell_exec   : Execute shell commands
```

---

## 出力仕様

- 標準出力に人間が読みやすい形式で表示する。
- 将来的に `--json` オプションによる機械可読出力も検討する。

## エラーハンドリング

- 不正なコマンド・引数の場合は usage を表示する。
- API 通信エラー時は内容を簡潔に表示する。

## 対応 OS

- Linux
- Windows
