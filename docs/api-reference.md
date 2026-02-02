# API リファレンス

このドキュメントは自動生成されています。手動で編集しないでください。

生成コマンド: `go run scripts/gen-docs.go`

## LocalFsConfig

ファイル: [local.go](p../pkg/filesystems/local.go)

### オプションパラメータ

| パラメータ | 型 | デフォルト | 説明 | 設定例 |
|-----------|-----|----------|------|--------|
|  | time.Duration | - |  | - |
|  | string | - |  | - |
|  | string | - |  | - |

---

## S3FsConfig

ファイル: [s3.go](p../pkg/filesystems/s3.go)

### 必須パラメータ

| パラメータ | 型 | 説明 | 設定例 |
|-----------|-----|------|--------|
| bucket | string | S3 バケット名 | my-config-bucket |
| region | string | AWS リージョン | ap-northeast-1 |

### オプションパラメータ

| パラメータ | 型 | デフォルト | 説明 | 設定例 |
|-----------|-----|----------|------|--------|
| root | string | - | バケット内のルートパス。このパスより上位には遡れない | config/production |
| access_key_id | string | - | AWS アクセスキー ID（省略時は環境変数または IAM ロールから取得） | AKIAIOSFODNN7EXAMPLE |
| secret_access_key | string | - | AWS シークレットアクセスキー（省略時は環境変数または IAM ロールから取得） | wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY |
| session_token | string | - | AWS セッショントークン（一時的な認証情報使用時） | FwoGZXIvYXdzE... |
| endpoint | string | - | カスタムエンドポイント（MinIO など S3 互換ストレージ用） | http://localhost:9000 |
| use_path_style | bool | false | パススタイルアクセスを使用（S3 互換ストレージで必要な場合あり） | true |
| transform | string | - | 読み込み時の変換方法（dotenv, json など） | dotenv |
| timeout | time.Duration | 30s | リクエストタイムアウト | 60s |

### 設定例

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

## FilesystemConfig

ファイル: [vault.go](p../pkg/filesystems/vault.go)

### オプションパラメータ

| パラメータ | 型 | デフォルト | 説明 | 設定例 |
|-----------|-----|----------|------|--------|
|  | string | - |  | - |
|  | map[string]any | - |  | - |

---

## VaultConfig

ファイル: [vault.go](p../pkg/filesystems/vault.go)

### オプションパラメータ

| パラメータ | 型 | デフォルト | 説明 | 設定例 |
|-----------|-----|----------|------|--------|
|  | string | - |  | - |
|  | string | - |  | - |
|  | string | - |  | - |
|  | string | - |  | - |
|  | int | - |  | - |
|  | int | - |  | - |
|  | string | - |  | - |
|  | bool | - |  | - |
|  | time.Duration | - |  | - |

---

