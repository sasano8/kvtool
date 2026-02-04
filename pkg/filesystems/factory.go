package filesystems

import (
	"context"
	"fmt"
	"time"
)

// StoreInfo represents store configuration
// This is a simplified version to avoid circular dependency with internal/config
type StoreInfo struct {
	Driver string
	Args   map[string]interface{}
}

// FilesystemFactory creates Filesystem instances from configuration
type FilesystemFactory struct {
	ctx context.Context
}

// NewFilesystemFactory creates a new factory instance
func NewFilesystemFactory(ctx context.Context) *FilesystemFactory {
	return &FilesystemFactory{ctx: ctx}
}

// Create creates a Filesystem based on the store configuration
func (f *FilesystemFactory) Create(storeInfo *StoreInfo) (Filesystem, error) {
	switch storeInfo.Driver {
	case "local":
		return f.createLocalFs(storeInfo)
	case "vault":
		return f.createVaultFs(storeInfo)
	case "env":
		return f.createEnvFs(storeInfo)
	case "s3":
		return f.createS3Fs(storeInfo)
	case "db":
		return f.createDbFs(storeInfo)
	default:
		return nil, fmt.Errorf("unsupported driver: %s", storeInfo.Driver)
	}
}

func (f *FilesystemFactory) createLocalFs(storeInfo *StoreInfo) (Filesystem, error) {
	args := storeInfo.Args
	root, _ := args["root"].(string)
	if root == "" {
		root = "."
	}

	timeout := 10 * time.Second
	if timeoutVal, ok := args["timeout"].(int); ok {
		timeout = time.Duration(timeoutVal) * time.Second
	}

	// Transform 設定を取得
	transform := getTransformFromArgs(args, "read")

	return GetLocalFs(f.ctx, &LocalFsConfig{
		Root:      root,
		Timeout:   timeout,
		Transform: transform,
	})
}

func (f *FilesystemFactory) createVaultFs(storeInfo *StoreInfo) (Filesystem, error) {
	args := storeInfo.Args

	// Parse vault connection args
	connArgs, ok := args["conn"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("vault store missing 'conn' configuration")
	}

	addr, _ := connArgs["addr"].(string)
	token, _ := connArgs["token"].(string)
	mount, _ := connArgs["mount"].(string)

	if addr == "" || token == "" || mount == "" {
		return nil, fmt.Errorf("vault store missing required configuration (addr, token, mount)")
	}

	namespace, _ := connArgs["namespace"].(string)
	if namespace == "" {
		namespace = "admin"
	}

	timeout := 10 * time.Second
	if timeoutVal, ok := args["timeout"].(int); ok {
		timeout = time.Duration(timeoutVal) * time.Second
	}

	config := VaultConfig{
		Addr:      addr,
		Token:     token,
		Namespace: namespace,
		Mount:     mount,
		KvVer:     2,
		Version:   0,
		Timeout:   timeout,
	}

	return GetVaultFs(f.ctx, &config)
}

func (f *FilesystemFactory) createEnvFs(storeInfo *StoreInfo) (Filesystem, error) {
	// Environment filesystem doesn't need any configuration
	// Just return a new instance with context
	return &FsEnvFilesystem{
		Ctx: f.ctx,
	}, nil
}

func (f *FilesystemFactory) createS3Fs(storeInfo *StoreInfo) (Filesystem, error) {
	args := storeInfo.Args

	// Parse S3 configuration
	bucket, _ := args["bucket"].(string)
	region, _ := args["region"].(string)
	root, _ := args["root"].(string)
	endpoint, _ := args["endpoint"].(string)
	accessKeyID, _ := args["access_key_id"].(string)
	secretAccessKey, _ := args["secret_access_key"].(string)
	sessionToken, _ := args["session_token"].(string)

	// Parse boolean with type assertion
	usePathStyle := false
	if val, ok := args["use_path_style"].(bool); ok {
		usePathStyle = val
	}

	// Parse timeout
	timeout := 30 * time.Second
	if timeoutVal, ok := args["timeout"].(int); ok {
		timeout = time.Duration(timeoutVal) * time.Second
	}

	// Transform 設定を取得
	transform := getTransformFromArgs(args, "read")
	if transform == "" {
		if val, ok := args["transform"].(string); ok {
			transform = val
		}
	}

	config := S3FsConfig{
		Bucket:          bucket,
		Region:          region,
		Root:            root,
		Endpoint:        endpoint,
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		SessionToken:    sessionToken,
		UsePathStyle:    usePathStyle,
		Transform:       transform,
		Timeout:         timeout,
	}

	return NewS3Fs(f.ctx, config)
}

func (f *FilesystemFactory) createDbFs(storeInfo *StoreInfo) (Filesystem, error) {
	args := storeInfo.Args

	connectionString, _ := args["connection_string"].(string)
	driver, _ := args["driver"].(string)
	query, _ := args["query"].(string)
	namespace, _ := args["namespace"].(string)

	timeout := 10 * time.Second
	if timeoutVal, ok := args["timeout"].(int); ok {
		timeout = time.Duration(timeoutVal) * time.Second
	}

	config := DbFsConfig{
		ConnectionString: connectionString,
		Driver:           driver,
		Query:            query,
		Namespace:        namespace,
		Timeout:          timeout,
	}

	return NewDbFs(f.ctx, &config)
}

// GetFilesystem is a convenience function that creates a filesystem without storing the factory
func GetFilesystem(ctx context.Context, storeInfo *StoreInfo) (Filesystem, error) {
	factory := NewFilesystemFactory(ctx)
	return factory.Create(storeInfo)
}

// getTransformFromArgs は args から transform.{direction} の値を取得する
func getTransformFromArgs(args map[string]interface{}, direction string) string {
	transformMap, ok := args["transform"].(map[string]interface{})
	if !ok {
		return ""
	}
	val, _ := transformMap[direction].(string)
	return val
}
