# kvtool

kvtool は、設定ファイル（.env、Vault など）を統一的なインターフェースで扱うためのツールです。

## 特徴

- **ストア機能**: 複数の設定ソース（ローカルファイル、Vault など）を統一的に管理
- **柔軟な設定**: ローカル設定とグローバル設定の自動解決
- **マルチテナント**: namespace による環境の切り替え
- **複数フォーマット対応**: .env、JSON、YAML、Vault など
- **変換機能**: 各種フォーマット間の相互変換

## ドキュメント

### ユーザー向けドキュメント

- [設定ファイル仕様](docs/configuration.md) - 設定ファイルの詳細仕様と解決ロジック
- [コマンドリファレンス](docs/commands.md) - 全コマンドの使い方
- [設計思想・仕様](docs/design.md) - アーキテクチャと設計判断
- [API リファレンス](docs/api-reference.md) - 設定パラメータ一覧（自動生成）

### ファイルシステムドライバー仕様

- [S3 Filesystem](docs/filesystems/s3.md) - Amazon S3 および S3 互換ストレージ対応

### 開発者向けドキュメント

- [CLAUDE.md](CLAUDE.md) - AI アシスタント向けの開発ガイド（コード規約、ドキュメント規約など）

### ドキュメント生成

kvtool では、コード内のコメントと構造体タグからドキュメントを自動生成できます。

```bash
# ドキュメントを生成
make gen-docs
```

詳細は [CLAUDE.md](CLAUDE.md) の「コード規約」セクションを参照してください。

## インストール

```bash
make install
```

## クイックスタート

### 1. 設定ファイルの初期化

```bash
# ローカル設定ファイルを作成（カレントディレクトリに .kvtool.yml）
kvtool store init

# グローバル設定ファイルを作成（~/.config/kvtool/.kvtool.yml）
kvtool store init --global
```

生成される設定ファイル例：

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

### 2. 設定ファイルの取得

パス形式: `store_name/file_path`

```bash
# .env ファイルを JSON 形式で取得
kvtool store load .env/test.env

# raw 形式（key=value）で取得
kvtool store load .env/test.env -o raw

# 特定の namespace を指定
kvtool store load .env/test.env --namespace production
kvtool store load .env/test.env -n staging

# グローバル設定を使用
kvtool store load .env/test.env --global
```

## 設定ファイルの解決

kvtool はデフォルトでカレントディレクトリの `.kvtool.yml` を参照します。

グローバル設定を使用する場合は、明示的に指定する必要があります：

```bash
# グローバル設定を使用
kvtool store load .env/APP_NAME --global

# カスタムパスを指定
kvtool store load .env/APP_NAME -c /path/to/.kvtool.yml
```

**注意**: ローカルに `.kvtool.yml` がない場合、自動的にグローバル設定にフォールバックすることはありません。明示的に `--global` フラグまたは `--config` フラグを指定してください。

## ストアの設定

### ローカルファイルシステム

```yaml
version: 0.1
namespaces:
  default:
    .env:
      driver: local
      args:
        root: .              # ルートディレクトリ
        transform:
          read: dotenv       # 読み込み時の変換方法
          write: dotenv      # 書き込み時の変換方法
      mount:
        file: ""
```

### Vault

```yaml
version: 0.1
namespaces:
  default:
    vault:
      driver: vault
      args:
        conn:
          addr: "http://localhost:8200"
          token: root
          mount: secret
        root: app/prod       # Vault 内のルートパス
```

## マルチテナント（namespace）

異なる環境（開発、本番など）を namespace で切り替えることができます：

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
```

使用例：

```bash
# 開発環境の設定を取得
kvtool store load .env/app.env -n development

# 本番環境の設定を取得
kvtool store load .env/app.env -n production
```

## 出力形式

```bash
# JSON 形式（デフォルト）
kvtool store load .env/test.env
# 出力:
# {
#   "APP_NAME": "myapp",
#   "APP_ENV": "production"
# }

# raw 形式（key=value）
kvtool store load .env/test.env -o raw
# 出力:
# APP_NAME=myapp
# APP_ENV=production
```

## その他のコマンド

### 環境変数を JSON に変換

```bash
kvtool load env
```

### .env ファイルを JSON に変換

```bash
kvtool load dotenv -i test_data/dot_env/simple.env
```

### Vault から直接読み込み

```bash
kvtool load vault -addr http://localhost:8200 -token root -mount secret app/prod
```

### JSON を他の形式に変換

```bash
cat test_data/json/simple.json | kvtool convert dotenv
cat test_data/json/simple.json | kvtool convert yaml
```

