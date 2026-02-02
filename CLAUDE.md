# kvtool - Claude AI 開発ガイド

## プロジェクト概要

**kvtool** は、様々な設定ファイル形式（.env, JSON, YAML, HCL, Vault, S3, データベースなど）を統一的なインターフェースで扱う CLI ツールです。

### 主要ユースケース

環境構築の初期段階で、様々な構成ファイルを統合して環境を初期化することを想定しています。

### アーキテクチャ

- **レイヤー**: Presentation（CLI） → Service → Infrastructure（Filesystem）
- **言語**: Go 1.21+
- **テスト**: テーブル駆動テスト、統合テスト
- **ドキュメント**: 今後 mdBook を導入予定

## 前提条件・実装済み機能

### 実装済みドライバー

- **LocalFs**: ローカルファイルシステム ([filesystems/local.go](filesystems/local.go))
- **VaultFs**: HashiCorp Vault ([filesystems/vault.go](filesystems/vault.go))
- **FsEnvFilesystem**: 環境変数 ([filesystems/env_fs.go](filesystems/env_fs.go))

### 実装済み Transform

- **dotenv**: .env 形式を JSON に変換 ([pkg/decoders/env_to_json.go](pkg/decoders/env_to_json.go))

### 統一インターフェース

全てのファイルシステムは以下のインターフェースを実装します ([filesystems/core.go](filesystems/core.go)):

```go
type Filesystem interface {
	GetFile(path string) (File, error)
}

type File interface {
	LoadAsJson() (any, error)
	OpenReader() (io.ReadCloser, error)
}
```

### テスト

- **統合テスト**: [filesystems/integration_test.go](filesystems/integration_test.go) で全ファイルシステムの共通動作を検証
- **テスト言語**: 日本語（コメント、テスト名）

## コード規約

### ドキュメントコメント

#### パッケージコメント

各ファイルシステムのメインファイルの先頭に、以下の形式でドキュメントコメントを記述します。

```go
// Package filesystems provides unified filesystem interfaces.
//
// # S3 Filesystem
//
// S3 ファイルシステムは、Amazon S3 および S3 互換ストレージからファイルを読み込みます。
//
// ## Driver Name
//
// s3
//
// ## Use Cases
//
// - CI/CD パイプラインでの設定取得
// - 複数環境での設定の中央管理
// - S3 互換ストレージ（MinIO, Cloudflare R2）との統合
//
// ## Example Configuration
//
//     version: 0.1
//     namespaces:
//       default:
//         s3config:
//           driver: s3
//           args:
//             bucket: my-config-bucket
//             region: ap-northeast-1
//
// ## Security
//
// - パストラバーサル攻撃の防止
// - 相対パスのみサポート
package filesystems
```

#### 構造体のドキュメント

設定構造体には、`doc`, `required`, `default`, `example` タグを使ってドキュメントを埋め込みます。

```go
// S3FsConfig は S3 ファイルシステムの設定
type S3FsConfig struct {
	// Bucket は S3 バケット名（必須）
	Bucket string `yaml:"bucket" doc:"S3 バケット名" required:"true" example:"my-config-bucket"`

	// Region は AWS リージョン（必須）
	Region string `yaml:"region" doc:"AWS リージョン" required:"true" example:"ap-northeast-1"`

	// Root はバケット内のルートパス（オプション）
	Root string `yaml:"root" doc:"バケット内のルートパス" required:"false" default:"" example:"config/production"`

	// Timeout はリクエストタイムアウト（オプション）
	Timeout string `yaml:"timeout" doc:"リクエストタイムアウト" required:"false" default:"30s" example:"60s"`
}
```

#### タグの説明

| タグ名 | 説明 | 例 |
|--------|------|-----|
| `yaml` | YAML フィールド名 | `yaml:"bucket"` |
| `doc` | フィールドの説明文 | `doc:"S3 バケット名"` |
| `required` | 必須かどうか | `required:"true"` |
| `default` | デフォルト値 | `default:"30s"` |
| `example` | 設定例 | `example:"my-bucket"` |

#### 関数/メソッドのドキュメント

公開関数・メソッドには godoc 形式のコメントを記述します。

```go
// NewS3Fs は S3 ファイルシステムを作成します。
//
// 認証情報は以下の優先順位で取得されます：
//   1. 設定ファイルでの明示的な指定
//   2. 環境変数（AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY）
//   3. IAM ロール（EC2, ECS などで実行時）
//
// エラーが発生する場合：
//   - 設定が不正な場合
//   - AWS 認証情報が見つからない場合
func NewS3Fs(ctx context.Context, config S3FsConfig) (*S3Fs, error) {
	// 実装
}
```

