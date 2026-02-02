# コマンドリファレンス

## kvtool store

ストア機能の中核コマンド。設定ファイルで定義された複数のデータソースを統一的に扱います。

### kvtool store init

設定ファイル（`.kvtool.yml`）を初期化します。

#### 使用法

```bash
kvtool store init [path] [flags]
```

#### 引数

- `path` (オプション): 設定ファイルのパス。省略時は `.kvtool.yml`

#### フラグ

| フラグ | 短縮形 | デフォルト | 説明 |
|--------|--------|-----------|------|
| --force | -f | false | 既存ファイルを上書き |
| --global | -g | false | グローバル設定ファイルを作成（`~/.config/kvtool/.kvtool.yml`） |

#### 例

```bash
# ローカル設定ファイルを作成
kvtool store init

# グローバル設定ファイルを作成
kvtool store init --global

# カスタムパスに作成
kvtool store init myconfig.yml

# 既存ファイルを上書き
kvtool store init --force
```

### kvtool store load

ストアからファイルを取得します。

#### 使用法

```bash
kvtool store load <store_name>/<file_path> [flags]
```

#### 引数

- `store_name/file_path`: ストア名とファイルパス
  - `store_name`: 設定ファイルで定義したストア名
  - `file_path`: ストア内のファイルパス（オプション）

#### フラグ

| フラグ | 短縮形 | デフォルト | 説明 |
|--------|--------|-----------|------|
| --config | -c | .kvtool.yml | 設定ファイルのパス |
| --global | -g | false | グローバル設定を使用 |
| --namespace | -n | default | 使用する namespace |
| --output | -o | json | 出力形式（json, raw） |

#### 出力形式

##### json

JSON 形式で出力します（デフォルト）。

```bash
kvtool store load .env/test.env
```

出力例：
```json
{
  "APP_NAME": "myapp",
  "APP_ENV": "production"
}
```

##### raw

key=value 形式で出力します。

```bash
kvtool store load .env/test.env -o raw
```

出力例：
```
APP_NAME=myapp
APP_ENV=production
```

#### 例

```bash
# 基本的な使用
kvtool store load .env/test.env

# raw 形式で出力
kvtool store load .env/test.env -o raw

# 特定の namespace を指定
kvtool store load .env/app.env -n production

# グローバル設定を使用
kvtool store load .env/test.env --global

# カスタム設定ファイルを使用
kvtool store load .env/test.env -c /path/to/config.yml

# Vault からシークレットを取得
kvtool store load vault/db/password

# 複数の namespace を使い分け
kvtool store load .env/app.env -n development
kvtool store load .env/app.env -n production
```

## その他のコマンド

### kvtool load

各種データソースから直接データを読み込みます（ストア機能を使わない単発の操作）。

#### kvtool load env

環境変数を JSON 形式で出力します。

```bash
kvtool load env
```

#### kvtool load dotenv

.env ファイルを JSON 形式で出力します。

```bash
kvtool load dotenv -i test.env
```

#### kvtool load vault

Vault から直接シークレットを読み込みます。

```bash
kvtool load vault -addr http://localhost:8200 -token root -mount secret app/prod
```

### kvtool convert

JSON を他の形式に変換します。

```bash
# JSON を .env 形式に変換
cat config.json | kvtool convert dotenv

# JSON を YAML 形式に変換
cat config.json | kvtool convert yaml
```

## コマンドの設計思想

### ストア機能が中心

kvtool の中核機能は **ストア（store）** です。

- `kvtool store` コマンドが主要な操作
- `kvtool load` や `kvtool convert` は補助的な機能
- 設定ファイルで複数のデータソースを統一的に管理

### 明示的な操作

- **グローバル設定は明示的に指定**: 自動フォールバックしない
- **namespace は明示的に切り替え**: デフォルトは `default`
- **エラーメッセージで誘導**: 設定がない場合は `kvtool store init` を促す

### UNIX 哲学

- **標準入出力の活用**: パイプで他のコマンドと連携
- **シンプルな責務**: 各コマンドは1つの責務に集中
- **組み合わせ可能**: 他のツールと組み合わせて使える
