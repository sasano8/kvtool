package filesystems

import (
	"context"
	"io"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

// UUID の形式を検証する正規表現（バージョン 7）
var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

// UUID の一般的な形式を検証する正規表現（任意のバージョン）
var uuidAnyRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func TestUUID7FsImplementsInterface(t *testing.T) {
	var _ Filesystem = (*UUID7Fs)(nil)
	var _ File = (*UUID7File)(nil)
}

func TestUUID7FsLoadAsJson(t *testing.T) {
	require := require.New(t)

	fs, err := NewUUID7Fs(context.Background(), &UUID7FsConfig{})
	require.NoError(err)

	file, err := fs.GetFile("any/path/is/ignored")
	require.NoError(err)

	result, err := file.LoadAsJson()
	require.NoError(err)

	// 結果は文字列であること
	uuid, ok := result.(string)
	require.True(ok, "result should be a string")

	// UUID7 の形式であること（バージョン 7）
	require.True(uuidRegex.MatchString(uuid), "should be a valid UUID7 format: %s", uuid)
}

func TestUUID7FsOpenReader(t *testing.T) {
	require := require.New(t)

	fs, err := NewUUID7Fs(context.Background(), &UUID7FsConfig{})
	require.NoError(err)

	file, err := fs.GetFile("")
	require.NoError(err)

	reader, err := file.OpenReader()
	require.NoError(err)
	defer reader.Close()

	data, err := io.ReadAll(reader)
	require.NoError(err)

	uuid := string(data)
	require.True(uuidRegex.MatchString(uuid), "should be a valid UUID7 format: %s", uuid)
}

func TestUUID7FsGeneratesUniqueValues(t *testing.T) {
	require := require.New(t)

	fs, err := NewUUID7Fs(context.Background(), &UUID7FsConfig{})
	require.NoError(err)

	// 複数回呼び出して全て異なることを確認
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		file, err := fs.GetFile("")
		require.NoError(err)

		result, err := file.LoadAsJson()
		require.NoError(err)

		uuid := result.(string)
		require.False(seen[uuid], "UUID should be unique, but got duplicate: %s", uuid)
		seen[uuid] = true
	}
}

func TestUUID7FsPathIsIgnored(t *testing.T) {
	require := require.New(t)

	fs, err := NewUUID7Fs(context.Background(), &UUID7FsConfig{})
	require.NoError(err)

	// 異なるパスでも動作することを確認
	paths := []string{
		"",
		"foo",
		"foo/bar",
		"/absolute/path",
		"../relative",
	}

	for _, path := range paths {
		file, err := fs.GetFile(path)
		require.NoError(err, "path: %s", path)

		result, err := file.LoadAsJson()
		require.NoError(err, "path: %s", path)

		uuid := result.(string)
		require.True(uuidRegex.MatchString(uuid), "path %s should generate valid UUID7: %s", path, uuid)
	}
}

func TestUUID7FsWithSeed(t *testing.T) {
	require := require.New(t)

	seed := int64(12345)
	fs1, err := NewUUID7Fs(context.Background(), &UUID7FsConfig{Seed: &seed})
	require.NoError(err)

	fs2, err := NewUUID7Fs(context.Background(), &UUID7FsConfig{Seed: &seed})
	require.NoError(err)

	// 同じシードで同じ UUID が生成されることを確認
	for i := 0; i < 10; i++ {
		file1, _ := fs1.GetFile("")
		file2, _ := fs2.GetFile("")

		result1, err := file1.LoadAsJson()
		require.NoError(err)

		result2, err := file2.LoadAsJson()
		require.NoError(err)

		require.Equal(result1, result2, "same seed should generate same UUIDs")
	}
}

func TestUUID7FsWithSeedGeneratesValidFormat(t *testing.T) {
	require := require.New(t)

	seed := int64(12345)
	fs, err := NewUUID7Fs(context.Background(), &UUID7FsConfig{Seed: &seed})
	require.NoError(err)

	file, _ := fs.GetFile("")
	result, err := file.LoadAsJson()
	require.NoError(err)

	uuid := result.(string)
	// シード指定時も UUID7 形式（バージョン 7、バリアント RFC 4122）であること
	require.True(uuidRegex.MatchString(uuid), "seeded UUID should be valid UUID7 format: %s", uuid)
}

