# 開発ガイド

kvtool の開発に参加する方向けのガイドです。

## 開発環境のセットアップ

### 必要なツール

- Go 1.21 以上
- Make
- Docker（統合テスト用）

### テスト実行

```bash
# 全テスト実行
make test

# Vault含む統合テスト
make test-full
```

### テスト用サービスの起動

```bash
# MinIO (S3互換) 起動
make minio-up           # http://localhost:9000
make minio-down

# Vault 起動
make vault-up           # http://localhost:8200
make vault-down
```

## コード規約

詳細は [CLAUDE.md](../CLAUDE.md) を参照してください。

### テスト規約

- **テーブル駆動テスト**: 全てのテストはテーブル駆動テストで記述
- **日本語**: テスト名、コメント、エラーメッセージは日本語で記述

### セキュリティ規約

- **パストラバーサル防止**: 全てのファイルシステムは `root` より上位に遡れないようにする
- **相対パスのみ**: 絶対パス（`/etc/passwd`）やチルダパス（`~/config`）は拒否

## ドキュメント生成

```bash
# APIリファレンスを自動生成
make gen-docs
```

コード内のコメントと構造体タグから `docs/api-reference.md` が自動生成されます。

## コード品質

```bash
# 静的解析
make lint

# コードフォーマット
make format
```

## 新しいファイルシステムドライバーの追加

1. **仕様策定**: `docs/filesystems/{driver}.md` を作成
2. **構造体定義**: 設定構造体を定義（`doc`, `required`, `default`, `example` タグを含む）
3. **インターフェース実装**: `Filesystem` と `File` インターフェースを実装
4. **テスト**: `filesystems/integration_test.go` にテストケースを追加
5. **ドキュメント生成**: `make gen-docs` を実行

詳細は [CLAUDE.md](../CLAUDE.md) の「開発ワークフロー」セクションを参照してください。
