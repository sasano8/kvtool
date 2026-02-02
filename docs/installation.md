# インストール

## ビルドとインストール

### Makefileを使用

```bash
# ビルドとインストール
make install
```

### 直接ビルド

```bash
# バイナリをbin/ディレクトリに生成
go build -o bin/kvtool .
```

## 動作確認

```bash
kvtool --version
kvtool --help
```

## 必要な環境

- Go 1.21 以上
- （オプション）HashiCorp Vault（Vault 連携時）
- （オプション）AWS CLI/認証情報（S3 連携時）
