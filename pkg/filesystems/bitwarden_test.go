package filesystems

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewBitwardenFs_MissingAccessToken(t *testing.T) {
	require := require.New(t)

	_, err := NewBitwardenFs(&BitwardenFsConfig{
		OrganizationID: "dummy-org",
	})
	require.Error(err)
	require.Contains(err.Error(), "access_token is required")
}

func TestNewBitwardenFs_MissingOrganizationID(t *testing.T) {
	require := require.New(t)

	_, err := NewBitwardenFs(&BitwardenFsConfig{
		AccessToken: "dummy-token",
	})
	require.Error(err)
	require.Contains(err.Error(), "organization_id is required")
}

func TestTryParseJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected any
	}{
		{
			name:     "JSON オブジェクト",
			input:    `{"key": "value"}`,
			expected: map[string]any{"key": "value"},
		},
		{
			name:     "JSON 配列",
			input:    `[1, 2, 3]`,
			expected: []any{float64(1), float64(2), float64(3)},
		},
		{
			name:     "JSON 文字列",
			input:    `"hello"`,
			expected: "hello",
		},
		{
			name:     "プレーン文字列",
			input:    "not-json-value",
			expected: "not-json-value",
		},
		{
			name:     "空文字列",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			result := tryParseJSON(tt.input)
			require.Equal(tt.expected, result)
		})
	}
}

func TestFactoryCreateBitwardenFs_MissingToken(t *testing.T) {
	require := require.New(t)

	factory := NewFilesystemFactory(context.Background())

	storeInfo := &StoreInfo{
		Driver: "bitwarden",
		Args:   map[string]interface{}{},
	}

	_, err := factory.Create(storeInfo)
	require.Error(err)
	require.Contains(err.Error(), "access_token")
}

func TestFactoryCreateBitwardenFs_MissingOrgID(t *testing.T) {
	require := require.New(t)

	factory := NewFilesystemFactory(context.Background())

	storeInfo := &StoreInfo{
		Driver: "bitwarden",
		Args: map[string]interface{}{
			"access_token": "dummy",
		},
	}

	_, err := factory.Create(storeInfo)
	require.Error(err)
	require.Contains(err.Error(), "organization_id")
}

// --- 統合テスト（アクセストークン + organization_id が必要） ---

func TestBitwardenFs_Integration(t *testing.T) {
	if os.Getenv("SKIP_BITWARDEN_TESTS") != "" {
		t.Skip("SKIP_BITWARDEN_TESTS is set")
	}

	accessToken := os.Getenv("BWS_ACCESS_TOKEN")
	if accessToken == "" {
		t.Skip("BWS_ACCESS_TOKEN not set")
	}

	organizationID := os.Getenv("BWS_ORGANIZATION_ID")
	if organizationID == "" {
		t.Skip("BWS_ORGANIZATION_ID not set")
	}

	require := require.New(t)

	projectID := os.Getenv("BWS_PROJECT_ID")

	fs, err := NewBitwardenFs(&BitwardenFsConfig{
		AccessToken:    accessToken,
		OrganizationID: organizationID,
		ProjectID:      projectID,
	})
	require.NoError(err)
	defer fs.Close()

	// 全シークレットを取得
	file, err := fs.GetFile("")
	require.NoError(err)

	data, err := file.LoadAsJson()
	require.NoError(err)
	require.NotNil(data)

	dataMap, ok := data.(map[string]any)
	require.True(ok, "全シークレットは map で返る")
	require.NotEmpty(dataMap, "少なくとも 1 つのシークレットが存在する")
}
