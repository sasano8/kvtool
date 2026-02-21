package config

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
)

// HCLStoreDefinition は個々のストア定義を表す
type HCLStoreDefinition struct {
	Driver   string                 `hcl:"driver"`
	Endpoint string                 `hcl:"endpoint,optional"`
	Token    string                 `hcl:"token,optional"`
	Path     string                 `hcl:"path,optional"`
	Root     string                 `hcl:"root,optional"`
	Bucket   string                 `hcl:"bucket,optional"`
	Args     map[string]interface{} `hcl:"-"` // 追加の引数
}

// LoadHCLConfig は .kvtool.hcl ファイルをロードし、KvtoolConfig に変換する
func LoadHCLConfig(path string) (*KvtoolConfig, error) {
	parser := hclparse.NewParser()

	// ファイルをパース
	file, diags := parser.ParseHCLFile(path)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to parse HCL file: %s", diags.Error())
	}

	// 評価コンテキストを作成（環境変数と local 変数用）
	evalCtx := &hcl.EvalContext{
		Variables: map[string]cty.Value{
			"env": buildEnvMap(),
		},
	}

	// まず locals を解析して local 変数を構築
	localVars, err := parseLocals(file.Body, evalCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse locals: %w", err)
	}

	// local 変数を評価コンテキストに追加
	evalCtx.Variables["local"] = localVars

	// namespaces を解析
	config, err := parseNamespaces(file.Body, evalCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse namespaces: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// buildEnvMap は環境変数を cty.Value の map に変換する
func buildEnvMap() cty.Value {
	envMap := make(map[string]cty.Value)
	for _, env := range os.Environ() {
		for i := 0; i < len(env); i++ {
			if env[i] == '=' {
				key := env[:i]
				value := env[i+1:]
				envMap[key] = cty.StringVal(value)
				break
			}
		}
	}
	return cty.ObjectVal(envMap)
}

// parseLocals は locals ブロックを解析し、cty.Value として返す
func parseLocals(body hcl.Body, evalCtx *hcl.EvalContext) (cty.Value, error) {
	// HCL ファイルの構造を定義
	type hclFile struct {
		Locals []struct {
			Remain hcl.Body `hcl:",remain"`
		} `hcl:"locals,block"`
		Remain hcl.Body `hcl:",remain"`
	}

	var parsed hclFile
	diags := gohcl.DecodeBody(body, nil, &parsed)
	if diags.HasErrors() {
		return cty.NilVal, fmt.Errorf("failed to decode locals: %s", diags.Error())
	}

	localVars := make(map[string]cty.Value)

	for _, localsBlock := range parsed.Locals {
		// locals ブロック内の全ての属性を取得
		attrs, diags := localsBlock.Remain.JustAttributes()
		if diags.HasErrors() {
			return cty.NilVal, fmt.Errorf("failed to get locals attributes: %s", diags.Error())
		}

		for name, attr := range attrs {
			val, diags := attr.Expr.Value(evalCtx)
			if diags.HasErrors() {
				return cty.NilVal, fmt.Errorf("failed to evaluate local %s: %s", name, diags.Error())
			}
			localVars[name] = val
		}
	}

	if len(localVars) == 0 {
		return cty.EmptyObjectVal, nil
	}

	return cty.ObjectVal(localVars), nil
}

// parseNamespaces は namespaces ブロックを解析し、KvtoolConfig を返す
// YAML 形式と同じ構造:
//
//	namespaces {
//	  default {
//	    vault { ... }
//	  }
//	  production {
//	    vault { ... }
//	  }
//	}
//
// namespace 名はラベル形式で指定:
//
//	namespaces {
//	  namespace "default" {
//	    vault { ... }
//	  }
//	  namespace "production" {
//	    vault { ... }
//	  }
//	}
func parseNamespaces(body hcl.Body, evalCtx *hcl.EvalContext) (*KvtoolConfig, error) {
	config := &KvtoolConfig{
		Version:    1.0,
		Namespaces: make(map[string]map[string]StoreInfo),
	}

	// namespaces ブロックを取得
	content, _, diags := body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "namespaces"},
		},
	})
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to get namespaces block: %s", diags.Error())
	}

	// namespaces ブロックがない場合は default namespace を初期化して返す
	if len(content.Blocks) == 0 {
		config.Namespaces["default"] = make(map[string]StoreInfo)
		return config, nil
	}

	// namespaces ブロック内の各 namespace を解析
	for _, namespacesBlock := range content.Blocks {
		// namespace ブロック（ラベル付き）を取得
		nsContent, _, diags := namespacesBlock.Body.PartialContent(&hcl.BodySchema{
			Blocks: []hcl.BlockHeaderSchema{
				{Type: "namespace", LabelNames: []string{"name"}},
			},
		})
		if diags.HasErrors() {
			return nil, fmt.Errorf("failed to get namespace blocks: %s", diags.Error())
		}

		for _, nsBlock := range nsContent.Blocks {
			namespace := nsBlock.Labels[0]

			// namespace を初期化
			if _, exists := config.Namespaces[namespace]; !exists {
				config.Namespaces[namespace] = make(map[string]StoreInfo)
			}

			// namespace ブロック内のストア定義を取得
			storeContent, _, diags := nsBlock.Body.PartialContent(&hcl.BodySchema{
				Blocks: []hcl.BlockHeaderSchema{
					{Type: "vault"},
					{Type: "s3"},
					{Type: "local"},
					{Type: "env"},
					{Type: "http"},
					{Type: "db"},
				},
			})
			if diags.HasErrors() {
				return nil, fmt.Errorf("failed to get store blocks in namespace %s: %s", namespace, diags.Error())
			}

			for _, block := range storeContent.Blocks {
				storeInfo, err := parseStoreBlock(block, evalCtx)
				if err != nil {
					return nil, fmt.Errorf("failed to parse store %s in namespace %s: %w", block.Type, namespace, err)
				}
				config.Namespaces[namespace][block.Type] = storeInfo
			}
		}
	}

	// namespaces ブロックがあるが空の場合は default namespace を初期化
	if len(config.Namespaces) == 0 {
		config.Namespaces["default"] = make(map[string]StoreInfo)
	}

	return config, nil
}