### セキュリティ規約

#### パストラバーサル防止

全てのファイルシステムは `root` より上位に遡れないようにする必要があります。

```go
// ✅ 良い例
func (fs *LocalFs) GetFile(path string) (File, error) {
	// パストラバーサルチェック
	fullPath := filepath.Join(fs.Root, path)
	cleanPath := filepath.Clean(fullPath)
	if !strings.HasPrefix(cleanPath, fs.Root) {
		return nil, fmt.Errorf("path escapes root: %s", path)
	}
	// ...
}
```

#### 相対パスのみ

絶対パス（`/etc/passwd`）やチルダパス（`~/config`）は拒否します。

```go
// ✅ 良い例
if filepath.IsAbs(path) || strings.HasPrefix(path, "~") {
	return nil, fmt.Errorf("path must be relative: %s", path)
}
```

### テスト規約

#### テーブル駆動テスト

全てのテストはテーブル駆動テストで記述します。

```go
func TestSomething(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:        "正常系",
			input:       "test",
			expected:    "test",
			expectError: false,
		},
		{
			name:        "エラー系",
			input:       "",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			// テスト実装
		})
	}
}
```

#### 日本語でのテスト記述

テスト名、コメント、エラーメッセージは日本語で記述します。

```go
// ✅ 良い例
{
	name: "LocalFs - パストラバーサル攻撃の防止",
	// ...
}

// ❌ 悪い例
{
	name: "LocalFs - prevent path traversal attack",
	// ...
}
```

## ドキュメント生成

### 自動生成の仕組み

コード内のコメントと構造体タグから API リファレンスを自動生成します。

```bash
# ドキュメントを生成
make gen-docs
```

実装: [scripts/gen-docs.go](scripts/gen-docs.go), [Makefile](Makefile) の `gen-docs` ターゲット

### 生成されるファイル

- `docs/api-reference.md` - API リファレンス（全ドライバーの設定パラメータ一覧）

### 生成内容

以下の情報がコードから自動抽出されます：

1. **ドライバー名**: パッケージコメントから抽出
2. **説明**: パッケージコメントから抽出
3. **設定パラメータ表**: 構造体タグから生成（`doc`, `required`, `default`, `example` タグ）
4. **設定例**: `example` タグから生成
5. **セキュリティ情報**: パッケージコメントの "Security" セクション
6. **ユースケース**: パッケージコメントの "Use Cases" セクション

### ドキュメント更新ワークフロー

1. コードを変更したらコメントとタグを更新
2. `make gen-docs` を実行
3. 生成されたドキュメントを確認
4. コミット前に `git add docs/`

## 開発ワークフロー

### 新しいファイルシステムドライバーの追加

1. **仕様策定**: `docs/filesystems/{driver}.md` を作成（手動）
2. **構造体定義**: 設定構造体を定義（`doc`, `required`, `default`, `example` タグを含む）
3. **インターフェース実装**: `Filesystem` と `File` インターフェースを実装
4. **テスト**: `filesystems/integration_test.go` にテストケースを追加
5. **ドキュメント生成**: `go run scripts/gen-docs.go` を実行
6. **TODO 更新**: `TODO.md` の該当項目をチェック

### 新しい Transform の追加

1. **パーサー実装**: `pkg/decoders/` に変換ロジックを実装
2. **テスト**: テーブル駆動テストを追加
3. **LocalFs に統合**: `LocalFile.LoadAsJson()` で Transform を適用
4. **ドキュメント**: `TODO.md` に実装状況を記載

## 今後の計画

### ドキュメント整備

- **mdBook 導入**: 公式ドキュメントサイトを構築予定
- **自動生成の拡充**: コードから API リファレンスを完全自動生成

### 新ドライバー

- **S3 Filesystem**: Amazon S3 および S3 互換ストレージ対応（仕様策定済み: [docs/filesystems/s3.md](docs/filesystems/s3.md)）
- **HTTP Filesystem**: REST API からの設定取得
- **Database Filesystem**: ユーザー定義 SQL クエリでの設定取得

### 新 Transform

- **HCL**: Terraform/Nomad/Consul 設定ファイルの読み込み

## 参考資料

- [TODO.md](TODO.md) - タスク管理・実装計画
- [docs/configuration.md](docs/configuration.md) - 設定ファイル仕様
- [docs/commands.md](docs/commands.md) - コマンドリファレンス
- [filesystems/core.go](filesystems/core.go) - 統一インターフェース定義
- [filesystems/integration_test.go](filesystems/integration_test.go) - 統合テスト