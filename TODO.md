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
- [x] store get コマンドの実装 ([cmd/cmd_store/get.go](cmd/cmd_store/get.go))
  - [x] パス形式: `store_name/file_path` ([internal/config/config.go:112-141](internal/config/config.go#L112-L141))
  - [x] namespace フラグによるテナント切り替え（デフォルト: default）
  - [x] ローカルファイルシステムからの読み込み ([cmd/cmd_store/get.go:76-107](cmd/cmd_store/get.go#L76-L107))
  - [x] Vault からの読み込み ([cmd/cmd_store/get.go:109-153](cmd/cmd_store/get.go#L109-L153))
  - [x] dotenv transform のサポート
  - [x] JSON 出力形式 ([cmd/cmd_store/get.go:165-175](cmd/cmd_store/get.go#L165-L175))
  - [x] raw 出力形式（key=value） ([cmd/cmd_store/get.go:177-194](cmd/cmd_store/get.go#L177-L194))
  - [x] 設定ファイル自動解決のサポート
- [x] store get コマンドのテスト ([cmd/cmd_store/get_test.go](cmd/cmd_store/get_test.go))
- [-] store ls コマンドの実装（不要）
- [-] store set コマンドの実装（不要）
- [-] store delete コマンドの実装（不要）
- [-] store load コマンドの実装（不要、store get と機能重複のため削除済み）

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
- [ ] 全ファイルシステムで共通の動作を確認するテスト
  - 提案: テーブル駆動テストで LocalFs, VaultFs, FsEnvFilesystem を同じテストケースで検証
  - 理由: 統一インターフェースの一貫性を保証
  - これはどこにおく？共通レベルのディレクトリが欲しい
  

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
- [ ] `internal/config/config.go` の責務を分割
  - 問題点: 設定ロード、パス解決、ストア取得が混在
  - 提案:
    - `ConfigLoader`: 設定ファイルのロード専用
    - `PathResolver`: パス文字列の解析専用
    - `StoreRegistry`: ストアの管理・取得専用
  - 理由: 単一責任の原則に従い、テストしやすくする

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
  - `store init`, `store get` の実行結果を検証
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