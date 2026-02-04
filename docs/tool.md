# Tool コマンド

`kvtool tool` は UUID、タイムスタンプ、ランダム値、パスワードなどを生成するユーティリティコマンドです。

## 使用方法

```bash
kvtool tool <tool-name>
```

## 利用可能なツール

### UUID

```bash
# UUID v7（時間ベース、ソート可能）
kvtool tool uuid7
# => 019c299f-c6ba-72dd-8d73-8bbe712f4cf0

# UUID v4（ランダム）
kvtool tool uuid4
# => 565ccde7-1f06-4c2c-92ff-6782568cd8a7
```

### タイムスタンプ

```bash
# 現在時刻（UTC ISO8601）
kvtool tool now
# => 2024-01-15T10:30:00Z

# Unix タイムスタンプ（秒）
kvtool tool timestamp
# => 1705315800
```

### ランダム値

```bash
# ランダムバイト（hex エンコード）
kvtool tool random/hex/16
# => a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6

# ランダムバイト（base64 エンコード）
kvtool tool random/base64/32
# => SGVsbG8gV29ybGQhIFRoaXMgaXMgYSB0ZXN0
```

バイト数は 1〜1024 の範囲で指定できます。

### パスワード

```bash
# 16文字のパスワード
kvtool tool password
# => xK9#mP2$nQ7@wL4!
```

パスワードは以下の文字セットから生成されます:
- 小文字: a-z
- 大文字: A-Z
- 数字: 0-9
- 記号: !@#$%^&*

## ユースケース

### CI/CD でのシークレット生成

```bash
# デプロイ用の一意な ID を生成
DEPLOY_ID=$(kvtool tool uuid7)

# API キーを生成
API_KEY=$(kvtool tool random/hex/32)
```

### スクリプトでの使用

```bash
#!/bin/bash

# ログファイル名にタイムスタンプを付与
LOG_FILE="app-$(kvtool tool timestamp).log"

# セッション ID を生成
SESSION_ID=$(kvtool tool uuid7)
```

### 環境変数の設定

```bash
export SECRET_KEY=$(kvtool tool random/base64/32)
export DB_PASSWORD=$(kvtool tool password)
```
