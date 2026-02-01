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
- [ ] Vault ファイルシステムのテスト
- [ ] Vault からの読み込みの統合テスト

## インターフェースの統一

- [x] FsFile インターフェースの定義 ([filesystems/env.go:8-10](filesystems/env.go#L8-L10))
- [ ] ストリームで取得するインターフェース（io.Reader を返す）
- [ ] JSON デコード済みで取得するインターフェース（LoadAsJson）の統一
  - [x] VaultFsFile.LoadAsJson ([filesystems/vault.go:85](filesystems/vault.go#L85))
  - [x] LocalFs.LoadAsJson ([filesystems/local.go:99](filesystems/local.go#L99))
  - [x] FsEnvFile.LoadAsJson ([filesystems/env.go:15](filesystems/env.go#L15))
  - [ ] すべてのファイルシステムで統一されたインターフェースを実装
  - [ ] ストリーム取得インターフェース（OpenReader）の統一

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
- [x] VaultFs の実装 ([filesystems/vault.go:48-76](filesystems/vault.go#L48-L76))
- [x] VaultFsFile の実装 ([filesystems/vault.go:78-83](filesystems/vault.go#L78-L83))
- [ ] Vault ファイルシステムのテスト
- [ ] エラーハンドリングのテスト

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