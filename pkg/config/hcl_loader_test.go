package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadHCLConfig_Basic(t *testing.T) {
	require := require.New(t)

	// テスト用 HCL ファイルを作成
	tmpDir := t.TempDir()
	hclPath := filepath.Join(tmpDir, ".kvtool.hcl")

	hclContent := `
locals {
  config = {
    environment = "development"
    port        = 8080
  }
}

namespaces {
  namespace "default" {
    vault {
      driver   = "vault"
      endpoint = "http://localhost:8200"
      token    = "root"
      path     = "secret/app"
    }
  }
}
`
	err := os.WriteFile(hclPath, []byte(hclContent), 0644)
	require.NoError(err)

	// ロード
	config, err := LoadHCLConfig(hclPath)
	require.NoError(err)
	require.NotNil(config)

	// 検証
	require.Equal(1.0, config.Version)
	require.Contains(config.Namespaces, "default")

	// vault ストアの確認
	vault, exists := config.Namespaces["default"]["vault"]
	require.True(exists, "vault store should exist")
	require.Equal("vault", vault.Driver)
	require.Equal("http://localhost:8200", vault.Args["endpoint"])
	require.Equal("root", vault.Args["token"])
	require.Equal("secret/app", vault.Args["path"])
}

func TestLoadHCLConfig_WithEnvVar(t *testing.T) {
	require := require.New(t)

	// 環境変数を設定
	os.Setenv("TEST_VAULT_ADDR", "http://vault.example.com:8200")
	os.Setenv("TEST_VAULT_TOKEN", "my-secret-token")
	defer os.Unsetenv("TEST_VAULT_ADDR")
	defer os.Unsetenv("TEST_VAULT_TOKEN")

	// テスト用 HCL ファイルを作成
	tmpDir := t.TempDir()
	hclPath := filepath.Join(tmpDir, ".kvtool.hcl")

	hclContent := `
namespaces {
  namespace "default" {
    vault {
      driver   = "vault"
      endpoint = env.TEST_VAULT_ADDR
      token    = env.TEST_VAULT_TOKEN
    }
  }
}
`
	err := os.WriteFile(hclPath, []byte(hclContent), 0644)
	require.NoError(err)

	// ロード
	config, err := LoadHCLConfig(hclPath)
	require.NoError(err)
	require.NotNil(config)

	// 検証
	vault := config.Namespaces["default"]["vault"]
	require.Equal("http://vault.example.com:8200", vault.Args["endpoint"])
	require.Equal("my-secret-token", vault.Args["token"])
}

func TestLoadHCLConfig_WithLocalVar(t *testing.T) {
	require := require.New(t)

	// テスト用 HCL ファイルを作成
	tmpDir := t.TempDir()
	hclPath := filepath.Join(tmpDir, ".kvtool.hcl")

	hclContent := `
locals {
  config = {
    environment = "production"
  }
}

namespaces {
  namespace "default" {
    vault {
      driver = "vault"
      path   = "secret/${local.config.environment}/app"
    }
  }
}
`
	err := os.WriteFile(hclPath, []byte(hclContent), 0644)
	require.NoError(err)

	// ロード
	config, err := LoadHCLConfig(hclPath)
	require.NoError(err)
	require.NotNil(config)

	// 検証
	vault := config.Namespaces["default"]["vault"]
	require.Equal("secret/production/app", vault.Args["path"])
}

func TestLoadHCLConfig_WithEnvAndLocalVar(t *testing.T) {
	require := require.New(t)

	// 環境変数を設定
	os.Setenv("TEST_VAULT_ADDR", "http://vault.prod.example.com:8200")
	defer os.Unsetenv("TEST_VAULT_ADDR")

	// テスト用 HCL ファイルを作成
	tmpDir := t.TempDir()
	hclPath := filepath.Join(tmpDir, ".kvtool.hcl")

	hclContent := `
locals {
  config = {
    environment = "staging"
    port        = 8080
  }
}

namespaces {
  namespace "default" {
    vault {
      driver   = "vault"
      endpoint = env.TEST_VAULT_ADDR
      path     = "secret/${local.config.environment}/app"
    }

    local {
      driver = "local"
      root   = "./configs/${local.config.environment}"
    }
  }
}
`
	err := os.WriteFile(hclPath, []byte(hclContent), 0644)
	require.NoError(err)

	// ロード
	config, err := LoadHCLConfig(hclPath)
	require.NoError(err)
	require.NotNil(config)

	// vault ストアの検証
	vault := config.Namespaces["default"]["vault"]
	require.Equal("vault", vault.Driver)
	require.Equal("http://vault.prod.example.com:8200", vault.Args["endpoint"])
	require.Equal("secret/staging/app", vault.Args["path"])

	// local ストアの検証
	local := config.Namespaces["default"]["local"]
	require.Equal("local", local.Driver)
	require.Equal("./configs/staging", local.Args["root"])
}

