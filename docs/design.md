# 設計思想・仕様

## 設計思想

### 1. ストア機能を中心とした統一インターフェース

kvtool は複数の設定ソース（.env、Vault、JSON など）を **ストア** として統一的に扱います。

**利点:**
- 設定ソースが増えてもインターフェースは変わらない
- チーム全体で統一された設定管理方法
- 環境ごとの設定切り替えが容易

### 2. ローカル優先の設定管理

設定ファイルの解決は **ローカル優先、グローバルは明示的** とします。

**理由:**
- プロジェクトごとの設定を優先
- グローバル設定への意図しない依存を防ぐ
- チーム間での設定の一貫性を保つ

**動作:**
```bash
# ローカル .kvtool.yml がない場合、エラー
$ kvtool store get .env/APP_NAME
Error: config file not found: .kvtool.yml (run 'kvtool store init' to create one)

# グローバルを使う場合は明示的に指定
$ kvtool store get .env/APP_NAME --global
```

### 3. マルチテナント対応

namespace により複数の環境（開発、ステージング、本番など）を切り替えます。

**利点:**
- 1つの設定ファイルで複数環境を管理
- 環境の切り替えがフラグ1つで可能
- 環境ごとの設定の見通しが良い

**使用例:**
```bash
# 開発環境
kvtool store get .env/config -n development

# 本番環境
kvtool store get .env/config -n production
```

### 4. セキュリティ第一

**パストラバーサル防止:**
- `root` で指定したディレクトリより上位には遡れない
- 相対パスのみ受け付け、絶対パスは拒否
- `..` による上位ディレクトリへのアクセスを防止

**実装:**
```go
// filesystems/local.go
func (fs *LocalFs) OpenReader(path string) (io.ReadCloser, error) {
    // 絶対パスを拒否
    if filepath.IsAbs(p) {
        return nil, fmt.Errorf("path must be relative: %q", p)
    }

    // root より上位への遡りを防止
    if strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." {
        return nil, fmt.Errorf("path escapes root: %q", p)
    }
    // ...
}
```

## アーキテクチャ

### レイヤー構造

```
┌─────────────────────────────────────┐
│  CLI Layer (cmd/)                   │
│  - store init, store get            │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│  Config Layer (internal/config)     │
│  - 設定ファイルの解決               │
│  - パスパーサー                     │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│  Filesystem Layer (filesystems/)    │
│  - ドライバー抽象化                 │
│  - local, vault                     │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│  Decoder/Encoder Layer (pkg/)       │
│  - dotenv, json, yaml               │
└─────────────────────────────────────┘
```

### コンポーネント

#### 1. CLI Layer (`cmd/`)

ユーザーインターフェース。Cobra を使用。

- `cmd/cmd_store/`: ストア関連コマンド
  - `init.go`: 設定ファイル初期化
  - `get.go`: ファイル取得

#### 2. Config Layer (`internal/config/`)

設定ファイルの管理。

- `config.go`: 設定ファイルの読み込み・解決
- `ParseStorePath()`: パスパーサー
- `ResolveConfigPath()`: 設定ファイル解決

#### 3. Filesystem Layer (`filesystems/`)

ストレージバックエンドの抽象化。

- `local.go`: ローカルファイルシステム
- `vault.go`: HashiCorp Vault
- `core.go`: 共通型定義

#### 4. Decoder/Encoder Layer (`pkg/`)

データフォーマットの変換。

- `decoders/`: 各種フォーマットを JSON に変換
- `encoders/`: JSON から各種フォーマットに変換

## データフロー

### store get コマンドの処理フロー

```
1. コマンドライン引数をパース
   ↓
2. 設定ファイルを解決
   - ローカル .kvtool.yml を探す
   - --global なら ~/.config/kvtool/.kvtool.yml
   ↓
3. パスをパース (store_name/file_path)
   ↓
4. namespace と store_name からストア情報を取得
   ↓
5. ドライバーに応じてファイルを読み込み
   - local: ローカルファイルシステムから読み込み
   - vault: Vault API 経由で読み込み
   ↓
6. transform 設定に応じて変換
   - dotenv: .env 形式を JSON に変換
   ↓
7. 出力形式に応じて整形
   - json: JSON 形式で出力
   - raw: key=value 形式で出力
```

## パス仕様の詳細

### パス形式

```
store_name/file_path
```

**設計判断:**
- namespace はパスの一部ではなく、フラグ（`--namespace`）で指定
- テナント（環境）の切り替えが柔軟になる
- REST API との相性が良い（コロン `:` を含まない）

### パースロジック

```go
func ParseStorePath(path string) (*StorePath, error) {
    parts := splitPath(path)  // "/" で分割、空要素を除外

    return &StorePath{
        StoreName: parts[0],         // 最初の要素
        FilePath:  joinPath(parts[1:])  // 残りを "/" で結合
    }, nil
}
```

**例:**
- `".env/APP_NAME"` → `{StoreName: ".env", FilePath: "APP_NAME"}`
- `"vault/app/prod/db"` → `{StoreName: "vault", FilePath: "app/prod/db"}`
- `".env"` → `{StoreName: ".env", FilePath: ""}`

## 拡張性

### 新しいドライバーの追加

新しいストレージバックエンドを追加する場合：

1. `filesystems/` に新しいドライバーを実装
2. `LoadAsJson()` メソッドを実装
3. `cmd/cmd_store/get.go` の switch 文に追加

例（S3 ドライバー）:

```go
// filesystems/s3.go
type S3Fs struct {
    Client *s3.Client
    Bucket string
    Root   string
}

func (fs *S3Fs) LoadAsJson(path string) (interface{}, error) {
    // S3 から読み込み
}

// cmd/cmd_store/get.go
switch storeInfo.Driver {
case "local":
    content, err = getFromLocalFs(storeInfo, parsed.FilePath)
case "vault":
    content, err = getFromVaultFs(storeInfo, parsed.FilePath)
case "s3":  // 追加
    content, err = getFromS3Fs(storeInfo, parsed.FilePath)
}
```

### 新しい transform の追加

新しいフォーマット変換を追加する場合：

1. `pkg/decoders/` に新しいデコーダーを実装
2. `cmd/cmd_store/get.go` の `getFromLocalFs()` に追加

例（YAML transform）:

```go
// pkg/decoders/yaml_to_json.go
func YamlToJson(r io.Reader) (map[string]any, error) {
    // YAML を JSON に変換
}

// cmd/cmd_store/get.go
transform := getTransform(args, "read")
switch transform {
case "dotenv":
    return decoders.DotenvToJson(reader)
case "yaml":  // 追加
    return decoders.YamlToJson(reader)
}
```

## 今後の拡張

### 予定している機能

- [ ] 書き込み機能（`kvtool store set`）
- [ ] ストア間のコピー（`kvtool store cp`）
- [ ] 暗号化サポート
- [ ] 他のストレージバックエンド（S3、GCS など）

### 検討中の機能

- [ ] ストア一覧表示（`kvtool store ls`）
- [ ] 差分表示（`kvtool store diff`）
- [ ] 履歴管理（バージョニング）
- [ ] 監視・同期機能
