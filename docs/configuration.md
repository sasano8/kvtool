# 設定ファイル

kvtool の設定ファイルは YAML 形式 (`.kvtool.yml`) または HCL 形式 (`.kvtool.hcl`) で記述できます。

## 対応形式

| 形式 | ファイル名 | 特徴 |
|------|-----------|------|
| YAML | `.kvtool.yml` | シンプル、広く普及 |
| HCL | `.kvtool.hcl` | 変数定義、環境変数展開をサポート |

## 基本構造

```yaml
version: 0.1
namespaces:
  <namespace名>:
    <ストア名>:
      driver: <ドライバー名>
      args:
        <ドライバー固有の設定>
```

## 設定ファイルの配置

kvtool は以下の順序で設定ファイルを探します：

1. カレントディレクトリの `.kvtool.yml`
2. `--config` フラグで指定されたパス
3. グローバル設定: `~/.config/kvtool/.kvtool.yml`

## namespace

namespace は環境（開発、ステージング、本番など）ごとに設定を切り替えるための仕組みです。

```yaml
namespaces:
  default:
    local-env:
      driver: local
      args:
        root: ./config
  production:
    vault-prod:
      driver: vault
      args:
        addr: https://vault.example.com
        token: ${VAULT_TOKEN}
```

デフォルトは `default` namespace が使用されます。

## ドライバー

kvtool は以下のドライバーをサポートしています：

- **local**: ローカルファイルシステム
- **vault**: HashiCorp Vault
- **env**: 環境変数
- **s3**: Amazon S3 / S3互換ストレージ

各ドライバーの詳細は [API リファレンス](./api-reference.md) を参照してください。

## 設定例

### ローカルファイル

```yaml
local-env:
  driver: local
  args:
    root: ./config
    transform: dotenv  # .env形式をJSON変換
```

### HashiCorp Vault

```yaml
vault-prod:
  driver: vault
  args:
    addr: http://localhost:8200
    token: root
    mount: secret
    namespace: admin
```

### Amazon S3

```yaml
s3-config:
  driver: s3
  args:
    bucket: my-config-bucket
    region: us-east-1
    endpoint: http://localhost:9000  # MinIOなどS3互換ストレージ用
    access_key_id: minioadmin
    secret_access_key: minioadmin
    use_path_style: true
```

### 環境変数

```yaml
env-vars:
  driver: env
  args: {}  # 設定不要
```

## HCL 形式 (.kvtool.hcl)

HCL (HashiCorp Configuration Language) 形式は、変数定義や環境変数展開など、より高度な機能をサポートします。

### 基本構造

```hcl
# 変数定義
locals {
  config = {
    environment = "development"
    port        = 8080
  }
}

# namespace 定義
namespaces {
  namespace "default" {
    <ストア名> {
      driver = "<ドライバー名>"
      args {
        <ドライバー固有の設定>
      }
      mount {
        dir  = "<マウントディレクトリ>"
        file = "<マウントファイル>"
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
| 環境変数 | `env.VAR_NAME` | 環境変数を直接参照（kvtool 拡張） |
| 文字列補間 | `"${local.var}"` | 文字列内で変数を展開 |

### HCL 設定例

#### 複数環境の設定

```hcl
locals {
  config = {
    environment = "development"
  }
}

namespaces {
  namespace "default" {
    vault {
      driver = "vault"
      args {
        endpoint = env.VAULT_ADDR
        token    = env.VAULT_TOKEN
        path     = "secret/${local.config.environment}/app"
      }
      mount {
        dir  = "app"
        file = "prod"
      }
    }

    local {
      driver = "local"
      args {
        root = "./configs/${local.config.environment}"
        transform = {
          read  = "dotenv"
          write = "dotenv"
        }
      }
    }

    env {
      driver = "env"
      args {}
    }
  }

  namespace "production" {
    vault {
      driver = "vault"
      args {
        endpoint = "https://vault.prod.example.com"
        token    = env.VAULT_TOKEN
        path     = "secret/production/app"
      }
    }

    s3 {
      driver = "s3"
      args {
        endpoint = "https://s3.amazonaws.com"
        bucket   = "prod-config-bucket"
        root     = "configs"
      }
    }
  }
}
```

### YAML と HCL の対応

| YAML | HCL |
|------|-----|
| `namespaces:` | `namespaces {` |
| `default:` | `namespace "default" {` |
| `vault:` | `vault {` |
| `driver: vault` | `driver = "vault"` |
| `args:` | `args {` |
| `${VAR}` | `env.VAR` |

### HCL 設定ファイルの配置

kvtool は以下の順序で HCL 設定ファイルを探します：

1. カレントディレクトリの `.kvtool.hcl`
2. `--config` フラグで指定されたパス