func TestLoadHCLConfig_MultipleLocalsBlocks(t *testing.T) {
	require := require.New(t)

	// テスト用 HCL ファイルを作成
	tmpDir := t.TempDir()
	hclPath := filepath.Join(tmpDir, ".kvtool.hcl")

	hclContent := `
locals {
  environment = "development"
}

locals {
  port = 8080
}

namespaces {
  namespace "default" {
    local {
      driver = "local"
      root   = "./${local.environment}"
    }
  }
}
`
	err := os.WriteFile(hclPath, []byte(hclContent), 0644)
	require.NoError(err)

	// ロード
	config, err := LoadHCLConfig(hclPath)
	require.NoError(err)
	require.NotNil(config)

	// 検証
	local := config.Namespaces["default"]["local"]
	require.Equal("./development", local.Args["root"])
}

func TestLoadHCLConfig_S3Store(t *testing.T) {
	require := require.New(t)

	// 環境変数を設定
	os.Setenv("TEST_S3_ENDPOINT", "http://minio:9000")
	defer os.Unsetenv("TEST_S3_ENDPOINT")

	// テスト用 HCL ファイルを作成
	tmpDir := t.TempDir()
	hclPath := filepath.Join(tmpDir, ".kvtool.hcl")

	hclContent := `
locals {
  bucket_name = "my-config-bucket"
}

namespaces {
  namespace "default" {
    s3 {
      driver   = "s3"
      endpoint = env.TEST_S3_ENDPOINT
      bucket   = local.bucket_name
      root     = "configs"
    }
  }
}
`
	err := os.WriteFile(hclPath, []byte(hclContent), 0644)
	require.NoError(err)

	// ロード
	config, err := LoadHCLConfig(hclPath)
	require.NoError(err)
	require.NotNil(config)

	// 検証
	s3 := config.Namespaces["default"]["s3"]
	require.Equal("s3", s3.Driver)
	require.Equal("http://minio:9000", s3.Args["endpoint"])
	require.Equal("my-config-bucket", s3.Args["bucket"])
	require.Equal("configs", s3.Args["root"])
}

func TestLoadHCLConfig_EmptyNamespaces(t *testing.T) {
	require := require.New(t)

	// テスト用 HCL ファイルを作成
	tmpDir := t.TempDir()
	hclPath := filepath.Join(tmpDir, ".kvtool.hcl")

	hclContent := `
locals {
  environment = "test"
}

namespaces {
}
`
	err := os.WriteFile(hclPath, []byte(hclContent), 0644)
	require.NoError(err)

	// ロード
	config, err := LoadHCLConfig(hclPath)
	require.NoError(err)
	require.NotNil(config)

	// 検証 - 空の namespaces ブロックがある場合は default namespace が作成される
	require.Contains(config.Namespaces, "default")
	require.Empty(config.Namespaces["default"])
}

func TestLoadHCLConfig_NoNamespacesBlock(t *testing.T) {
	require := require.New(t)

	// テスト用 HCL ファイルを作成（namespaces ブロックなし）
	tmpDir := t.TempDir()
	hclPath := filepath.Join(tmpDir, ".kvtool.hcl")

	hclContent := `
locals {
  environment = "test"
}
`
	err := os.WriteFile(hclPath, []byte(hclContent), 0644)
	require.NoError(err)

	// ロード
	config, err := LoadHCLConfig(hclPath)
	require.NoError(err)
	require.NotNil(config)

	// 検証 - namespaces ブロックがない場合は default namespace が作成される
	require.Contains(config.Namespaces, "default")
	require.Empty(config.Namespaces["default"])
}

func TestLoadHCLConfig_InvalidSyntax(t *testing.T) {
	require := require.New(t)

	// テスト用 HCL ファイルを作成
	tmpDir := t.TempDir()
	hclPath := filepath.Join(tmpDir, ".kvtool.hcl")

	hclContent := `
this is not valid HCL
`
	err := os.WriteFile(hclPath, []byte(hclContent), 0644)
	require.NoError(err)

	// ロード
	_, err = LoadHCLConfig(hclPath)
	require.Error(err)
	require.Contains(err.Error(), "failed to parse HCL file")
}

func TestLoadHCLConfig_UndefinedEnvVar(t *testing.T) {
	require := require.New(t)

	// 環境変数が未定義であることを確認
	os.Unsetenv("UNDEFINED_ENV_VAR")

	// テスト用 HCL ファイルを作成
	tmpDir := t.TempDir()
	hclPath := filepath.Join(tmpDir, ".kvtool.hcl")

	hclContent := `
namespaces {
  namespace "default" {
    vault {
      driver   = "vault"
      endpoint = env.UNDEFINED_ENV_VAR
    }
  }
}
`
	err := os.WriteFile(hclPath, []byte(hclContent), 0644)
	require.NoError(err)

	// ロード - 未定義の環境変数はエラーになる
	_, err = LoadHCLConfig(hclPath)
	require.Error(err)
}

