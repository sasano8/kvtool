# クイックスタート

## 1. 設定ファイルの初期化

```bash
# カレントディレクトリに .kvtool.yml を作成
kvtool store init

# グローバル設定ファイル (~/.config/kvtool/.kvtool.yml) を作成
kvtool store init --global
```

## 2. ストアからデータ取得

```bash
# .envストアからAPP_NAMEを取得（JSON形式で出力）
kvtool store load .env/APP_NAME

# key=value形式で出力
kvtool store load .env/APP_NAME -o raw

# Vaultから取得
kvtool store load vault/db/password

# S3から取得
kvtool store load s3config/app.json
```

## 3. 接続確認

```bash
# 設定ファイル内の全ストアをテスト
kvtool store connect

# 特定のストアのみテスト
kvtool store connect s3-test
```

## 4. ファイル直接操作（低レベル）

設定ファイルを使わず、直接ファイルシステムにアクセスする場合：

```bash
# 環境変数をJSON化
kvtool file load env

# Vaultから直接取得
kvtool file load vault <path> --addr http://localhost:8200 --token root

# 標準入力からdotenv変換
cat .env | kvtool file convert dotenv
```

## 設定例

`.kvtool.yml` の例：

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
