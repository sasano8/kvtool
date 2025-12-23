# tool-kv

# Getting Started

```
make install
```

## env to json

```
kvtool env2json
```

## json to .env

```
kvtool json2env -i test_data/json/simple.json
```

## .env to json

```
kvtool dotenv2json -i test_data/dot_env/simple.env
```


## use store

次のコマンドでストアコンフィグ(`.kvtool.json`)を生成します。

```
kvtool init
```

ストアコンフィグは以下の構成になっています。
必要に応じて編集してください。

```
{
  "version": 0.1,
  "namespaces": {
    // 名前空間を定義します。default はデフォルトで指定される名前空間です。
    "default": {
      // キー（ルートパス）とタイプとそのタイプの引数を指定します
      ".env": {
        "type": ".env",
        "args": {
          "input": ".env"
        }
      }
    }
  }
}

```

ここでは ストアコンフィグ例に従い `.env` を作成します。

```
A=test
```

以下のように構成ファイルを読み込むことができます。

```
kvtool store -ns default .env
```