func TestUUID7FsWithSeedGeneratesUniqueSequence(t *testing.T) {
	require := require.New(t)

	seed := int64(12345)
	fs, err := NewUUID7Fs(context.Background(), &UUID7FsConfig{Seed: &seed})
	require.NoError(err)

	// シード指定時でも連続呼び出しは異なる値を生成する
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		file, _ := fs.GetFile("")
		result, _ := file.LoadAsJson()
		uuid := result.(string)

		require.False(seen[uuid], "seeded UUID should be unique in sequence: %s", uuid)
		seen[uuid] = true
	}
}

func TestUUID7FsWithFixed(t *testing.T) {
	require := require.New(t)

	fixedUUID := "0190a5e4-1234-7abc-8def-0123456789ab"
	fs, err := NewUUID7Fs(context.Background(), &UUID7FsConfig{Fixed: fixedUUID})
	require.NoError(err)

	// 複数回呼び出しても同じ値が返る
	for i := 0; i < 10; i++ {
		file, _ := fs.GetFile("")
		result, err := file.LoadAsJson()
		require.NoError(err)

		require.Equal(fixedUUID, result, "fixed UUID should always return the same value")
	}
}

func TestUUID7FsFixedTakesPrecedenceOverSeed(t *testing.T) {
	require := require.New(t)

	seed := int64(12345)
	fixedUUID := "0190a5e4-1234-7abc-8def-0123456789ab"

	// 両方指定した場合、fixed が優先される
	fs, err := NewUUID7Fs(context.Background(), &UUID7FsConfig{
		Seed:  &seed,
		Fixed: fixedUUID,
	})
	require.NoError(err)

	file, _ := fs.GetFile("")
	result, err := file.LoadAsJson()
	require.NoError(err)

	require.Equal(fixedUUID, result, "fixed should take precedence over seed")
}

func TestUUID7FsInvalidFixedFormat(t *testing.T) {
	require := require.New(t)

	// 不正な形式の UUID はエラーになる
	invalidUUIDs := []string{
		"not-a-uuid",
		"0190a5e4-1234-7abc-8def",            // 短すぎる
		"0190a5e4-1234-7abc-8def-0123456789", // 短すぎる
		"0190a5e4123470bc8def0123456789ab",   // ハイフンなし
		"ZZZZZZZZ-ZZZZ-ZZZZ-ZZZZ-ZZZZZZZZZZZZ", // 不正な文字
	}

	for _, invalid := range invalidUUIDs {
		_, err := NewUUID7Fs(context.Background(), &UUID7FsConfig{Fixed: invalid})
		require.Error(err, "invalid UUID should cause error: %s", invalid)
	}
}

func TestUUID7FsFactoryIntegration(t *testing.T) {
	require := require.New(t)

	factory := NewFilesystemFactory(context.Background())

	t.Run("基本", func(t *testing.T) {
		fs, err := factory.Create(&StoreInfo{Driver: "uuid7"})
		require.NoError(err)

		file, _ := fs.GetFile("")
		result, err := file.LoadAsJson()
		require.NoError(err)

		uuid := result.(string)
		require.True(uuidRegex.MatchString(uuid), "should be valid UUID7: %s", uuid)
	})

	t.Run("seed指定", func(t *testing.T) {
		fs1, _ := factory.Create(&StoreInfo{
			Driver: "uuid7",
			Args:   map[string]interface{}{"seed": 12345},
		})
		fs2, _ := factory.Create(&StoreInfo{
			Driver: "uuid7",
			Args:   map[string]interface{}{"seed": 12345},
		})

		file1, _ := fs1.GetFile("")
		file2, _ := fs2.GetFile("")
		result1, _ := file1.LoadAsJson()
		result2, _ := file2.LoadAsJson()

		require.Equal(result1, result2, "same seed should generate same UUID")
	})

	t.Run("fixed指定", func(t *testing.T) {
		fixedUUID := "0190a5e4-1234-7abc-8def-0123456789ab"
		fs, _ := factory.Create(&StoreInfo{
			Driver: "uuid7",
			Args:   map[string]interface{}{"fixed": fixedUUID},
		})

		file, _ := fs.GetFile("")
		result, _ := file.LoadAsJson()

		require.Equal(fixedUUID, result, "should return fixed UUID")
	})
}
