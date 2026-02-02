# .env ファイル仕様

kvtool における .env ファイルの読み書き仕様を定義します。

## 設計方針

- **Node.js dotenv ベース**: 最も広く使われている実装に準拠
- **行末コメント対応**: 実用性のために追加
- **セキュリティ優先**: 変数展開・コマンド置換は非サポート
- **読み取り専用**: 書き込み操作は提供しない（下記参照）

### なぜ書き込み操作を提供しないか

kvtool は設定の**読み取り専用**ツールです。書き込み操作（Write）は意図的にサポートしません。

**セキュリティ上の理由:**

```bash
# 例: もし HCL への書き込みが可能だった場合
kvtool convert env:// --to hcl > terraform.tfvars

# 環境変数に含まれる機密情報（API キー、パスワード等）が
# 意図せずファイルに書き出され、Git にコミットされるリスク
```

**設計意図:**

- 設定の「取得」と「変換」に特化
- 書き込みは既存のツール（`echo`, `cat`, エディタ）で十分
- 機密情報の意図しない永続化を防止

## 機能対応表

| 機能 | 対応 | 備考 |
|------|------|------|
| ダブルクォート | ✅ | エスケープ処理あり |
| シングルクォート | ✅ | リテラル（エスケープなし） |
| エスケープ `\n` | ✅ | 改行 |
| エスケープ `\r` | ✅ | CR（Windows互換） |
| エスケープ `\"` | ✅ | クォート内に `"` を含めるため |
| エスケープ `\t` | ❌ | Node.js 互換で非対応 |
| エスケープ `\\` | ❌ | Node.js 互換で非対応 |
| 行末コメント | ✅ | `KEY=value # comment` |
| キー名検証 (POSIX) | ✅ | `[a-zA-Z_][a-zA-Z0-9_]*` |
| 空のキー `=value` | ❌ | エラー（POSIX 準拠） |
| 数字で始まるキー | ❌ | エラー（POSIX 準拠） |
| `export` プレフィックス | ❌ | サポートしない |
| 変数展開 `${VAR}` | ❌ | リテラル扱い |
| コマンド置換 `$(cmd)` | ❌ | リテラル扱い |
| 複数行 (heredoc) | ❌ | サポートしない |

## 基本構文

### キー名のルール

キー名は **POSIX シェル変数名** に準拠します:

- パターン: `[a-zA-Z_][a-zA-Z0-9_]*`
- 英字（a-z, A-Z）またはアンダースコア（_）で開始
- 以降は英字、数字、アンダースコアのみ

```bash
# ✅ 有効なキー名
FOO=bar
_PRIVATE=value
MY_VAR_123=value

# ❌ 無効なキー名（エラーになる）
1KEY=value           # 数字で始まる
MY-VAR=value         # ハイフンを含む
MY.VAR=value         # ドットを含む
キー=value            # 非ASCII文字
```

### KEY=VALUE 形式

```bash
# 基本形式
KEY=value

# 空の値
EMPTY_KEY=

# 値に = を含む
DATABASE_URL=postgres://user:pass@host:5432/db
```

### クォート

```bash
# ダブルクォート: エスケープシーケンスが有効
MESSAGE="Hello\nWorld"        # Hello + 改行 + World

# シングルクォート: リテラル（エスケープなし）
MESSAGE='Hello\nWorld'        # Hello\nWorld（そのまま）

# クォートなし
MESSAGE=HelloWorld
```

### エスケープシーケンス

ダブルクォート内でのみ有効:

| シーケンス | 変換後 |
|-----------|--------|
| `\n` | 改行 (LF) |
| `\r` | キャリッジリターン (CR) |
| `\"` | ダブルクォート |

**非対応のエスケープ**:
- `\t` → そのまま `\t` として扱う
- `\\` → そのまま `\\` として扱う

### コメント

```bash
# 行頭コメント
KEY=value

KEY=value # 行末コメント（クォート外）

KEY="value # not a comment"  # クォート内は値の一部
```

## 非サポート機能

以下の機能は**意図的にサポートしません**:

### export プレフィックス

```bash
# ❌ サポートしない
export KEY=value

# ✅ 代わりにこう書く
KEY=value
```

**理由**: シンプルさを優先。bash での `source .env` が必要な場合は別途対応。

### 変数展開

```bash
# ❌ 展開しない（リテラル扱い）
BASE_URL=http://localhost
API_URL=${BASE_URL}/api    # → "${BASE_URL}/api" という文字列

# ✅ 展開が必要な場合は呼び出し側で処理
```

**理由**: セキュリティと予測可能性を優先。意図しない値の漏洩を防止。

### コマンド置換

```bash
# ❌ 実行しない（リテラル扱い）
DATE=$(date +%Y-%m-%d)     # → "$(date +%Y-%m-%d)" という文字列
```

**理由**: セキュリティを優先。任意のコマンド実行を防止。

### 複数行値 (heredoc)

```bash
# ❌ サポートしない
PRIVATE_KEY="""
-----BEGIN RSA PRIVATE KEY-----
...
-----END RSA PRIVATE KEY-----
"""

# ✅ 代わりにエスケープシーケンスを使用
PRIVATE_KEY="-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----"
```

**理由**: シンプルさを優先。複雑な値は JSON や Vault を推奨。

## 他ツールとの互換性

| ツール | 互換性 | 差異 |
|--------|--------|------|
| Node.js dotenv | ほぼ同等 | 行末コメント追加 |
| docker-compose | 部分的 | export, 変数展開なし |
| Python python-dotenv | 部分的 | 変数展開なし |

### Node.js dotenv との比較

```bash
# 両方で同じ動作
KEY="hello\nworld"           # → hello + 改行 + world

# kvtool のみ対応
KEY=value # comment          # → value（行末コメント）

# どちらも非対応（リテラル扱い）
KEY=${OTHER_KEY}             # → "${OTHER_KEY}"
```

## 使用例

### .env ファイルの読み込み

```bash
# ローカルファイルから
kvtool file load .env --transform dotenv

# ストア経由
kvtool store load config/.env
```

### 環境変数の出力

```bash
# 環境変数を .env 形式で出力
kvtool file load env:// --format raw
```

## エンコーディング

- **UTF-8 のみ**: BOM なし
- **改行コード**: LF (`\n`) を推奨、CRLF (`\r\n`) も許容