func TestResolveHCLConfigPath(t *testing.T) {
	require := require.New(t)

	// テスト用ディレクトリを作成
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	// .kvtool.hcl が存在しない場合
	_, err := ResolveHCLConfigPath("")
	require.Error(err)

	// .kvtool.hcl を作成
	hclPath := filepath.Join(tmpDir, ".kvtool.hcl")
	err = os.WriteFile(hclPath, []byte("namespaces {}"), 0644)
	require.NoError(err)

	// .kvtool.hcl が存在する場合
	path, err := ResolveHCLConfigPath("")
	require.NoError(err)
	require.Equal(".kvtool.hcl", path)

	// 明示的なパスを指定
	customPath := filepath.Join(tmpDir, "custom.hcl")
	err = os.WriteFile(customPath, []byte("namespaces {}"), 0644)
	require.NoError(err)

	path, err = ResolveHCLConfigPath(customPath)
	require.NoError(err)
	require.Equal(customPath, path)
}

func TestCtyToGo(t *testing.T) {
	require := require.New(t)

	// テスト用 HCL ファイルを作成（様々な型をテスト）
	tmpDir := t.TempDir()
	hclPath := filepath.Join(tmpDir, ".kvtool.hcl")

	hclContent := `
namespaces {
  namespace "default" {
    local {
      driver  = "local"
      root    = "./data"
      enabled = true
      port    = 8080
    }
  }
}
`
	err := os.WriteFile(hclPath, []byte(hclContent), 0644)
	require.NoError(err)

	// ロード
	config, err := LoadHCLConfig(hclPath)
	require.NoError(err)

	local := config.Namespaces["default"]["local"]
	require.Equal("./data", local.Args["root"])
	require.Equal(true, local.Args["enabled"])
	require.Equal(int64(8080), local.Args["port"])
}

func TestLoadHCLConfig_WithArgsBlock(t *testing.T) {
	require := require.New(t)

	// 環境変数を設定
	os.Setenv("TEST_VAULT_ADDR", "http://vault.example.com:8200")
	os.Setenv("TEST_VAULT_TOKEN", "my-token")
	defer os.Unsetenv("TEST_VAULT_ADDR")
	defer os.Unsetenv("TEST_VAULT_TOKEN")

	// テスト用 HCL ファイルを作成（args ブロック形式）
	tmpDir := t.TempDir()
	hclPath := filepath.Join(tmpDir, ".kvtool.hcl")

	hclContent := `
locals {
  config = {
    environment = "production"
  }
}

namespaces {
  namespace "default" {
    vault {
      driver = "vault"
      args {
        endpoint = env.TEST_VAULT_ADDR
        token    = env.TEST_VAULT_TOKEN
        path     = "secret/${local.config.environment}/app"
      }
      mount {
        dir  = "app"
        file = "prod"
      }
    }
  }
}
`
	err := os.WriteFile(hclPath, []byte(hclContent), 0644)
	require.NoError(err)

	// ロード
	config, err := LoadHCLConfig(hclPath)
	require.NoError(err)
	require.NotNil(config)

	// 検証
	vault := config.Namespaces["default"]["vault"]
	require.Equal("vault", vault.Driver)
	require.Equal("http://vault.example.com:8200", vault.Args["endpoint"])
	require.Equal("my-token", vault.Args["token"])
	require.Equal("secret/production/app", vault.Args["path"])

	// mount の検証
	require.NotNil(vault.Mount)
	require.NotNil(vault.Mount.Dir)
	require.Equal("app", *vault.Mount.Dir)
	require.NotNil(vault.Mount.File)
	require.Equal("prod", *vault.Mount.File)
}

func TestLoadHCLConfig_WithArgsAndMountBlocks(t *testing.T) {
	require := require.New(t)

	// テスト用 HCL ファイルを作成（YAML 形式と同等の構造）
	tmpDir := t.TempDir()
	hclPath := filepath.Join(tmpDir, ".kvtool.hcl")

	hclContent := `
namespaces {
  namespace "default" {
    local {
      driver = "local"
      args {
        root = "./configs"
        transform = {
          read  = "dotenv"
          write = "dotenv"
        }
      }
      mount {
        file = ""
      }
    }

    env {
      driver = "env"
      args {}
      mount {
        file = ""
      }
    }
  }
}
`
	err := os.WriteFile(hclPath, []byte(hclContent), 0644)
	require.NoError(err)

	// ロード
	config, err := LoadHCLConfig(hclPath)
	require.NoError(err)
	require.NotNil(config)

	// local ストアの検証
	local := config.Namespaces["default"]["local"]
	require.Equal("local", local.Driver)
	require.Equal("./configs", local.Args["root"])

	// transform の検証（map 形式）
	transform, ok := local.Args["transform"].(map[string]interface{})
	require.True(ok, "transform should be a map")
	require.Equal("dotenv", transform["read"])
	require.Equal("dotenv", transform["write"])

	// mount の検証
	require.NotNil(local.Mount)
	require.NotNil(local.Mount.File)
	require.Equal("", *local.Mount.File)

	// env ストアの検証
	env := config.Namespaces["default"]["env"]
	require.Equal("env", env.Driver)
	require.NotNil(env.Mount)
}

