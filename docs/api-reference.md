# API リファレンス

このドキュメントは自動生成されています。手動で編集しないでください。

生成コマンド: `go run scripts/gen-docs.go`

## LocalFsConfig

ファイル: [local.go](pkg/filesystems/local.go)

| パラメータ | 型 | 必須 | デフォルト | 説明 |
|-----------|-----|:----:|----------|------|
| timeout | time.Duration |  | 30s | リクエストタイムアウト |
| root | string |  | . | ルートディレクトリ。このパスより上位には遡れない |
| transform | string |  | json | 読み込み時の変換方法（dotenv, json など） |

**設定例:**

```yaml
localfs:
  driver: local
  args:
    timeout: 60s
    root: ./config
    transform: dotenv
```

---

## S3FsConfig

ファイル: [s3.go](pkg/filesystems/s3.go)

| パラメータ | 型 | 必須 | デフォルト | 説明 |
|-----------|-----|:----:|----------|------|
| bucket | string | ✓ | - | S3 バケット名 |
| region | string | ✓ | - | AWS リージョン |
| root | string |  | - | バケット内のルートパス。このパスより上位には遡れない |
| access_key_id | string |  | - | AWS アクセスキー ID（省略時は環境変数または IAM ロールから取得） |
| secret_access_key | string |  | - | AWS シークレットアクセスキー（省略時は環境変数または IAM ロールから取得） |
| session_token | string |  | - | AWS セッショントークン（一時的な認証情報使用時） |
| endpoint | string |  | - | カスタムエンドポイント（MinIO など S3 互換ストレージ用） |
| use_path_style | bool |  | false | パススタイルアクセスを使用（S3 互換ストレージで必要な場合あり） |
| transform | string |  | - | 読み込み時の変換方法（dotenv, json など） |
| timeout | time.Duration |  | 30s | リクエストタイムアウト |

**設定例:**

```yaml
s3fs:
  driver: s3
  args:
    bucket: my-config-bucket
    region: ap-northeast-1
    root: config/production
    access_key_id: AKIAIOSFODNN7EXAMPLE
    secret_access_key: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
    session_token: FwoGZXIvYXdzE...
    endpoint: http://localhost:9000
    use_path_style: true
    transform: dotenv
    timeout: 60s
```

---

## VaultConfig

ファイル: [vault.go](pkg/filesystems/vault.go)

| パラメータ | 型 | 必須 | デフォルト | 説明 |
|-----------|-----|:----:|----------|------|
| addr | string | ✓ | - | Vault サーバーのアドレス |
| token | string |  | - | Vault 認証トークン（省略時は VAULT_TOKEN 環境変数から取得） |
| namespace | string |  | - | Vault 名前空間（Enterprise 機能） |
| mount | string | ✓ | secret | KV シークレットエンジンのマウントパス |
| kv_ver | int |  | 2 | KV シークレットエンジンのバージョン（現在は 2 のみサポート） |
| version | int |  | 0 | 取得するシークレットのバージョン（0 = 最新） |
| field | string |  | - | 取得する特定のフィールド名（省略時は全フィールド） |
| pretty | bool |  | false | JSON 出力を整形するか |
| timeout | time.Duration |  | 30s | リクエストタイムアウト |

**設定例:**

```yaml
vault:
  driver: vault
  args:
    addr: http://localhost:8200
    token: root
    namespace: admin
    mount: secret
    timeout: 60s
```

---

