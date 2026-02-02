# 設定ファイル

kvtool の設定ファイル (`.kvtool.yml`) は YAML 形式で記述します。

## 基本構造

```yaml
version: 0.1
namespaces:
  <namespace名>:
    <ストア名>:
      driver: <ドライバー名>
      args:
        <ドライバー固有の設定>
```

## 設定ファイルの配置

kvtool は以下の順序で設定ファイルを探します：

1. カレントディレクトリの `.kvtool.yml`
2. `--config` フラグで指定されたパス
3. グローバル設定: `~/.config/kvtool/.kvtool.yml`

## namespace

namespace は環境（開発、ステージング、本番など）ごとに設定を切り替えるための仕組みです。

```yaml
namespaces:
  default:
    local-env:
      driver: local
      args:
        root: ./config
  production:
    vault-prod:
      driver: vault
      args:
        addr: https://vault.example.com
        token: ${VAULT_TOKEN}
```

デフォルトは `default` namespace が使用されます。

## ドライバー

kvtool は以下のドライバーをサポートしています：

- **local**: ローカルファイルシステム
- **vault**: HashiCorp Vault
- **env**: 環境変数
- **s3**: Amazon S3 / S3互換ストレージ

各ドライバーの詳細は [API リファレンス](./api-reference.md) を参照してください。

## 設定例

### ローカルファイル

```yaml
local-env:
  driver: local
  args:
    root: ./config
    transform: dotenv  # .env形式をJSON変換
```

### HashiCorp Vault

```yaml
vault-prod:
  driver: vault
  args:
    addr: http://localhost:8200
    token: root
    mount: secret
    namespace: admin
```

### Amazon S3

```yaml
s3-config:
  driver: s3
  args:
    bucket: my-config-bucket
    region: us-east-1
    endpoint: http://localhost:9000  # MinIOなどS3互換ストレージ用
    access_key_id: minioadmin
    secret_access_key: minioadmin
    use_path_style: true
```

### 環境変数

```yaml
env-vars:
  driver: env
  args: {}  # 設定不要
```