// parseStoreBlock は個々のストアブロックを解析する
// YAML 形式との整合性のため、args ブロックと mount ブロックをサポート
func parseStoreBlock(block *hcl.Block, evalCtx *hcl.EvalContext) (StoreInfo, error) {
	storeInfo := StoreInfo{
		Driver: block.Type, // ブロックタイプがデフォルトのドライバー名
		Args:   make(map[string]interface{}),
	}

	// ブロック内の構造を解析
	content, _, diags := block.Body.PartialContent(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{Name: "driver", Required: false},
		},
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "args"},
			{Type: "mount"},
			{Type: "context"},
		},
	})
	if diags.HasErrors() {
		return StoreInfo{}, fmt.Errorf("failed to parse store block: %s", diags.Error())
	}

	// driver 属性を処理
	if attr, exists := content.Attributes["driver"]; exists {
		val, diags := attr.Expr.Value(evalCtx)
		if diags.HasErrors() {
			return StoreInfo{}, fmt.Errorf("failed to evaluate driver: %s", diags.Error())
		}
		storeInfo.Driver = val.AsString()
	}

	// args ブロックを処理
	for _, argsBlock := range content.Blocks {
		if argsBlock.Type == "args" {
			args, err := parseArgsBlock(argsBlock, evalCtx)
			if err != nil {
				return StoreInfo{}, fmt.Errorf("failed to parse args block: %w", err)
			}
			// 既存の Args にマージ
			for k, v := range args {
				storeInfo.Args[k] = v
			}
		}
	}

	// mount ブロックを処理
	for _, mountBlock := range content.Blocks {
		if mountBlock.Type == "mount" {
			mount, err := parseMountBlock(mountBlock, evalCtx)
			if err != nil {
				return StoreInfo{}, fmt.Errorf("failed to parse mount block: %w", err)
			}
			storeInfo.Mount = mount
		}
	}

	// context ブロックを処理
	for _, ctxBlock := range content.Blocks {
		if ctxBlock.Type == "context" {
			ctx, err := parseContextBlock(ctxBlock, evalCtx)
			if err != nil {
				return StoreInfo{}, fmt.Errorf("failed to parse context block: %w", err)
			}
			storeInfo.Context = ctx
		}
	}

	// ブロック形式を使わない場合（フラットな属性指定）も許可
	// 残りの属性を args として処理
	remainAttrs, diags := block.Body.JustAttributes()
	if !diags.HasErrors() {
		for name, attr := range remainAttrs {
			if name == "driver" {
				continue // 既に処理済み
			}
			val, diags := attr.Expr.Value(evalCtx)
			if diags.HasErrors() {
				return StoreInfo{}, fmt.Errorf("failed to evaluate attribute %s: %s", name, diags.Error())
			}
			storeInfo.Args[name] = ctyToGo(val)
		}
	}

	return storeInfo, nil
}

