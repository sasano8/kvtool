# 設定ファイル仕様

## 概要

kvtool は YAML 形式の設定ファイル（`.kvtool.yml`）を使用して、複数の設定ソースを統一的に管理します。

## 設定ファイルの解決

### 優先順位

kvtool は以下の優先順位で設定ファイルを探します：

1. `--config` フラグで明示的に指定されたパス
2. カレントディレクトリの `.kvtool.yml`

**重要な仕様**:
- **ローカル優先**: デフォルトでカレントディレクトリの `.kvtool.yml` を参照
- **グローバルは明示的**: グローバル設定（`~/.config/kvtool/.kvtool.yml`）は `--global` フラグまたは `--config` フラグで明示的に指定
- **自動フォールバックなし**: ローカルに `.kvtool.yml` がない場合、自動的にグローバル設定にフォールバックしない
- **エラーメッセージ**: ローカル設定がない場合、`kvtool store init` を実行するよう促すエラーメッセージを表示

### 使用例

```bash
# デフォルト: ローカル設定を使用
kvtool store load .env/test.env

# グローバル設定を明示的に使用
kvtool store load .env/test.env --global

# カスタムパスを指定
kvtool store load .env/test.env -c /path/to/custom.yml
```

## 設定ファイルの構造

### 基本構造

```yaml
version: 0.1
namespaces:
  <namespace_name>:
    <store_name>:
      driver: <driver_type>
      args:
        <driver_specific_args>
      mount:
        dir: <directory>   # オプション
        file: <file>       # オプション
```

### version

- **型**: float
- **説明**: 設定ファイルのバージョン
- **現在のバージョン**: 0.1

### namespaces

- **型**: map[string]map[string]StoreInfo
- **説明**: 名前空間の定義。マルチテナント環境で環境を切り替えるために使用

#### namespace の使い方

異なる環境（開発、ステージング、本番など）を namespace で切り替えます：

```yaml
version: 0.1
namespaces:
  development:
    .env:
      driver: local
      args:
        root: ./config/dev
        transform:
          read: dotenv
          write: dotenv

  production:
    .env:
      driver: local
      args:
        root: ./config/prod
        transform:
          read: dotenv
          write: dotenv

    vault:
      driver: vault
      args:
        conn:
          addr: "http://vault.prod.example.com:8200"
          token: ${VAULT_TOKEN}
          mount: secret
        root: app/prod
```

使用例：

```bash
# 開発環境
kvtool store load .env/app.env -n development

# 本番環境
kvtool store load .env/app.env -n production
kvtool store load vault/db/password -n production
```

### store

各 namespace 内で複数のストアを定義できます。

#### store_name

- **型**: string
- **説明**: ストアの識別子。パス指定時に使用
- **命名規則**: 任意の文字列（例: `.env`, `vault`, `config`）

#### driver

- **型**: string
- **必須**: はい
- **値**: `local` または `vault`
- **説明**: ストアのバックエンドドライバー

#### args

- **型**: map[string]interface{}
- **説明**: ドライバー固有の引数

#### mount

- **型**: MountInfo（オプション）
- **説明**: マウント設定（現在は未使用）

## ドライバー仕様

### local ドライバー

ローカルファイルシステムからファイルを読み込みます。

#### 必須パラメータ

なし（すべてオプション）

#### オプションパラメータ

| パラメータ | 型 | デフォルト | 説明 |
|-----------|-----|----------|------|
| root | string | "." | ルートディレクトリ。このディレクトリより上位には遡れない |
| transform.read | string | null | 読み込み時の変換方法（`dotenv`、`json` など） |
| transform.write | string | null | 書き込み時の変換方法（現在未使用） |

#### セキュリティ

- **パストラバーサル防止**: `root` で指定されたディレクトリより上位には遡れない
- **相対パスのみ**: ファイルパスは相対パスのみ受け付ける
- **絶対パス拒否**: 絶対パスや `~` を含むパスは拒否

#### 設定例

```yaml
.env:
  driver: local
  args:
    root: ./config          # ルートディレクトリ
    transform:
      read: dotenv          # .env 形式として読み込み
      write: dotenv
```

### vault ドライバー

HashiCorp Vault からシークレットを読み込みます。

#### 必須パラメータ

| パラメータ | 型 | 説明 |
|-----------|-----|------|
| conn.addr | string | Vault サーバーのアドレス（例: `http://localhost:8200`） |
| conn.token | string | Vault 認証トークン |
| conn.mount | string | KV マウントポイント（例: `secret`） |

#### オプションパラメータ

| パラメータ | 型 | デフォルト | 説明 |
|-----------|-----|----------|------|
| root | string | "" | Vault 内のルートパス |
| transform.read | string | null | 読み込み時の変換方法（現在未使用） |
| transform.write | string | null | 書き込み時の変換方法（現在未使用） |

#### 対応バージョン

- **KV v2 のみ対応**: 現在は KV v2 エンジンのみサポート

#### 設定例

```yaml
vault:
  driver: vault
  args:
    conn:
      addr: "http://localhost:8200"
      token: root
      mount: secret
    root: app/prod          # Vault 内のルートパス
```

使用例：

```bash
# vault の app/prod/db/password にアクセス
kvtool store load vault/db/password
```

## transform 仕様

### dotenv

- `.env` 形式のファイルを読み込み、JSON に変換
- 行頭コメント（`#`）をサポート
- クォート囲み（ダブル・シングル）をサポート
- エスケープシーケンス（`\n`, `\r`, `\t`, `\\`, `\"`）をサポート

### 制限事項

- **マルチライン値は非対応**: 複数行にわたる値は使用できない
- **エンコーディング**: UTF-8 のみサポート

## パス仕様

### パス形式

```
store_name/file_path
```

- **store_name**: 設定ファイルで定義したストア名
- **file_path**: ストア内のファイルパス（オプション）

### パース規則

- 先頭・末尾のスラッシュは無視される
- 連続するスラッシュは1つのスラッシュとして扱われる
- 空のパス部分は無視される

### 例

```bash
# 基本形式
kvtool store load .env/APP_NAME

# ネストしたパス
kvtool store load vault/app/prod/db/password

# ストア名のみ（ファイルパスなし）
kvtool store load .env

# 先頭にスラッシュがあっても OK
kvtool store load /.env/APP_NAME
```

## 設定ファイルの作成

### 初期化コマンド

```bash
# ローカル設定ファイルを作成
kvtool store init

# グローバル設定ファイルを作成
kvtool store init --global

# カスタムパスに作成
kvtool store init /path/to/config.yml

# 既存ファイルを上書き
kvtool store init --force
```

### デフォルト設定

`kvtool store init` で作成される設定：

```yaml
version: 0.1
namespaces:
  default:
    .env:
      driver: local
      args:
        root: .
        transform:
          read: dotenv
          write: dotenv
      mount:
        file: ""
```
