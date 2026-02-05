# kvtool

設定ファイル（.env、JSON、YAML、Vault、S3 など）を統一的なインターフェースで扱う CLI ツール。

## 特徴

- **統一インターフェース**: ローカルファイル、Vault、S3 などを同じ方法でアクセス
- **マルチテナント**: namespace による環境切り替え
- **フォーマット変換**: dotenv ↔ JSON ↔ YAML
- **接続確認**: ストアの接続テスト機能

## Before / After

### Before: 複数ツールを個別に呼び出し

```bash
# Vault からシークレット取得
export VAULT_ADDR=http://localhost:8200
export VAULT_TOKEN=s.xxxxx
DB_PASS=$(vault kv get -field=password secret/prod/db)

# S3 から設定ファイル取得
aws s3 cp s3://my-bucket/config/app.json /tmp/app.json
APP_CONFIG=$(cat /tmp/app.json)

# ローカル .env 読み込み
export $(cat .env | xargs)

# 環境変数から取得
API_KEY=$API_KEY
```

### After: kvtool で統一アクセス

```bash
# 全て同じインターフェースで取得
kvtool store load vault/prod/db/password      # Vault
kvtool store load s3config/app.json           # S3
kvtool store load local/.env                  # ローカル .env
kvtool store load env/API_KEY                 # 環境変数
```

```yaml
# .kvtool.yml - 一度設定すれば、どこからでも同じ方法でアクセス
namespaces:
  default:
    vault:
      driver: vault
      args: { addr: "http://localhost:8200", token: "${VAULT_TOKEN}", mount: "secret" }
    s3config:
      driver: s3
      args: { bucket: "my-bucket", region: "us-east-1", root: "config" }
    local:
      driver: local
      args: { root: ".", transform: { read: "dotenv" } }
    env:
      driver: env
```

## インストール

```bash
# ビルドとインストール
make install

# または直接ビルド
go build -o bin/kvtool .
```

## クイックスタート

### 1. 設定ファイルの初期化

```bash
kvtool store init                 # カレントディレクトリに .kvtool.yml
kvtool store init --global        # ~/.config/kvtool/.kvtool.yml
```

### 2. ストアからデータ取得

```bash
kvtool store load .env/APP_NAME               # JSON形式で出力
kvtool store load .env/APP_NAME -o raw        # key=value形式
kvtool store load vault/db/password           # Vaultから取得
kvtool store load s3config/app.json           # S3から取得
```

### 3. 接続確認

```bash
kvtool store connect                          # 全ストアをテスト
kvtool store connect s3-test                  # 特定ストアのみ
```

### 4. ファイル直接操作（低レベル）

```bash
kvtool file load env                          # 環境変数をJSON化
kvtool file load vault <path> --addr ... --token ...
kvtool file convert dotenv                    # 標準入力からdotenv変換
```

## コマンド体系

```
kvtool
├── file (低レベル: 直接アクセス)
│   ├── load (env, vault, json, yaml, etc.)
│   └── convert (dotenv, yaml)
└── store (高レベル: 設定ファイル経由)
    ├── init
    ├── connect
    ├── load
    └── serve (未実装)
```

## 設定例

```yaml
version: 0.1
namespaces:
  default:
    local-env:
      driver: local
      args:
        root: ./config
        transform: dotenv
    vault-prod:
      driver: vault
      args:
        addr: http://localhost:8200
        token: root
        mount: secret
    s3-config:
      driver: s3
      args:
        bucket: my-bucket
        region: us-east-1
        endpoint: http://localhost:9000
```

## ドキュメント

- **[オンラインドキュメント](https://sasano8.github.io/kvtool/)**: mdBook による公式ドキュメントサイト
- **[API リファレンス](docs/api-reference.md)**: 設定パラメータ一覧（自動生成）
- **[CLAUDE.md](CLAUDE.md)**: 開発ガイド（コード規約、アーキテクチャ、暗黙知）
- **[S3 仕様](do../pkg/filesystems/s3.md)**: S3 ドライバーの詳細仕様

## 開発

```bash
# テスト
make test                         # 全テスト実行
make test-full                    # Vault含む統合テスト

# MinIO (S3互換) 起動
make minio-up                     # http://localhost:9000
make minio-down

# Vault 起動
make vault-up                     # http://localhost:8200
make vault-down

# ドキュメント生成
make doc-gen                      # docs/api-reference.md を生成
make book-build                   # mdBook をビルド
make book-serve                   # mdBook サーバー起動

# コード品質
make lint                         # 静的解析
make format                       # コードフォーマット
```

詳細は [CLAUDE.md](CLAUDE.md) を参照。

## ライセンス

MIT
