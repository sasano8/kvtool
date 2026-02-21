# kvtool - Claude AI 開発ガイド

## プロジェクト概要

設定ファイル（.env, JSON, YAML, Vault, S3 など）を統一的なインターフェースで扱う CLI ツール。

### アーキテクチャ

- **レイヤー**: CLI → Service → Filesystem（統一インターフェース）
- **言語**: Go 1.21+
- **テスト**: テーブル駆動テスト、統合テスト（日本語）

### コマンド体系

```
kvtool
├── file (低レベル: 直接アクセス)
│   ├── load (env, vault, json, yaml, etc.)
│   └── convert (dotenv, yaml)
└── store (高レベル: 設定ファイル経由)
    ├── init
    ├── connect
    ├── load
    ├── sync
    └── serve (未実装)
```

## 実装済み機能

### ドライバー

- **LocalFs**: ローカルファイルシステム ([p../pkg/filesystems/local.go](p../pkg/filesystems/local.go))
- **VaultFs**: HashiCorp Vault ([p../pkg/filesystems/vault.go](p../pkg/filesystems/vault.go))
- **S3Fs**: Amazon S3 / MinIO ([p../pkg/filesystems/s3.go](p../pkg/filesystems/s3.go))
- **EnvFs**: 環境変数 ([p../pkg/filesystems/env_fs.go](p../pkg/filesystems/env_fs.go))

### Transform

- **dotenv**: .env 形式を JSON に変換 ([pkg/decoders/env_to_json.go](pkg/decoders/env_to_json.go))

### 統一インターフェース

全ドライバーは以下を実装 ([p../pkg/filesystems/core.go](p../pkg/filesystems/core.go)):

```go
type Filesystem interface {
    GetFile(path string) (File, error)
}

type File interface {
    LoadAsJson() (any, error)
    OpenReader() (io.ReadCloser, error)
}
```

### Factory パターン

`FilesystemFactory` ([p../pkg/filesystems/factory.go](p../pkg/filesystems/factory.go)) が設定から適切なドライバーを生成。

### 設定ファイル形式

#### YAML 形式 (.kvtool.yml)

```yaml
version: 0.1
namespaces:
  default:
    vault:
      driver: "vault"
      args:
        endpoint: "http://localhost:8200"
        token: "root"
      mount:
        dir: "app"
        file: "prod"
```

#### HCL 形式 (.kvtool.hcl)

HCL 標準 + 環境変数展開（kvtool 拡張）をサポート ([pkg/config/hcl_loader.go](pkg/config/hcl_loader.go))

```hcl
# 変数定義（辞書形式）
locals {
  config = {
    environment = "development"
    port        = 8080
  }
}

# namespace 定義（YAML 形式と同じ構造）
namespaces {
  namespace "default" {
    vault {
      driver = "vault"
      args {
        endpoint = env.VAULT_ADDR           # 環境変数展開
        path     = "secret/${local.config.environment}/app"  # 変数参照
      }
      mount {
        dir  = "app"
        file = "prod"
      }
    }
  }

  namespace "production" {
    vault {
      driver = "vault"
      args {
        endpoint = "https://vault.prod.example.com"
      }
    }
  }
}
```

| 機能 | 構文 | 備考 |
|------|------|------|
| 変数定義 | `locals { ... }` | HCL 標準、辞書形式 |
| 変数参照 | `local.config.port` | HCL 標準、属性アクセス |
| 環境変数 | `env.VAR` | kvtool 拡張 |
| namespace 定義 | `namespaces { namespace "name" { ... } }` | YAML と同じ構造 |
| args ブロック | `args { ... }` | YAML の args と同等 |
| mount ブロック | `mount { ... }` | YAML の mount と同等 |

## コード規約

### パッケージコメント

