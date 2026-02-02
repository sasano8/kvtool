# kvtool について

**kvtool** は、様々な設定ファイル形式（.env、JSON、YAML、Vault、S3 など）を統一的なインターフェースで扱う CLI ツールです。

## 特徴

- **統一インターフェース**: ローカルファイル、Vault、S3 などを同じ方法でアクセス
- **マルチテナント**: namespace による環境切り替え
- **フォーマット変換**: dotenv ↔ JSON ↔ YAML
- **接続確認**: ストアの接続テスト機能

## ユースケース

環境構築の初期段階で、様々な構成ファイルを統合して環境を初期化することを想定しています。

## コマンド体系

```
kvtool
├── file (低レベル: 直接アクセス)
│   ├── load (env, vault, json, yaml, etc.)
│   └── convert (dotenv, yaml)
└── store (高レベル: 設定ファイル経由)
    ├── init
    ├── connect
    ├── load
    └── serve (未実装)
```

## ライセンス

MIT
