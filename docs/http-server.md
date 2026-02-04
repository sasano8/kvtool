# HTTP サーバー

`kvtool store serve` コマンドで、設定ファイルで定義したストアを REST API として公開できます。

## 起動方法

```bash
# デフォルト設定（ポート 8080）
kvtool store serve

# ポート指定
kvtool store serve --port 3000

# グローバル設定を使用
kvtool store serve --global
```

## エンドポイント

### KV エンドポイント（JSON）

```
GET /v1/namespaces/{namespace}/kv/{store}/{path}
```

ファイルを JSON としてパースして返します。

```bash
# production 環境の vault ストアから app/settings を取得
curl http://localhost:8080/v1/namespaces/production/kv/vault/app/settings

# default 環境の local ストアから config.json を取得
curl http://localhost:8080/v1/namespaces/default/kv/local/config.json
```

### Storage エンドポイント（Raw）

```
GET /v1/namespaces/{namespace}/storage/{store}/{path}
```

ファイルをそのまま（`application/octet-stream`）返します。

```bash
# production 環境の s3 ストアから data.bin を取得
curl http://localhost:8080/v1/namespaces/production/storage/s3/data.bin
```

### ヘルスチェック

```
GET /health
```

```bash
curl http://localhost:8080/health
# => ok
```

## 設定例

```yaml
version: 1.0
namespaces:
  production:
    vault:
      driver: vault
      args:
        address: https://vault.example.com
        token_env: VAULT_TOKEN
        mount: secret
        root: myapp
    s3:
      driver: s3
      args:
        bucket: prod-configs
        region: ap-northeast-1

  development:
    local:
      driver: local
      args:
        root: ./configs/dev
```

この設定で以下のエンドポイントが利用可能になります:

- `/v1/namespaces/production/kv/vault/{path}`
- `/v1/namespaces/production/kv/s3/{path}`
- `/v1/namespaces/development/kv/local/{path}`

## エラーレスポンス

| ステータス | 説明 |
|-----------|------|
| 404 | namespace、store、またはファイルが存在しない |
| 405 | GET 以外のメソッド |
| 500 | サーバー内部エラー |

## CLI オプション

| オプション | 短縮 | デフォルト | 説明 |
|-----------|------|-----------|------|
| `--config` | `-c` | `.kvtool.yml` | 設定ファイルパス |
| `--global` | `-g` | `false` | グローバル設定を使用 |
| `--port` | `-p` | `8080` | リッスンポート |