func TestLoadHCLConfig_WithNamespace(t *testing.T) {
	require := require.New(t)

	// 環境変数を設定
	os.Setenv("TEST_VAULT_ADDR_DEV", "http://vault-dev:8200")
	os.Setenv("TEST_VAULT_ADDR_PROD", "http://vault-prod:8200")
	defer os.Unsetenv("TEST_VAULT_ADDR_DEV")
	defer os.Unsetenv("TEST_VAULT_ADDR_PROD")

	// テスト用 HCL ファイルを作成（複数 namespace）
	tmpDir := t.TempDir()
	hclPath := filepath.Join(tmpDir, ".kvtool.hcl")

	hclContent := `
locals {
  app_name = "myapp"
}

namespaces {
  namespace "default" {
    vault {
      driver = "vault"
      args {
        endpoint = env.TEST_VAULT_ADDR_DEV
        path     = "secret/dev/${local.app_name}"
      }
    }
  }

  namespace "production" {
    vault {
      driver = "vault"
      args {
        endpoint = env.TEST_VAULT_ADDR_PROD
        path     = "secret/prod/${local.app_name}"
      }
    }

    s3 {
      driver = "s3"
      args {
        bucket = "prod-config"
      }
    }
  }

  namespace "staging" {
    local {
      driver = "local"
      args {
        root = "./staging-configs"
      }
    }
  }
}
`
	err := os.WriteFile(hclPath, []byte(hclContent), 0644)
	require.NoError(err)

	// ロード
	config, err := LoadHCLConfig(hclPath)
	require.NoError(err)
	require.NotNil(config)

	// namespace の確認
	require.Len(config.Namespaces, 3)
	require.Contains(config.Namespaces, "default")
	require.Contains(config.Namespaces, "production")
	require.Contains(config.Namespaces, "staging")

	// default namespace の検証
	defaultVault := config.Namespaces["default"]["vault"]
	require.Equal("vault", defaultVault.Driver)
	require.Equal("http://vault-dev:8200", defaultVault.Args["endpoint"])
	require.Equal("secret/dev/myapp", defaultVault.Args["path"])

	// production namespace の検証
	prodVault := config.Namespaces["production"]["vault"]
	require.Equal("vault", prodVault.Driver)
	require.Equal("http://vault-prod:8200", prodVault.Args["endpoint"])
	require.Equal("secret/prod/myapp", prodVault.Args["path"])

	prodS3 := config.Namespaces["production"]["s3"]
	require.Equal("s3", prodS3.Driver)
	require.Equal("prod-config", prodS3.Args["bucket"])

	// staging namespace の検証
	stagingLocal := config.Namespaces["staging"]["local"]
	require.Equal("local", stagingLocal.Driver)
	require.Equal("./staging-configs", stagingLocal.Args["root"])
}

func TestLoadHCLConfig_OnlyNamespacedStores(t *testing.T) {
	require := require.New(t)

	// テスト用 HCL ファイルを作成（全てラベル付き）
	tmpDir := t.TempDir()
	hclPath := filepath.Join(tmpDir, ".kvtool.hcl")

	hclContent := `
namespaces {
  namespace "dev" {
    local {
      driver = "local"
      args {
        root = "./dev"
      }
    }
  }

  namespace "prod" {
    local {
      driver = "local"
      args {
        root = "./prod"
      }
    }
  }
}
`
	err := os.WriteFile(hclPath, []byte(hclContent), 0644)
	require.NoError(err)

	// ロード
	config, err := LoadHCLConfig(hclPath)
	require.NoError(err)
	require.NotNil(config)

	// namespace の確認（default は作成されない）
	require.Len(config.Namespaces, 2)
	require.Contains(config.Namespaces, "dev")
	require.Contains(config.Namespaces, "prod")
	require.NotContains(config.Namespaces, "default")

	// 各 namespace の検証
	require.Equal("./dev", config.Namespaces["dev"]["local"].Args["root"])
	require.Equal("./prod", config.Namespaces["prod"]["local"].Args["root"])
}
