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
- [ ] ローカルファイルシステムのテスト
- [ ] セキュリティ制限のテスト（パストラバーサル攻撃の防止）

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
- [x] 構成ファイルのロード機能の実装 ([internal/shared/config.go](internal/shared/config.go))
- [x] パスパーサーの実装 ([internal/shared/config.go:62-109](internal/shared/config.go#L62-L109))
- [x] パスパーサーのテスト ([internal/shared/config_test.go:45-98](internal/shared/config_test.go#L45-L98))
- [ ] 構成ファイルのバリデーション

### ストアの機能
- [x] ストアで様々なデータソースを定義できるようにする
- [x] 構成ファイルを読んで、キーバリューストレージとして機能させる
- [x] 設定ファイルの自動解決 ([internal/shared/config.go:45-95](internal/shared/config.go#L45-L95))
  - [x] ローカル設定 (.kvtool.yml) の優先
  - [x] グローバル設定 (~/.config/kvtool/.kvtool.yml) へのフォールバック
- [x] store init コマンドの実装 ([cmd/cmd_store/init.go](cmd/cmd_store/init.go))
  - [x] ローカル設定ファイルの生成
  - [x] グローバル設定ファイルの生成 (--global)
  - [x] 既存ファイルの上書き保護 (--force で上書き可能)
  - [x] YAML 形式での出力
- [x] store init コマンドのテスト ([cmd/cmd_store/init_test.go](cmd/cmd_store/init_test.go))
- [x] store get コマンドの実装 ([cmd/cmd_store/get.go](cmd/cmd_store/get.go))
  - [x] パス形式: `store_name/file_path` ([internal/shared/config.go:112-141](internal/shared/config.go#L112-L141))
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
- [ ] LocalFs の基本機能テスト
  - GetFile, LoadAsJson, OpenReader の動作確認
  - ファイルが存在しない場合のエラーハンドリング
  - 様々なファイル形式（JSON, dotenv）の読み込み

- [ ] セキュリティテスト
  - パストラバーサル攻撃の防止テスト（`../../../etc/passwd` など）
  - Root ディレクトリ外へのアクセス拒否の確認

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
- [ ] すべての Filesystem で context.Context を受け取るように統一
  - 問題点: VaultFs のみ context を扱い、LocalFs や FsEnvFilesystem は扱わない
  - 提案: `GetFile(ctx context.Context, path string)` に統一
  - 理由: タイムアウト、キャンセル処理の統一

#### 環境変数ファイルシステムの設定対応
- [ ] FsEnvFilesystem を .kvtool.yml で設定可能にする
  - 問題点: 現状、env ファイルシステムは設定ファイルから利用できない
  - 提案: driver: "env" を追加
  - 理由: 統一的な設定管理

### 優先度：中（重要だが複雑）

#### パッケージ構造の改善

- [ ] `internal/shared/` パッケージのリネーム
  - 問題点: `shared` は catch-all パッケージ名で意図が不明確
  - 提案: `internal/store/` または `internal/registry/` にリネーム
  - 理由: パッケージの責務（ストアの管理・登録）を名前で明確にする

- [ ] `filesystems/core.go` から設定構造体を分離
  - 問題点: `FilesystemConfig`, `VaultConfig`, `StoreConfig`, `MountConfig`, `TransformConfig` が filesystem パッケージに混在
  - 提案: `config/` パッケージを作成し、設定関連を移動
  - 理由: ファイルシステムの抽象化と設定管理は別の責務

#### ファイルシステムファクトリーの導入
- [ ] Filesystem の生成を統一するファクトリーパターンの実装
  - 問題点: コマンド層が直接ファイルシステムインスタンスを生成している
  - 提案: `FilesystemFactory` または `FilesystemRegistry` インターフェースの導入
  ```go
  type FilesystemFactory interface {
      Create(ctx context.Context, config *StoreConfig) (Filesystem, error)
  }
  ```
  - 理由: 設定から Filesystem への変換ロジックを一元管理

#### 統合インターフェーステスト
- [ ] 全ファイルシステムで共通の動作を確認するテスト
  - 提案: テーブル駆動テストで LocalFs, VaultFs, FsEnvFilesystem を同じテストケースで検証
  - 理由: 統一インターフェースの一貫性を保証

#### Transform の統一インターフェース化
- [ ] Transform を File インターフェースに統合
  - 問題点: `TransformConfig` が定義されているが、統一インターフェースに組み込まれていない
  - 提案: `File` インターフェースに `Transform` オプションを追加、または `TransformableFile` インターフェースを定義
  - 理由: dotenv などの変換処理を統一的に扱えるようにする

#### Store サービス層の導入
- [ ] コマンド層とファイルシステム層の間にサービス層を挿入
  - 問題点: コマンド層（cmd/）にビジネスロジックが混在
  - 提案: `internal/service/store.go` を作成し、以下を実装
  ```go
  type StoreService interface {
      Get(ctx context.Context, namespace, storePath string) (any, error)
      Set(ctx context.Context, namespace, storePath string, data any) error
      List(ctx context.Context, namespace, storePath string) ([]string, error)
  }
  ```
  - 理由: コマンドライン以外（API、ライブラリとしての利用）からも同じロジックを使える

#### 設定管理とパス解決の分離
- [ ] `internal/shared/config.go` の責務を分割
  - 問題点: 設定ロード、パス解決、ストア取得が混在
  - 提案:
    - `ConfigLoader`: 設定ファイルのロード専用
    - `PathResolver`: パス文字列の解析専用
    - `StoreRegistry`: ストアの管理・取得専用
  - 理由: 単一責任の原則に従い、テストしやすくする

#### デコーダとソースの統合検討
- [ ] `pkg/decoders/` と `pkg/sources/` の関係を整理
  - 問題点: 密結合だが分離されている
  - 提案: `pkg/transform/` パッケージとして統合、または `sources.Source` と `transform.Decoder` の明確な分離
  - 理由: データの取得（source）と変換（decoder）の責務を明確にする


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