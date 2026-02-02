# アーキテクチャ

kvtool のアーキテクチャ概要です。

## レイヤー構造

```
Presentation (CLI)
    ↓
Service
    ↓
Infrastructure (Filesystem)
```

- **Presentation**: Cobra を使った CLI インターフェース (`cmd/`)
- **Service**: ビジネスロジック
- **Infrastructure**: ファイルシステムドライバー (`pkg/filesystems/`)

## 統一インターフェース

全てのファイルシステムは以下のインターフェースを実装します ([p../pkg/filesystems/core.go](../pkg/filesystems/core.go)):

```go
type Filesystem interface {
    GetFile(path string) (File, error)
}

type File interface {
    LoadAsJson() (any, error)
    OpenReader() (io.ReadCloser, error)
}
```

## ファイルシステムドライバー

### 実装済みドライバー

- **LocalFs**: ローカルファイルシステム ([p../pkg/filesystems/local.go](../pkg/filesystems/local.go))
- **VaultFs**: HashiCorp Vault ([p../pkg/filesystems/vault.go](../pkg/filesystems/vault.go))
- **FsEnvFilesystem**: 環境変数 ([p../pkg/filesystems/env_fs.go](../pkg/filesystems/env_fs.go))
- **S3Fs**: Amazon S3 / S3互換ストレージ ([p../pkg/filesystems/s3.go](../pkg/filesystems/s3.go))

### ドライバーの作成

`FilesystemFactory` がドライバー名から適切なファイルシステムインスタンスを生成します ([p../pkg/filesystems/factory.go](../pkg/filesystems/factory.go))。

## Transform（変換機能）

ファイルを読み込む際に、特定の形式（.env など）を JSON に変換する機能です。

### 実装済み Transform

- **dotenv**: .env 形式を JSON に変換 ([pkg/decoders/env_to_json.go](../pkg/decoders/env_to_json.go))

## テスト戦略

- **統合テスト**: [p../pkg/filesystems/integration_test.go](../pkg/filesystems/integration_test.go) で全ファイルシステムの共通動作を検証
- **テーブル駆動テスト**: 全てのテストはテーブル駆動テストで記述
- **Docker 統合**: Vault などの外部サービスは Docker で起動してテスト

詳細は [CLAUDE.md](../CLAUDE.md) を参照してください。