// parseArgsBlock は args ブロックを解析する
func parseArgsBlock(block *hcl.Block, evalCtx *hcl.EvalContext) (map[string]interface{}, error) {
	attrs, diags := block.Body.JustAttributes()
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to get args attributes: %s", diags.Error())
	}

	result := make(map[string]interface{})
	for name, attr := range attrs {
		val, diags := attr.Expr.Value(evalCtx)
		if diags.HasErrors() {
			return nil, fmt.Errorf("failed to evaluate %s: %s", name, diags.Error())
		}
		result[name] = ctyToGo(val)
	}

	return result, nil
}

// parseMountBlock は mount ブロックを解析する
func parseMountBlock(block *hcl.Block, evalCtx *hcl.EvalContext) (*MountInfo, error) {
	attrs, diags := block.Body.JustAttributes()
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to get mount attributes: %s", diags.Error())
	}

	mount := &MountInfo{}

	if attr, exists := attrs["dir"]; exists {
		val, diags := attr.Expr.Value(evalCtx)
		if diags.HasErrors() {
			return nil, fmt.Errorf("failed to evaluate dir: %s", diags.Error())
		}
		s := val.AsString()
		mount.Dir = &s
	}

	if attr, exists := attrs["file"]; exists {
		val, diags := attr.Expr.Value(evalCtx)
		if diags.HasErrors() {
			return nil, fmt.Errorf("failed to evaluate file: %s", diags.Error())
		}
		s := val.AsString()
		mount.File = &s
	}

	return mount, nil
}

// parseContextBlock は context ブロックを解析する
func parseContextBlock(block *hcl.Block, evalCtx *hcl.EvalContext) (*ContextInfo, error) {
	attrs, diags := block.Body.JustAttributes()
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to get context attributes: %s", diags.Error())
	}

	ctx := &ContextInfo{}

	if attr, exists := attrs["timeout"]; exists {
		val, diags := attr.Expr.Value(evalCtx)
		if diags.HasErrors() {
			return nil, fmt.Errorf("failed to evaluate timeout: %s", diags.Error())
		}
		v := ctyToGo(val)
		switch t := v.(type) {
		case int64:
			i := int(t)
			ctx.Timeout = &i
		case float64:
			i := int(t)
			ctx.Timeout = &i
		default:
			return nil, fmt.Errorf("timeout must be a number (seconds), got %T", v)
		}
	}

	return ctx, nil
}

// ctyToGo は cty.Value を Go の値に変換する
func ctyToGo(val cty.Value) interface{} {
	if val.IsNull() {
		return nil
	}

	switch {
	case val.Type() == cty.String:
		return val.AsString()
	case val.Type() == cty.Number:
		bf := val.AsBigFloat()
		if bf.IsInt() {
			i, _ := bf.Int64()
			return i
		}
		f, _ := bf.Float64()
		return f
	case val.Type() == cty.Bool:
		return val.True()
	case val.Type().IsListType() || val.Type().IsTupleType():
		var result []interface{}
		for it := val.ElementIterator(); it.Next(); {
			_, v := it.Element()
			result = append(result, ctyToGo(v))
		}
		return result
	case val.Type().IsMapType() || val.Type().IsObjectType():
		result := make(map[string]interface{})
		for it := val.ElementIterator(); it.Next(); {
			k, v := it.Element()
			result[k.AsString()] = ctyToGo(v)
		}
		return result
	default:
		return val.GoString()
	}
}

// ResolveHCLConfigPath は .kvtool.hcl ファイルのパスを解決する
func ResolveHCLConfigPath(explicitPath string) (string, error) {
	// 1. 明示的に指定された場合
	if explicitPath != "" && explicitPath != ".kvtool.hcl" {
		if _, err := os.Stat(explicitPath); err == nil {
			return explicitPath, nil
		}
		return "", fmt.Errorf("config file not found at specified path: %s", explicitPath)
	}

	// 2. ローカルの .kvtool.hcl をチェック
	localPath := ".kvtool.hcl"
	if _, err := os.Stat(localPath); err == nil {
		return localPath, nil
	}

	return "", fmt.Errorf("config file not found: %s", localPath)
}

// LoadHCLConfigAuto は自動パス解決で HCL 設定をロードする
func LoadHCLConfigAuto(explicitPath string) (*KvtoolConfig, error) {
	path, err := ResolveHCLConfigPath(explicitPath)
	if err != nil {
		return nil, err
	}

	return LoadHCLConfig(path)
}
