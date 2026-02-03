package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// FieldInfo はフィールドの情報を保持します
type FieldInfo struct {
	Name         string
	Type         string
	YAMLName     string
	Doc          string
	Required     bool
	DefaultValue string
	Example      string
	Comment      string
}

// ConfigStructInfo は設定構造体の情報を保持します
type ConfigStructInfo struct {
	Name        string
	Comment     string
	Fields      []FieldInfo
	PackageName string
	FileName    string
}

func main() {
	driver := flag.String("driver", "", "特定のドライバーのみ生成（例: s3, local, vault）")
	flag.Parse()

	// filesystems ディレクトリを解析
	configs, err := parseFilesystems("./pkg/filesystems")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing filesystems: %v\n", err)
		os.Exit(1)
	}

	// ドライバーフィルタ
	if *driver != "" {
		filtered := []ConfigStructInfo{}
		for _, cfg := range configs {
			if strings.Contains(strings.ToLower(cfg.Name), strings.ToLower(*driver)) {
				filtered = append(filtered, cfg)
			}
		}
		configs = filtered
	}

	if len(configs) == 0 {
		fmt.Println("No configuration structures found.")
		return
	}

	// ドライバーリファレンスを生成
	if err := generateAPIReference(configs); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating driver reference: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Generated documentation for %d configuration structures\n", len(configs))
	fmt.Println("📄 Output: docs/api-reference.md")
}

// parseFilesystems は filesystems ディレクトリ内の Go ファイルを解析します
func parseFilesystems(dir string) ([]ConfigStructInfo, error) {
	var configs []ConfigStructInfo

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Go ファイルのみ処理
		if !info.IsDir() && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			fileConfigs, err := parseGoFile(path)
			if err != nil {
				return fmt.Errorf("failed to parse %s: %w", path, err)
			}
			configs = append(configs, fileConfigs...)
		}

		return nil
	})

	return configs, err
}

// parseGoFile は Go ファイルを解析して設定構造体を抽出します
func parseGoFile(filename string) ([]ConfigStructInfo, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var configs []ConfigStructInfo

	// 構造体定義を探す
	ast.Inspect(node, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true
		}

		// Config で終わる構造体のみ処理
		if !strings.HasSuffix(typeSpec.Name.Name, "Config") {
			return true
		}

		config := ConfigStructInfo{
			Name:        typeSpec.Name.Name,
			PackageName: node.Name.Name,
			FileName:    filename,
		}

		// 構造体のコメントを取得
		if typeSpec.Doc != nil {
			config.Comment = typeSpec.Doc.Text()
		}

		// フィールドを解析
		for _, field := range structType.Fields.List {
			if len(field.Names) == 0 {
				continue // 埋め込みフィールドはスキップ
			}

			fieldName := field.Names[0].Name

			// 非公開フィールドはスキップ
			if !ast.IsExported(fieldName) {
				continue
			}

			fieldInfo := FieldInfo{
				Name: fieldName,
				Type: getTypeName(field.Type),
			}

			// コメントを取得
			if field.Doc != nil {
				fieldInfo.Comment = strings.TrimSpace(field.Doc.Text())
			}

			// タグを解析
			if field.Tag != nil {
				tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))

				fieldInfo.YAMLName = tag.Get("yaml")
				fieldInfo.Doc = tag.Get("doc")
				fieldInfo.Required = tag.Get("required") == "true"
				fieldInfo.DefaultValue = tag.Get("default")
				fieldInfo.Example = tag.Get("example")
			}

			config.Fields = append(config.Fields, fieldInfo)
		}

		if len(config.Fields) > 0 {
			configs = append(configs, config)
		}

		return true
	})

	return configs, nil
}

// getTypeName は型名を文字列として取得します
func getTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", getTypeName(t.X), t.Sel.Name)
	case *ast.StarExpr:
		return "*" + getTypeName(t.X)
	case *ast.ArrayType:
		return "[]" + getTypeName(t.Elt)
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", getTypeName(t.Key), getTypeName(t.Value))
	default:
		return "unknown"
	}
}

// generateAPIReference はドライバーリファレンスを生成します
func generateAPIReference(configs []ConfigStructInfo) error {
	var sb strings.Builder

	sb.WriteString("# ドライバー\n\n")
	sb.WriteString("このドキュメントは自動生成されています。手動で編集しないでください。\n\n")
	sb.WriteString("生成コマンド: `go run scripts/gen-docs.go`\n\n")

	for _, config := range configs {
		sb.WriteString(fmt.Sprintf("## %s\n\n", config.Name))

		if config.Comment != "" {
			sb.WriteString(config.Comment)
			sb.WriteString("\n")
		}

		sb.WriteString(fmt.Sprintf("ファイル: [%s](%s)\n\n", filepath.Base(config.FileName), config.FileName))

		// パラメータテーブル（必須/オプションを一つのテーブルに）
		if len(config.Fields) > 0 {
			sb.WriteString("| パラメータ | 型 | 必須 | デフォルト | 説明 |\n")
			sb.WriteString("|-----------|-----|:----:|----------|------|\n")

			for _, field := range config.Fields {
				doc := field.Doc
				if doc == "" {
					doc = field.Comment
				}
				doc = strings.ReplaceAll(doc, "\n", " ")

				required := ""
				if field.Required {
					required = "✓"
				}

				defaultValue := field.DefaultValue
				if defaultValue == "" {
					defaultValue = "-"
				}

				sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
					field.YAMLName,
					field.Type,
					required,
					defaultValue,
					doc,
				))
			}
			sb.WriteString("\n")
		}

		// 設定例を生成
		if hasExamples(config.Fields) {
			sb.WriteString("**設定例:**\n\n")
			sb.WriteString("```yaml\n")
			sb.WriteString(fmt.Sprintf("%s:\n", strings.ToLower(strings.TrimSuffix(config.Name, "Config"))))
			sb.WriteString("  driver: " + inferDriverName(config.Name) + "\n")
			sb.WriteString("  args:\n")

			for _, field := range config.Fields {
				if field.Example != "" {
					sb.WriteString(fmt.Sprintf("    %s: %s\n", field.YAMLName, field.Example))
				}
			}

			sb.WriteString("```\n\n")
		}

		sb.WriteString("---\n\n")
	}

	// ファイルに書き込み
	outputPath := "docs/api-reference.md"
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}

	return os.WriteFile(outputPath, []byte(sb.String()), 0644)
}

// hasExamples はフィールドに example タグがあるかチェックします
func hasExamples(fields []FieldInfo) bool {
	for _, field := range fields {
		if field.Example != "" {
			return true
		}
	}
	return false
}

// inferDriverName は構造体名からドライバー名を推測します
func inferDriverName(structName string) string {
	name := strings.TrimSuffix(structName, "Config")
	name = strings.TrimSuffix(name, "Fs")
	return strings.ToLower(name)
}
