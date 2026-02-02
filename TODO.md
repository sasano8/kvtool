# TODO

## データソースから JSON への変換

### .env ファイル
- [x] .env ファイルを読んで JSON にする ([pkg/decoders/dotenv_to_json.go](pkg/decoders/dotenv_to_json.go))
- [x] 行頭（空白含む）コメントアウトの解釈 ([pkg/decoders/env_to_json.go:18](pkg/decoders/env_to_json.go#L18))
- [x] KEY=VALUE 形式のパース ([pkg/decoders/env_to_json.go:22-28](pkg/decoders/env_to_json.go#L22-L28))
- [x] クォート囲みの処理（ダブル・シングル） ([pkg/decoders/env_to_json.go:34-43](pkg/decoders/env_to_json.go#L34-L43))
- [x] エスケープシーケンスの処理 ([pkg/decoders/env_to_json.go:55-83](pkg/decoders/env_to_json.go#L55-L83))
- [x] テスト ([pkg/decoders/dotenv_to_json_test.go](pkg/decoders/dotenv_to_json_test.go))
- [ ] 行末コメント処理（未実装）
- [ ] クォート囲みがないスペースのエラー処理（`VALUE=hello world`）

### 環境変数
- [x] 環境変数を読んで JSON にする ([pkg/sources/env.go](pkg/sources/env.go), [pkg/decoders/env_to_json.go](pkg/decoders/env_to_json.go))
- [x] テスト ([pkg/decoders/env_to_json_test.go](pkg/decoders/env_to_json_test.go))

### Vault
- [x] Vault から構成を読んで JSON にする ([filesystems/vault.go:85-118](filesystems/vault.go#L85-L118))
- [x] Vault ファイルシステムのテスト ([filesystems/vault_test.go](filesystems/vault_test.go))
- [x] Vault からの読み込みの統合テスト ([Makefile:40-58](Makefile#L40-L58))

### HCL (HashiCorp Configuration Language)
- [ ] HCL ファイルを読んで JSON にする
  - [ ] HCL パーサーの実装（github.com/hashicorp/hcl/v2 使用）
  - [ ] HCL から JSON への変換
  - [ ] ネストされた構造のサポート
  - [ ] 変数展開のサポート
  - [ ] Transform 設定での利用
    ```yaml
    namespaces:
      default:
        terraform:
          driver: local
          args:
            root: "./terraform"
            transform:
              read: hcl
    ```
  - [ ] テストの実装
  - ユースケース:
    - Terraform 設定ファイルの読み込み
    - Nomad/Consul 設定との統合
    - HCL 形式の設定ファイルの JSON 化

## インターフェースの統一

- [x] 統一インターフェースの定義 ([filesystems/core.go:7-19](filesystems/core.go#L7-L19))
  - [x] Filesystem インターフェース (GetFile)
  - [x] File インターフェース (LoadAsJson, OpenReader)
- [x] VaultFs の統一インターフェース実装
  - [x] VaultFs.GetFile ([filesystems/vault.go:80](filesystems/vault.go#L80))
  - [x] VaultFsFile.LoadAsJson ([filesystems/vault.go:87](filesystems/vault.go#L87))
  - [x] VaultFsFile.OpenReader ([filesystems/vault.go:109](filesystems/vault.go#L109))
- [x] LocalFs の統一インターフェース実装
  - [x] LocalFs.GetFile ([filesystems/local.go:54](filesystems/local.go#L54))
  - [x] LocalFile.LoadAsJson ([filesystems/local.go:135](filesystems/local.go#L135))
  - [x] LocalFile.OpenReader ([filesystems/local.go:139](filesystems/local.go#L139))
- [x] FsEnvFile の統一インターフェース実装
  - [x] FsEnvFilesystem.GetFile ([filesystems/env.go](filesystems/env.go))
  - [x] FsEnvFile.LoadAsJson ([filesystems/env.go](filesystems/env.go))
  - [x] FsEnvFile.OpenReader ([filesystems/env.go](filesystems/env.go))
  - [x] Tests ([filesystems/env_test.go](filesystems/env_test.go))

## ファイルシステムの実装

### ローカルファイルシステム
- [x] ファイルシステムの実装 ([filesystems/local.go](filesystems/local.go))
- [x] ルートディレクトリより前に辿れないセキュリティ制限 ([filesystems/local.go:79-90](filesystems/local.go#L79-L90))
- [x] OpenReader の実装 ([filesystems/local.go:49-97](filesystems/local.go#L49-L97))
- [x] LoadAsJson の実装 ([filesystems/local.go:99-115](filesystems/local.go#L99-L115))
- [x] ローカルファイルシステムのテスト ([filesystems/local_test.go](filesystems/local_test.go))
- [x] セキュリティ制限のテスト（パストラバーサル攻撃の防止） ([filesystems/local_test.go:191-231](filesystems/local_test.go#L191-L231))

### Vault ファイルシステム
- [x] Vault ファイルシステムの実装 ([filesystems/vault.go](filesystems/vault.go))
- [x] VaultFs の実装 ([filesystems/vault.go:51-79](filesystems/vault.go#L51-L79))
- [x] VaultFsFile の実装 ([filesystems/vault.go:80-85](filesystems/vault.go#L80-L85))
- [x] Vault ファイルシステムのテスト ([filesystems/vault_test.go](filesystems/vault_test.go))
- [x] エラーハンドリングのテスト

### ファイル抽象化
- [x] ファイルはキー（パス）を固定 ([filesystems/vault.go:43-46](filesystems/vault.go#L43-L46), [filesystems/local.go:28-31](filesystems/local.go#L28-L31))
- [x] ファイルシステムをバックエンドの参照に持つ ([filesystems/vault.go:44](filesystems/vault.go#L44))
- [ ] ファイル操作の統合テスト

## ストアの実装

### ストア設定
- [x] ストア設定の構造体定義 ([filesystems/core.go](filesystems/core.go))
  - [x] StoreConfig ([filesystems/core.go:13-17](filesystems/core.go#L13-L17))
  - [x] MountConfig ([filesystems/core.go:3-6](filesystems/core.go#L3-L6))
  - [x] TransformConfig ([filesystems/core.go:8-11](filesystems/core.go#L8-L11))
- [x] 構成ファイルの例 ([.kvtool.yml](.kvtool.yml))
- [x] 構成ファイルのロード機能の実装 ([internal/config/config.go](internal/config/config.go))
- [x] パスパーサーの実装 ([internal/config/config.go:62-109](internal/config/config.go#L62-L109))
- [x] パスパーサーのテスト ([internal/config/config_test.go:45-98](internal/config/config_test.go#L45-L98))
- [ ] 構成ファイルのバリデーション

### ストアの機能
- [x] ストアで様々なデータソースを定義できるようにする
- [x] 構成ファイルを読んで、キーバリューストレージとして機能させる
- [x] 設定ファイルの自動解決 ([internal/config/config.go:45-95](internal/config/config.go#L45-L95))
  - [x] ローカル設定 (.kvtool.yml) の優先
  - [x] グローバル設定 (~/.config/kvtool/.kvtool.yml) へのフォールバック
- [x] store init コマンドの実装 ([cmd/cmd_store/init.go](cmd/cmd_store/init.go))
  - [x] ローカル設定ファイルの生成
  - [x] グローバル設定ファイルの生成 (--global)
  - [x] 既存ファイルの上書き保護 (--force で上書き可能)
  - [x] YAML 形式での出力
- [x] store init コマンドのテスト ([cmd/cmd_store/init_test.go](cmd/cmd_store/init_test.go))
- [x] store load コマンドの実装 ([cmd/cmd_store/get.go](cmd/cmd_store/get.go))
  - [x] パス形式: `store_name/file_path` ([internal/config/config.go:112-141](internal/config/config.go#L112-L141))
  - [x] namespace フラグによるテナント切り替え（デフォルト: default）
  - [x] ローカルファイルシステムからの読み込み ([cmd/cmd_store/get.go:76-107](cmd/cmd_store/get.go#L76-L107))
  - [x] Vault からの読み込み ([cmd/cmd_store/get.go:109-153](cmd/cmd_store/get.go#L109-L153))
  - [x] dotenv transform のサポート
  - [x] JSON 出力形式 ([cmd/cmd_store/get.go:165-175](cmd/cmd_store/get.go#L165-L175))
  - [x] raw 出力形式（key=value） ([cmd/cmd_store/get.go:177-194](cmd/cmd_store/get.go#L177-L194))
  - [x] 設定ファイル自動解決のサポート
- [x] store load コマンドのテスト ([cmd/cmd_store/get_test.go](cmd/cmd_store/get_test.go))
- [-] store ls コマンドの実装（不要）
- [-] store set コマンドの実装（不要）
- [-] store delete コマンドの実装（不要）
- [-] store load コマンドの実装（不要、store load と機能重複のため削除済み）

## その他
- [x] エンコーディングは UTF-8
- [x] マルチライン値は使わない方針

## 今後の改善計画

### 優先度：高（簡単かつ重要）

#### テストの充実
- [x] LocalFs の基本機能テスト ([filesystems/local_test.go](filesystems/local_test.go))
  - [x] GetFile, LoadAsJson, OpenReader の動作確認
  - [x] ファイルが存在しない場合のエラーハンドリング
  - [x] 様々なファイル形式（JSON, dotenv）の読み込み
  - [x] 空パス、絶対パス、~ パスのエラー処理
  - [x] サブディレクトリアクセス
  - [x] デフォルトルート（カレントディレクトリ）の動作

- [x] セキュリティテスト ([filesystems/local_test.go:191-231](filesystems/local_test.go#L191-L231))
  - [x] パストラバーサル攻撃の防止テスト（`../../../etc/passwd` など）
  - [x] Root ディレクトリ外へのアクセス拒否の確認
  - [x] Root ディレクトリ自体へのアクセス拒否

#### エラー型の統一
- [ ] カスタムエラー型の定義
  - 問題点: `fmt.Errorf` でエラーを生成しており、エラー種別の判定が困難
  - 提案: エラー型の定義
  ```go
  var (
      ErrFileNotFound = errors.New("file not found")
      ErrPermissionDenied = errors.New("permission denied")
      ErrInvalidPath = errors.New("invalid path")
  )
  ```
  - 理由: エラーハンドリングの統一と、呼び出し側でのエラー種別判定を可能にする

#### 設定ファイルのバリデーション
- [ ] StoreConfig の妥当性検証
  - 必須フィールドのチェック（driver, args など）
  - driver タイプの検証（local, vault, env のみ許可）
  - 矛盾する設定の検出（mount.dir と mount.file が両方指定されている、など）

#### コンテキスト伝播の統一
- [x] すべての Filesystem で context.Context を保持するように統一
  - [x] FsEnvFilesystem に Ctx フィールドを追加 ([filesystems/env.go:16](filesystems/env.go#L16))
  - [x] FsEnvFile が親の Filesystem を参照 ([filesystems/env.go:28-30](filesystems/env.go#L28-L30))
  - [x] Factory で FsEnvFilesystem に context を渡す ([filesystems/factory.go:98-103](filesystems/factory.go#L98-L103))
  - 設計方針: Filesystem レベルで context を保持し、GetFile は context を受け取らない
  - 理由: LocalFs, VaultFs が既に context を保持する設計であり、統一性を保つため



#### 環境変数ファイルシステムの設定対応
- [x] FsEnvFilesystem を .kvtool.yml で設定可能にする
  - [x] driver: "env" のサポート ([cmd/cmd_store/get.go:79](cmd/cmd_store/get.go#L79), [cmd/cmd_store/get.go:171-179](cmd/cmd_store/get.go#L171-L179))
  - [x] getFromEnvFs 関数の実装 ([cmd/cmd_store/get.go:171-179](cmd/cmd_store/get.go#L171-L179))
  - [x] テスト ([cmd/cmd_store/get_test.go:53-77](cmd/cmd_store/get_test.go#L53-L77))
  - [x] サンプル設定ファイル ([test_data/configs/.kvtool.env.example.yml](test_data/configs/.kvtool.env.example.yml))

### 優先度：中（重要だが複雑）

#### パッケージ構造の改善

- [x] `internal/shared/` から `internal/config/` へのリネーム
  - [x] パッケージディレクトリの移動とパッケージ宣言の更新
  - [x] すべてのインポート文を更新 ([cmd/cmd_store/get.go](cmd/cmd_store/get.go), [cmd/cmd_store/init.go](cmd/cmd_store/init.go))
  - [x] 変数シャドーイングの修正 (config パッケージと config 変数の衝突を解消)
  - [x] コメントとドキュメントの更新 ([filesystems/factory.go](filesystems/factory.go), [TODO.md](TODO.md), [docs/design.md](docs/design.md))
  - 実装方針: `internal/config` を選択（設定ファイル管理が主な責務のため）
  - 理由: パッケージ名で責務（.kvtool.yml の管理）を明確化、Go の標準的な命名規則に準拠

- [x] `filesystems/core.go` から未使用の設定構造体を削除
  - [x] 未使用の構造体を削除: `StoreConfig`, `MountConfig`, `TransformConfig` ([filesystems/core.go](filesystems/core.go))
  - [x] コアインターフェースのみを保持: `Filesystem`, `File`
  - 実装方針: 削除を選択（使われていないコードの削除）
  - 理由:
    - これらの構造体は実際には使用されていない
    - 類似の構造体が `internal/config` に既に存在し、活用されている (`StoreInfo`, `MountInfo`)
    - ファイルシステムパッケージは抽象化（インターフェース）に集中すべき

#### ファイルシステムファクトリーの導入
- [x] Filesystem の生成を統一するファクトリーパターンの実装
  - [x] FilesystemFactory 構造体の実装 ([filesystems/factory.go](filesystems/factory.go))
  - [x] Create メソッドによる各ドライバーの統一生成
    - [x] createLocalFs ([filesystems/factory.go:36-50](filesystems/factory.go#L36-L50))
    - [x] createVaultFs ([filesystems/factory.go:52-87](filesystems/factory.go#L52-L87))
    - [x] createEnvFs ([filesystems/factory.go:89-93](filesystems/factory.go#L89-L93))
  - [x] コマンド層のリファクタリング ([cmd/cmd_store/get.go:71-90](cmd/cmd_store/get.go#L71-L90))
  - [x] getFileContent 関数による統一的なファイル取得 ([cmd/cmd_store/get.go:92-123](cmd/cmd_store/get.go#L92-L123))
  - [x] テスト ([filesystems/factory_test.go](filesystems/factory_test.go))

#### 統合インターフェーステスト
- [x] 全ファイルシステムで共通の動作を確認するテスト ([filesystems/integration_test.go](filesystems/integration_test.go))
  - [x] GetFile インターフェースの一貫性テスト
  - [x] LoadAsJson インターフェースの一貫性テスト
  - [x] OpenReader インターフェースの一貫性テスト
  - [x] LoadAsJson と OpenReader の一貫性テスト
  - [x] エラーハンドリングの一貫性テスト
  - 実装方針: テーブル駆動テストで LocalFs と FsEnvFilesystem を同じテストケースで検証
  - 配置: filesystems/ パッケージ内（ファイルシステム実装と同じ場所）
  - テストカバレッジ:
    - ✅ 有効なパスでのファイル取得
    - ✅ ネストされたパスの処理
    - ✅ 単純なJSONの読み込み
    - ✅ ネストされたJSONの読み込み
    - ✅ dotenv transform の動作
    - ✅ ファイル不在時のエラー処理
    - ✅ 環境変数の読み込み
    - ✅ 空パス、絶対パス、チルダパスの拒否
    - ✅ パストラバーサル攻撃の防止
  

#### Transform の統一インターフェース化
- [x] Transform を LocalFs に統合（ファイルシステム層への移動）
  - [x] LocalFsConfig に Transform フィールドを追加 ([filesystems/local.go:23-26](filesystems/local.go#L23-L26))
  - [x] LocalFs に Transform フィールドを追加 ([filesystems/local.go:33-38](filesystems/local.go#L33-L38))
  - [x] LocalFile.LoadAsJson() で Transform を自動適用 ([filesystems/local.go:134-157](filesystems/local.go#L134-L157))
  - [x] Factory で transform 設定を抽出 ([filesystems/factory.go:40-60](filesystems/factory.go#L40-L60))
  - [x] コマンド層から Transform ロジックを削除 ([cmd/cmd_store/get.go:98-116](cmd/cmd_store/get.go#L98-L116))
  - [x] Transform 機能のテスト ([filesystems/transform_test.go](filesystems/transform_test.go))
  - 実装方針: LocalFs に設定を持たせ、LoadAsJson() で自動適用
  - 理由:
    - インターフェース変更なし（後方互換性を維持）
    - ビジネスロジックをコマンド層からファイルシステム層に移動
    - 各ファイルシステムが独自の Transform 実装を持てる
    - 再利用可能（どこから File を取得しても Transform が適用される）
  - サポートされる Transform: "dotenv", "json" (デフォルト)

#### Store サービス層の導入
- [x] コマンド層とファイルシステム層の間にサービス層を挿入
  - [x] StoreService インターフェースの定義 ([internal/service/store.go:12-15](internal/service/store.go#L12-L15))
  - [x] Get メソッドの実装 ([internal/service/store.go:48-83](internal/service/store.go#L48-L83))
  - [x] ConfigLoader インターフェースの定義 ([internal/service/store.go:29-31](internal/service/store.go#L29-L31))
  - [x] FilesystemFactory インターフェースの定義 ([internal/service/store.go:34-36](internal/service/store.go#L34-L36))
  - [x] コマンド層のリファクタリング ([cmd/cmd_store/get.go:35-65](cmd/cmd_store/get.go#L35-L65))
  - [x] サービス層のテスト ([internal/service/store_test.go](internal/service/store_test.go))
    - [x] 基本的な Get 操作のテスト
    - [x] Transform 機能のテスト
    - [x] エラーハンドリングのテスト (無効なパス、設定ファイル不在、ストア不在)
    - [x] モックを使った単体テスト
  - 実装方針:
    - ビジネスロジックをコマンド層からサービス層に移動
    - 依存性注入を使った疎結合な設計
    - インターフェースを使ったテスト可能な実装
  - 効果:
    - ✅ コマンド層のコード削減（95行 → 65行、約32%削減）
    - ✅ CLI 以外（API、ライブラリ）から再利用可能
    - ✅ ビジネスロジックの独立したテスト可能
    - ✅ 責務の明確化（Command = UI、Service = ビジネスロジック）

#### 設定管理とパス解決の分離
- [x] `internal/config/config.go` の責務を分割
  - [x] types.go への分割 - データ構造定義 ([internal/config/types.go](internal/config/types.go))
  - [x] loader.go への分割 - 設定ファイルのロード ([internal/config/loader.go](internal/config/loader.go))
  - [x] registry.go への分割 - ストアの管理・取得 ([internal/config/registry.go](internal/config/registry.go))
  - [x] path.go への分割 - パス文字列の解析 ([internal/config/path.go](internal/config/path.go))
  - [x] 既存テストの動作確認 ([internal/config/config_test.go](internal/config/config_test.go))
  - 実装方針: 単一責任の原則に従い、4つのファイルに分割（types, loader, registry, path）
  - 効果:
    - ✅ config.go を 166行 → 4ファイル 169行に分割（types: 20行, loader: 68行, registry: 17行, path: 64行）
    - ✅ 各ファイルが明確な責務を持つ
    - ✅ テスト容易性の向上
    - ✅ コード可読性の向上

#### デコーダとソースの統合検討
- [x] `pkg/decoders/` と `pkg/sources/` の関係を整理
  - [x] Source インターフェースの定義 ([pkg/sources/source.go](pkg/sources/source.go))
  - [x] Decoder インターフェースの定義 ([pkg/decoders/decoder.go](pkg/decoders/decoder.go))
  - [x] SourceEnv が Source インターフェースを実装 ([pkg/sources/env.go](pkg/sources/env.go))
  - [x] EnvDecoder が Decoder インターフェースを実装 ([pkg/decoders/decoder.go:25-35](pkg/decoders/decoder.go#L25-L35))
  - [x] フォーマット契約を明示的にドキュメント化（KEY=VALUE 形式）
  - [x] テストの追加 ([pkg/sources/source_test.go](pkg/sources/source_test.go), [pkg/decoders/decoder_test.go](pkg/decoders/decoder_test.go))
  - 実装方針: 疎結合を選択（統合ではなく、明示的なインターフェースで分離）
  - 理由: 拡張性、テスト容易性、Go の標準的な設計パターンに準拠

#### 新しいファイルシステムドライバーの実装

##### HTTP ファイルシステム
- [ ] HTTP ファイルシステムの実装
  - [ ] HttpFs 構造体の実装
  - [ ] HTTP リクエストによるファイル取得
  - [ ] レスポンスボディを JSON として返す
  - [ ] 設定ファイルでの定義
    ```yaml
    namespaces:
      default:
        api:
          driver: http
          args:
            base_url: "https://api.example.com"
            headers:
              Authorization: "Bearer token"
            timeout: 30s
    ```
  - [ ] エラーハンドリング（タイムアウト、接続エラー、HTTP エラーステータス）
  - [ ] テストの実装
  - ユースケース: REST API からの設定取得、外部サービスとの連携

##### データベースファイルシステム
- [ ] データベースファイルシステムの実装
  - [ ] DbFs 構造体の実装
  - [ ] ユーザー定義 SQL クエリのサポート
    - クエリテンプレートで `{key}` と `{namespace}` をプレースホルダーとして使用
    - 例: `SELECT value FROM config WHERE key = {key} AND namespace = {namespace}`
  - [ ] VARCHAR（JSON 想定）を返す
  - [ ] 設定ファイルでの定義
    ```yaml
    namespaces:
      default:
        config:
          driver: db
          args:
            connection_string: "postgres://user:pass@localhost/db"
            query: "SELECT value FROM env WHERE key = {key} AND namespace = {namespace}"
            # または、より柔軟な設定:
            # query: "SELECT data FROM my_table WHERE id = {key}"
            # query: "SELECT json_col FROM config_v2 WHERE tenant = {namespace} AND config_key = {key}"
            timeout: 10s
    ```
  - [ ] プレースホルダーの置換処理
    - SQL インジェクション対策（プリペアドステートメント使用）
    - {key} と {namespace} の適切なエスケープ
  - [ ] サポートするデータベース
    - [ ] PostgreSQL
    - [ ] MySQL
    - [ ] SQLite
  - [ ] コネクションプーリング
  - [ ] エラーハンドリング（接続エラー、クエリエラー、データ不在）
  - [ ] テストの実装（モックまたは testcontainers 使用）
  - ユースケース:
    - マルチテナント設定管理
    - 既存データベーススキーマとの統合
    - レガシーシステムとの互換性維持

##### S3 ファイルシステム
- [ ] S3 ファイルシステムの実装


### 優先度：低（将来的に必要だが現在は後回し）

#### File インターフェースの拡張
- [ ] 書き込み操作のインターフェース定義
  - 問題点: 読み込み専用インターフェースで、書き込み操作が未定義
  - 提案: `WritableFile` インターフェースまたは `SaveAsJson(data any) error` メソッドの追加
  - 理由: store set コマンドの実装を見据えた設計

- [ ] リスト操作のインターフェース定義
  - 問題点: ディレクトリ内のファイル一覧取得が未定義
  - 提案: `Filesystem.List(ctx context.Context, path string) ([]FileInfo, error)` の追加
  - 理由: store ls コマンドの実装を見据えた設計

#### E2E テスト
- [ ] コマンドラインの E2E テスト
  - `store init`, `store load` の実行結果を検証
  - 設定ファイルの自動解決動作の確認
  - namespace フラグの動作確認


#### ドキュメント整備

- [ ] Filesystem / File インターフェースの仕様ドキュメント
  - 各メソッドの振る舞い、エラー条件、利用例
  - 実装者向けガイドライン

- [ ] パッケージ構成図とレイヤー図
  - コマンド層 → サービス層 → ファイルシステム層 の関係
  - 依存関係の方向性

- [ ] .kvtool.yml のスキーマドキュメント
  - 各フィールドの説明と例
  - driver ごとの args 仕様

- [ ] **mdBook による公式ドキュメントサイトの構築**
  - [ ] mdBook のインストールと初期化
  - [ ] ドキュメント構造の整理
    - [ ] Getting Started ガイド
    - [ ] ファイルシステムドライバー仕様（Local, Vault, S3, HTTP, DB）
    - [ ] Transform 仕様（dotenv, JSON, HCL）
    - [ ] API リファレンス
    - [ ] チュートリアル集
  - [ ] 既存マークダウンの移行と整理
  - [ ] CI/CD での自動ビルド & デプロイ
    - GitHub Actions で mdBook をビルド
    - GitHub Pages または Vercel へ自動デプロイ
  - [ ] カスタムテーマの適用（オプション）
  - [ ] 検索機能の有効化

  **おすすめ理由:**
  1. **設定ファイルが最小限** (book.toml のみ、5行程度)
  2. **Rust 製で超高速** - ビルドが一瞬で完了
  3. **ゼロ依存** - バイナリ1つで完結、Node.js 不要
  4. **検索機能組み込み** - デフォルトで全文検索が利用可能
  5. **オフライン対応** - ビルド済み HTML は完全に静的
  6. **PDF 出力対応** - ドキュメントを PDF として配布可能
  7. **Go ユーザーに親和性高い** - Rust/Go 両方ビルドツールという共通点
  8. **Markdown のみ** - 既存ドキュメントをそのまま利用可能
  9. **GitHub Actions 対応** - mdbook-action で簡単にデプロイ
  10. **実績豊富** - Rust Book など多くの公式ドキュメントで採用

  **実装例:**
  ```bash
  # インストール（1回のみ）
  cargo install mdbook

  # 初期化
  mdbook init docs

  # ローカルプレビュー（ホットリロード付き）
  mdbook serve

  # ビルド
  mdbook build
  ```

  **最小限の設定ファイル (book.toml):**
  ```toml
  [book]
  title = "kvtool Documentation"
  authors = ["kvtool contributors"]
  language = "ja"

  [output.html]
  git-repository-url = "https://github.com/your-username/kvtool"
  ```

#### パフォーマンス最適化（現時点では不要）

- [ ] ファイル内容のキャッシュ（オプション）
  - 同一ファイルへの複数回アクセス時のパフォーマンス改善
  - TTL ベースの無効化

- [ ] 複数ファイルの並行読み込み
  - context.Context を活用した並行処理
  - エラーハンドリングの統一（errgroup の利用など）

#### ファイルレベルのコンテキスト管理

**現在の課題:**
- GetFile が context を受け取れないため、ファイル操作ごとに異なるタイムアウトやキャンセル処理を設定できない
- Filesystem のライフサイクル全体で1つの context を共有するため、柔軟性が制限される

**将来的な改善案:**
1. **オプション1: GetFile に context を追加**
   - `GetFile(ctx context.Context, path string) (File, error)` に変更
   - メリット: 標準的な Go の I/O パターン、操作ごとに異なる context を使用可能
   - デメリット: 既存コード全体の変更が必要、インターフェースが複雑化

2. **オプション2: WithContext メソッドの追加**
   - `Filesystem.WithContext(ctx context.Context) Filesystem` を追加
   - 新しい context で Filesystem のコピーを作成
   - メリット: 既存インターフェースを変更せず、必要に応じて context を変更可能
   - デメリット: Filesystem のコピー生成コスト

3. **オプション3: FileOptions の導入**
   - `GetFile(path string, opts ...FileOption) (File, error)` のようなオプショナルパラメータ
   - `WithContext(ctx)`, `WithTimeout(duration)` などのオプション関数
   - メリット: 後方互換性を保ちつつ拡張可能
   - デメリット: やや複雑な API

**推奨:** 現時点では現在の設計（Filesystem レベルで context 保持）で十分。将来的に必要になった場合はオプション2（WithContext メソッド）が最も影響が少ない

---

## プロダクト総評と市場需要予測

### 概要

**kvtool** は、様々な設定ファイル形式（.env, JSON, YAML, HCL, Vault, データベースなど）を統一的なインターフェースで扱い、環境構築の初期段階で設定を統合・管理するための CLI ツールです。

### コアコンセプト

開発者が直面する「設定の複雑性」を解決します：
- 環境変数、.env ファイル、Vault、データベース、HTTP API など、設定が様々な場所に分散
- 各データソースごとに異なるツールやライブラリが必要
- 設定の統合と変換が煩雑
- ローカル開発環境とプロダクション環境での設定管理の乖離

**kvtool** は、これら全てを統一的な方法で扱える「設定のためのファイルシステム抽象化」を提供します。

### 技術的な強み

#### 1. クリーンアーキテクチャ
- **レイヤー分離**: Presentation（CLI） → Service → Infrastructure（Filesystem）
- **依存性注入**: テスト可能で保守性の高い設計
- **単一責任の原則**: 各パッケージが明確な責務を持つ

#### 2. 拡張性
- **プラグイン可能なドライバー**: Local, Vault, Env, HTTP, DB など、新しいソースを簡単に追加可能
- **Transform 機能**: dotenv, JSON, HCL など、任意の形式変換をサポート
- **統一インターフェース**: Filesystem と File の2つのインターフェースで全てを抽象化

#### 3. セキュリティ
- パストラバーサル攻撃の防止
- SQL インジェクション対策（プリペアドステートメント）
- Vault との統合によるシークレット管理

#### 4. Go の標準的なパターン
- context.Context による適切なライフサイクル管理
- io.Reader/io.ReadCloser による標準的な I/O
- エラーハンドリングのベストプラクティス

### 主要ユースケース

#### 1. 環境構築の初期化（Primary Use Case）
```bash
# 開発環境のセットアップスクリプト
kvtool store load config/app | jq -r '.DATABASE_URL' > .env
kvtool store load secrets/api-keys --namespace=production | apply-to-k8s
```
- Docker Compose の起動前に環境変数を準備
- Kubernetes のシークレットを一元管理
- CI/CD パイプラインでの設定注入

#### 2. マルチテナント設定管理
```yaml
# データベースから各テナントの設定を取得
namespaces:
  tenant-a:
    config:
      driver: db
      args:
        query: "SELECT value FROM config WHERE tenant = 'tenant-a' AND key = {key}"
  tenant-b:
    config:
      driver: db
      args:
        query: "SELECT value FROM config WHERE tenant = 'tenant-b' AND key = {key}"
```

#### 3. ハイブリッド環境での設定統合
```yaml
# ローカルは .env、本番は Vault
namespaces:
  development:
    secrets:
      driver: local
      args:
        root: "./"
  production:
    secrets:
      driver: vault
      args:
        address: "https://vault.company.com"
```

#### 4. レガシーシステムとの統合
- 既存データベースの設定テーブルから読み込み
- HTTP API 経由で設定サービスと連携
- .env ファイルと Vault の段階的な移行

### 市場需要予測

#### 🟢 高い需要が見込まれる領域

##### 1. DevOps / SRE チーム（需要: 高）
- **課題**: 複数環境（dev/staging/prod）での設定管理の煩雑さ
- **価値**: 統一的なインターフェースで環境間の差異を吸収
- **市場規模**: 中〜大企業の開発チームで広く採用される可能性

##### 2. マイクロサービス環境（需要: 高）
- **課題**: 数十〜数百のサービスそれぞれが異なる設定ソースを持つ
- **価値**: 設定の中央集権化と標準化
- **市場規模**: コンテナオーケストレーション（Kubernetes）利用企業

##### 3. スタートアップ / 中小企業（需要: 中〜高）
- **課題**: 初期段階では .env、成長に伴い Vault や設定サービスへ移行
- **価値**: 段階的な移行パスを提供、技術的負債の軽減
- **市場規模**: 急成長中の tech スタートアップ

##### 4. コンサルティング / システムインテグレーター（需要: 中）
- **課題**: 顧客ごとに異なる設定管理システム
- **価値**: 標準化されたツールで再利用性向上
- **市場規模**: クラウド移行案件での採用

#### 🟡 中程度の需要が見込まれる領域

##### 5. CI/CD パイプライン（需要: 中）
- **課題**: 各ステージで異なる設定ソースからの値取得
- **価値**: パイプラインスクリプトの簡素化
- **市場規模**: GitHub Actions / GitLab CI ユーザー

##### 6. データエンジニアリング（需要: 中）
- **課題**: データパイプラインの設定が複雑化
- **価値**: Airflow/dbt などとの統合
- **市場規模**: データ基盤を持つ企業

#### 🔴 競合との比較

**類似ツール:**
- **Vault CLI**: シークレット特化、他のソースとの統合は弱い
- **direnv**: ローカル環境特化、リモートソース非対応
- **dotenv 系ツール**: .env ファイルのみ、拡張性に乏しい
- **config management tools (Ansible/Chef)**: 重厚、軽量な設定取得には不向き

**kvtool の差別化ポイント:**
1. **軽量**: 単一バイナリ、依存なし
2. **柔軟**: プラグイン可能なドライバー
3. **統一**: 全ての設定ソースを同じ方法で扱う
4. **シンプル**: 学習コストが低い

### 成長戦略

#### フェーズ 1: MVP と初期ユーザー獲得（現在）
- ✅ コア機能の実装（Local, Env, Vault）
- 🔄 統合テストの充実
- 📝 ドキュメント整備
- 🎯 ターゲット: OSS コントリビューター、DevOps コミュニティ

#### フェーズ 2: エコシステム拡大（3-6ヶ月）
- HTTP / DB ファイルシステムの実装
- HCL / YAML / TOML トランスフォーマー
- プラグインシステムの導入
- 🎯 ターゲット: Terraform / Kubernetes ユーザー

#### フェーズ 3: エンタープライズ対応（6-12ヶ月）
- 監査ログ機能
- RBAC（ロールベースアクセス制御）
- GUI / TUI の提供
- クラウドサービスとの公式統合（AWS SSM, GCP Secret Manager, Azure Key Vault）
- 🎯 ターゲット: エンタープライズ企業

### 収益化の可能性

#### オープンソースモデル
- **MIT / Apache 2.0 ライセンス**: 広く採用されやすい
- **スポンサーシップ**: GitHub Sponsors 経由
- **コミュニティ駆動**: コントリビューション文化の醸成

#### 商用サービス展開
1. **kvtool Cloud**: マネージドサービス版
   - 設定のホスティング
   - チーム管理機能
   - SaaS モデル（$10-50/月）

2. **エンタープライズサポート**:
   - 導入支援
   - カスタマイズ開発
   - SLA 付きサポート契約

3. **トレーニング / コンサルティング**:
   - 企業向けワークショップ
   - ベストプラクティス提供

### 潜在的なリスク

#### 技術的リスク
- **Go エコシステムの変化**: 標準ライブラリの変更に追従
- **依存ライブラリのメンテナンス**: Vault SDK などの更新対応

#### 市場リスク
- **大手プレイヤーの参入**: HashiCorp が同様の機能を Vault に統合する可能性
- **クラウドベンダーのロックイン**: AWS/GCP/Azure が自社ツールを強化

#### ユーザー獲得リスク
- **既存ツールへの慣れ**: 移行コストが障壁に
- **ニッチ過ぎる可能性**: 市場が想定より小さい

### リスク軽減策

1. **コミュニティファースト**: 早期からユーザーフィードバックを取り入れる
2. **相互運用性重視**: 既存ツールとの併用を前提とした設計
3. **マイグレーションガイド**: 既存ツールからの移行を容易に
4. **クラウド中立**: 特定ベンダーに依存しない

### 総合評価

#### 強み
- ✅ **明確な課題解決**: 設定管理の複雑性を実際に解決
- ✅ **技術的な洗練**: Go のベストプラクティスに従った設計
- ✅ **拡張性**: 新しいドライバーやトランスフォーマーの追加が容易
- ✅ **軽量**: シングルバイナリで導入が簡単

#### 弱み・改善点
- ⚠️ **認知度**: まだ知られていない新しいツール
- ⚠️ **エコシステム**: プラグインやコミュニティが未成熟
- ⚠️ **ドキュメント**: 実例やチュートリアルの充実が必要
- ⚠️ **GUI**: CLI のみでは非エンジニアに敷居が高い

### 市場需要スコア（10点満点）

| 項目 | スコア | 説明 |
|------|--------|------|
| 課題の深刻度 | 8/10 | 設定管理は実際に多くの開発チームが抱える問題 |
| ソリューションの独自性 | 7/10 | 類似ツールはあるが、統一的なアプローチは新しい |
| 技術的実現性 | 9/10 | 既に MVP レベルの実装が完了 |
| 市場規模 | 7/10 | DevOps/SRE 市場は成長中だが、ニッチな可能性も |
| 収益化可能性 | 6/10 | OSS として広まれば商用サービス展開も可能 |
| 競合優位性 | 7/10 | 軽量・柔軟・統一という差別化要素あり |
| 成長ポテンシャル | 8/10 | マイクロサービス/Kubernetes の普及が追い風 |

**総合スコア: 7.4/10** 🟢

### 結論

**kvtool** は、環境構築と設定管理の領域において**実用的で需要のあるツール**です。

#### 成功の鍵
1. **早期のユーザー獲得**: DevOps コミュニティでの認知度向上
2. **実践的なドキュメント**: 実例とベストプラクティスの提供
3. **エコシステムの充実**: プラグイン、統合、ツールチェーンの拡大
4. **継続的な改善**: ユーザーフィードバックに基づく機能追加

#### 推奨される次のステップ
1. 📝 **ドキュメント整備**: README、チュートリアル、ユースケース集
2. 🎥 **デモ動画作成**: 実際の使用例を視覚的に示す
3. 🌟 **コミュニティ構築**: GitHub Discussions、Slack/Discord チャンネル
4. 📢 **プロモーション**: Hacker News、Reddit (r/devops)、Dev.to などでの紹介
5. 🔧 **パートナーシップ**: Terraform、Vault、Kubernetes コミュニティとの連携

**市場タイミング**: ✅ **Good**
マイクロサービス・Kubernetes・クラウドネイティブの普及により、設定管理の複雑性は増大しています。今がツールを投入する好機です。

**長期的な展望**: 🌟 **Promising**
適切に成長戦略を実行すれば、DevOps ツールチェーンの標準的な一部となる可能性があります。特に「環境構築の最初のステップ」という明確なポジショニングは、採用の障壁を下げる重要な要素です。