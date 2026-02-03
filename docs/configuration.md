# ストア構成ファイル

kvtool のストア構成ファイルは YAML 形式 (`.kvtool.yml`) または HCL 形式 (`.kvtool.hcl`) で記述できます。

## 設定ファイルの配置

以下の順序で設定ファイルを探します：

1. カレントディレクトリの `.kvtool.yml` または `.kvtool.hcl`
2. `--config` フラグで指定されたパス
3. グローバル設定: `~/.config/kvtool/.kvtool.yml`

## 基本構造（YAML）

```yaml
version: 0.1
namespaces:
  default:
    local:
      driver: local
      args:
        root: ./config
    vault:
      driver: vault
      args:
        addr: http://localhost:8200
        token: ${VAULT_TOKEN}
  production:
    vault-prod:
      driver: vault
      args:
        addr: https://vault.example.com
```

- `namespaces`: 環境（開発、本番など）ごとに設定を切り替える仕組み
- デフォルトは `default` namespace を使用
- 各ドライバーの設定パラメータは [ドライバー](./api-reference.md) を参照

## HCL 形式

HCL は変数定義や環境変数展開をサポートします。

```hcl
locals {
  env = "development"
}

namespaces {
  namespace "default" {
    vault {
      driver = "vault"
      args {
        addr  = env.VAULT_ADDR
        token = env.VAULT_TOKEN
        path  = "secret/${local.env}/app"
      }
    }

    local {
      driver = "local"
      args {
        root = "./configs/${local.env}"
      }
    }
  }
}
```

### HCL 固有機能

| 機能 | 構文 | 説明 |
|------|------|------|
| 変数定義 | `locals { ... }` | 辞書形式で変数を定義 |
| 変数参照 | `local.config.port` | 属性アクセスで変数を参照 |
| 環境変数 | `env.VAR_NAME` | 環境変数を直接参照 |
| 文字列補間 | `"${local.var}"` | 文字列内で変数を展開 |

### YAML と HCL の対応

| YAML | HCL |
|------|-----|
| `namespaces:` | `namespaces {` |
| `default:` | `namespace "default" {` |
| `driver: vault` | `driver = "vault"` |
| `args:` | `args {` |
| `${VAR}` | `env.VAR` |