各ドライバーのメインファイル先頭に記述:

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
//
// ## Security
//
// - パストラバーサル攻撃の防止
package filesystems
```

### 構造体タグ

設定構造体には `doc`, `required`, `default`, `example` タグを記述:

```go
type S3FsConfig struct {
    Bucket string `yaml:"bucket" doc:"S3 バケット名" required:"true" example:"my-bucket"`
    Region string `yaml:"region" doc:"AWS リージョン" required:"true" example:"us-east-1"`
    Root   string `yaml:"root" doc:"ルートパス" required:"false" default:"" example:"config"`
}
```

これらのタグから API リファレンスを自動生成:

```bash
make doc-gen  # docs/api-reference.md を生成
```

### セキュリティ規約

#### パストラバーサル防止

```go
func (fs *LocalFs) GetFile(path string) (File, error) {
    fullPath := filepath.Join(fs.Root, path)
    cleanPath := filepath.Clean(fullPath)
    if !strings.HasPrefix(cleanPath, fs.Root) {
        return nil, fmt.Errorf("path escapes root: %s", path)
    }
}
```

#### 相対パスのみ

```go
if filepath.IsAbs(path) || strings.HasPrefix(path, "~") {
    return nil, fmt.Errorf("path must be relative: %s", path)
}
```

### テスト規約

#### テーブル駆動テスト

```go
func TestSomething(t *testing.T) {
    tests := []struct {
        name        string
        input       string
        expected    string
        expectError bool
    }{
        {name: "正常系", input: "test", expected: "test", expectError: false},
        {name: "エラー系", input: "", expected: "", expectError: true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            require := require.New(t)
            // テスト実装
        })
    }
}
```

#### 統合テスト

[p../pkg/filesystems/integration_test.go](p../pkg/filesystems/integration_test.go) に全ドライバーの共通テストを追加。

## ドキュメント生成

コード内のコメントと構造体タグから自動生成:

```bash
make doc-gen  # docs/api-reference.md を生成
```

実装: [scripts/gen-docs.go](scripts/gen-docs.go)

### 生成されるファイル

- `docs/api-reference.md`: 全ドライバーの設定パラメータ一覧

### 生成内容

1. **ドライバー名**: パッケージコメントから
2. **設定パラメータ表**: 構造体タグから
3. **設定例**: `example` タグから

## 開発ワークフロー

### 新ドライバー追加

1. **仕様策定**: `do../pkg/filesystems/{driver}.md` を作成（手動）
2. **構造体定義**: 設定構造体を定義（タグ含む）
3. **インターフェース実装**: `Filesystem` と `File` を実装
4. **Factory 登録**: `filesystems/factory.go` に追加
5. **テスト**: `filesystems/integration_test.go` にテストケース追加
6. **ドキュメント生成**: `make doc-gen` を実行

### 新 Transform 追加

1. **パーサー実装**: `pkg/decoders/` に変換ロジック実装
2. **テスト**: テーブル駆動テスト追加
3. **LocalFs に統合**: `LocalFile.LoadAsJson()` で Transform 適用

## 暗黙知・設計判断

### ストア設定の3層分離（args / context / mount）

ストア設定は3つのブロックに分離する:

| ブロック | 役割 | 例 |
|---------|------|-----|
| `args` | ドライバー固有の接続情報 | bucket, url, addr, token |
| `context` | 共通の運用パラメータ | timeout（将来: retry 等） |
| `mount` | 共通の参照スコープ | dir, file |

```yaml
s3-store:
  driver: "s3"
  args:
    bucket: "my-bucket"       # 接続情報（ドライバー固有）
    region: "ap-northeast-1"
  context:
    timeout: 30               # 運用パラメータ（共通、秒単位）
  mount:
    dir: "config/production"  # バケット内のどこを見るか（共通）
```

`context.timeout` は `args.timeout` より優先される（後方互換のため `args.timeout` も動作する）。

#### 設計判断: args.root を廃止し mount.dir に統一

以前は S3 と REST で `args.root` を使っていたが、以下の理由で `mount.dir` に移行した:

| 項目 | args（ドライバー固有） | mount（共通） |
|------|----------------------|--------------|
| 目的 | 接続先を特定する | 参照範囲を絞る |
| 例 | bucket, endpoint, url | dir, file |
| ドライバー依存 | あり | なし |

- **local**: `args.root` はローカル物理パス → 接続先そのものなので args に残す
- **s3**: 旧 `args.root` はバケット内パス → 参照範囲なので `mount.dir` に移行
- **rest**: 旧 `args.root` は URL パス接頭辞 → 参照範囲なので `mount.dir` に移行
- **git**（将来）: キャッシュ先は kvtool 自動管理、リポジトリ内パスは `mount.dir`

#### S3 の設定

- **endpoint**: S3 エンドポイント（パス含む可、マルチテナント対応）
- **bucket**: バケット名（スラッシュ禁止）

検証:
- `bucket` にスラッシュが含まれる場合エラー
- `endpoint + bucket` が実際のバケットを示すか HeadBucket API で確認

### コマンド名の統一

- **`load`**: ファイルシステムの `cat` 的に「バイトストリームを流す」
- **`store get` → `store load`**: 粒度を揃えるため `load` に統一
- **`kvtool load` → `kvtool file load`**: 階層を明確化

### ドキュメント戦略

- **README.md**: 導入の概要のみ（コンパクト）
- **CLAUDE.md**: 開発ガイド、暗黙知、コード規約
- **docs/api-reference.md**: 自動生成
- **do../pkg/filesystems/**: ドライバーごとの詳細仕様

### Makefile のコメント

各ターゲットに概念説明を追加:

```makefile
# MinIO (S3 互換) 統合テスト用
# 開発用 MinIO を起動します（http://localhost:9000, console: http://localhost:9001）
.PHONY: minio-up
```

### テスト環境

- **Vault**: `make vault-up` で起動（http://localhost:8200, token: root）
- **MinIO**: `make minio-up` で起動（http://localhost:9000, user: minioadmin）
- テストスキップ: `SKIP_S3_TESTS=true` など環境変数で制御

## 参考資料

- [TODO.md](TODO.md): タスク管理・実装計画
- [docs/api-reference.md](docs/api-reference.md): API リファレンス（自動生成）
- [do../pkg/filesystems/s3.md](do../pkg/filesystems/s3.md): S3 仕様
- [p../pkg/filesystems/core.go](p../pkg/filesystems/core.go): 統一インターフェース
- [p../pkg/filesystems/integration_test.go](p../pkg/filesystems/integration_test.go): 統合テスト
- [pkg/config/hcl_loader.go](pkg/config/hcl_loader.go): HCL 構成ファイルローダー
- [test_data/configs/.kvtool.hcl.example](test_data/configs/.kvtool.hcl.example): HCL 設定例
