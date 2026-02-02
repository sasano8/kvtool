package sources

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSourceEnvImplementsInterface(t *testing.T) {
	// Source インターフェースを実装していることを確認
	var _ Source = (*SourceEnv)(nil)
}

func TestSourceEnvLoad(t *testing.T) {
	require := require.New(t)

	// テスト用の環境変数を設定
	os.Setenv("TEST_SOURCE_KEY", "test_value")
	defer os.Unsetenv("TEST_SOURCE_KEY")

	// Source インターフェースとして使用
	var source Source = &SourceEnv{}
	reader, err := source.Load()
	require.NoError(err)
	require.NotNil(reader)

	// データを読み取る
	data, err := io.ReadAll(reader)
	require.NoError(err)

	// TEST_SOURCE_KEY が含まれていることを確認
	content := string(data)
	require.Contains(content, "TEST_SOURCE_KEY=test_value")
}

func TestSourceEnvFormat(t *testing.T) {
	require := require.New(t)

	os.Setenv("KEY1", "value1")
	os.Setenv("KEY2", "value2")
	defer os.Unsetenv("KEY1")
	defer os.Unsetenv("KEY2")

	source := &SourceEnv{}
	reader, err := source.Load()
	require.NoError(err)

	data, err := io.ReadAll(reader)
	require.NoError(err)

	// フォーマットが KEY=VALUE\n であることを確認
	content := string(data)
	lines := strings.Split(content, "\n")

	foundKey1 := false
	foundKey2 := false
	for _, line := range lines {
		if line == "KEY1=value1" {
			foundKey1 = true
		}
		if line == "KEY2=value2" {
			foundKey2 = true
		}
	}

	require.True(foundKey1, "KEY1=value1 should be in output")
	require.True(foundKey2, "KEY2=value2 should be in output")
}
