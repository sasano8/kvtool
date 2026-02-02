# Filesystem ドライバー

kvtool は様々なファイルシステムドライバーをサポートしています。

## サポートドライバー

- **[S3 Filesystem](./s3.md)**: Amazon S3 および S3 互換ストレージ

## 共通インターフェース

全てのドライバーは以下のインターフェースを実装しています：

```go
type Filesystem interface {
    GetFile(path string) (File, error)
}

type File interface {
    LoadAsJson() (any, error)
    OpenReader() (io.ReadCloser, error)
}
```

## その他のドライバー

以下のドライバーも実装されています：

- **LocalFs**: ローカルファイルシステム
- **VaultFs**: HashiCorp Vault
- **EnvFs**: 環境変数

詳細は [API リファレンス](../api-reference.md) を参照してください。
