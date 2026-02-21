# ドライバー

このドキュメントは自動生成されています。手動で編集しないでください。

生成コマンド: `go run scripts/gen-docs.go`

## LocalFsConfig

LocalFsConfig はローカルファイルシステムの設定

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

S3FsConfig は S3 ファイルシステムの設定

ファイル: [s3.go](pkg/filesystems/s3.go)

| パラメータ | 型 | 必須 | デフォルト | 説明 |
|-----------|-----|:----:|----------|------|
| bucket | string | ✓ | - | S3 バケット名 |
| region | string | ✓ | - | AWS リージョン |
| access_key_id | string | ✓ | - | AWS アクセスキー ID |
| secret_access_key | string | ✓ | - | AWS シークレットアクセスキー |
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

VaultConfig は HashiCorp Vault ファイルシステムの設定

ファイル: [vault.go](pkg/filesystems/vault.go)

| パラメータ | 型 | 必須 | デフォルト | 説明 |
|-----------|-----|:----:|----------|------|
| addr | string | ✓ | - | Vault サーバーのアドレス |
| token | string | ✓ | - | Vault 認証トークン |
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

## RestFsConfig

RestFsConfig は REST ファイルシステムの設定

ファイル: [rest.go](pkg/filesystems/rest.go)

| パラメータ | 型 | 必須 | デフォルト | 説明 |
|-----------|-----|:----:|----------|------|
| base_url | string | ✓ | - | ベース URL |
| auth_type | string |  | - | 認証タイプ（bearer, basic） |
| token | string |  | - | Bearer トークン |
| token_file | string |  | - | Bearer トークンファイルパス |
| username | string |  | - | Basic 認証ユーザー名 |
| password | string |  | - | Basic 認証パスワード |
| ca_file | string |  | - | CA 証明書ファイルパス |
| insecure | bool |  | false | TLS 証明書検証をスキップ |
| timeout | time.Duration |  | 30s | リクエストタイムアウト |

**設定例:**

```yaml
restfs:
  driver: rest
  args:
    base_url: https://api.example.com
    auth_type: bearer
    token_file: /var/run/secrets/token
```

---

## DbFsConfig

DbFsConfig はデータベースファイルシステムの設定

ファイル: [db.go](pkg/filesystems/db.go)

| パラメータ | 型 | 必須 | デフォルト | 説明 |
|-----------|-----|:----:|----------|------|
| connection_string | string | ✓ | - | データベース接続文字列 |
| driver | string |  | - | データベースドライバー（postgres, mysql, sqlite）。省略時は接続文字列から自動判定 |
| query | string | ✓ | - | SQL クエリ。{key} と {namespace} がプレースホルダーとして使用可能 |
| timeout | time.Duration |  | 10s | クエリタイムアウト |
| namespace | string |  | default | デフォルトの namespace 値 |

**設定例:**

```yaml
dbfs:
  driver: db
  args:
    connection_string: postgres://user:pass@localhost/db
    query: SELECT value FROM config WHERE key = {key}
    timeout: 30s
    namespace: production
```

---

## NatsFsConfig

NatsFsConfig は NATS JetStream KV ファイルシステムの設定

ファイル: [nats.go](pkg/filesystems/nats.go)

| パラメータ | 型 | 必須 | デフォルト | 説明 |
|-----------|-----|:----:|----------|------|
| url | string | ✓ | - | NATS サーバーの接続 URL |
| bucket | string | ✓ | - | JetStream KV バケット名 |
| token | string |  | - | NATS 認証トークン |
| user | string |  | - | NATS ユーザー名 |
| password | string |  | - | NATS パスワード |
| creds_file | string |  | - | NATS 認証情報ファイルパス |
| timeout | time.Duration |  | 10s | 接続およびリクエストタイムアウト |

**設定例:**

```yaml
natsfs:
  driver: nats
  args:
    url: nats://localhost:4222
    bucket: config
    timeout: 30s
```

---

## RedisFsConfig

RedisFsConfig は Redis ファイルシステムの設定

ファイル: [redis.go](pkg/filesystems/redis.go)

| パラメータ | 型 | 必須 | デフォルト | 説明 |
|-----------|-----|:----:|----------|------|
| addr | string | ✓ | - | Redis サーバーのアドレス |
| password | string |  | - | Redis 認証パスワード |
| db | int |  | 0 | Redis データベース番号 |
| prefix | string |  | - | キーのプレフィックス |
| timeout | time.Duration |  | 10s | 接続およびリクエストタイムアウト |

**設定例:**

```yaml
redisfs:
  driver: redis
  args:
    addr: localhost:6379
    db: 0
    prefix: config:
    timeout: 30s
```

---

## ToolFsConfig

ToolFsConfig は Tool ファイルシステムの設定
設定パラメータは不要です。
利用可能なパス: uuid7, uuid4, now, timestamp, random/hex/{bytes}, random/base64/{bytes}, password

ファイル: [tool.go](pkg/filesystems/tool.go)

---

