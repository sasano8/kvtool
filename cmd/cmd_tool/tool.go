package cmd_tool

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// CmdTool は tool コマンドのルート
var CmdTool = &cobra.Command{
	Use:   "tool",
	Short: "Utility tools for generating values",
	Long: `Utility tools for generating UUIDs, timestamps, random values, and passwords.

Examples:
  kvtool tool uuid7
  kvtool tool uuid4
  kvtool tool now
  kvtool tool timestamp
  kvtool tool random/hex/16
  kvtool tool random/base64/32
  kvtool tool password`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTool(args[0])
	},
}

func runTool(toolPath string) error {
	parts := strings.Split(toolPath, "/")
	toolName := parts[0]

	switch toolName {
	case "uuid7":
		return generateUUID7()
	case "uuid4":
		return generateUUID4()
	case "now":
		return generateNow()
	case "timestamp":
		return generateTimestamp()
	case "random":
		if len(parts) < 3 {
			return fmt.Errorf("usage: random/{hex|base64}/{bytes}")
		}
		format := parts[1]
		bytes, err := strconv.Atoi(parts[2])
		if err != nil {
			return fmt.Errorf("invalid byte count: %s", parts[2])
		}
		return generateRandom(format, bytes)
	case "password":
		return generatePassword()
	default:
		return fmt.Errorf("unknown tool: %s", toolName)
	}
}

func generateUUID7() error {
	id, err := uuid.NewV7()
	if err != nil {
		return err
	}
	fmt.Println(id.String())
	return nil
}

func generateUUID4() error {
	id := uuid.New()
	fmt.Println(id.String())
	return nil
}

func generateNow() error {
	fmt.Println(time.Now().UTC().Format(time.RFC3339))
	return nil
}

func generateTimestamp() error {
	fmt.Println(time.Now().Unix())
	return nil
}

func generateRandom(format string, bytes int) error {
	if bytes <= 0 || bytes > 1024 {
		return fmt.Errorf("bytes must be between 1 and 1024")
	}

	data := make([]byte, bytes)
	if _, err := rand.Read(data); err != nil {
		return err
	}

	switch format {
	case "hex":
		fmt.Println(hex.EncodeToString(data))
	case "base64":
		fmt.Println(base64.StdEncoding.EncodeToString(data))
	default:
		return fmt.Errorf("unknown format: %s (use hex or base64)", format)
	}
	return nil
}

func generatePassword() error {
	// 16文字のパスワードを生成
	const length = 16
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"

	password := make([]byte, length)
	randomBytes := make([]byte, length)
	if _, err := rand.Read(randomBytes); err != nil {
		return err
	}

	for i := 0; i < length; i++ {
		password[i] = charset[int(randomBytes[i])%len(charset)]
	}

	fmt.Println(string(password))
	return nil
}
