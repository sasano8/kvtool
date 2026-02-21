package filesystems

import (
	"context"
	"fmt"
	"time"
)

// StoreInfo represents store configuration
// This is a simplified version to avoid circular dependency with internal/config
type StoreInfo struct {
	Driver  string
	Args    map[string]interface{}
	Context *ContextInfo
}

// ContextInfo represents common operational parameters (from config file)
type ContextInfo struct {
	Timeout *int // seconds
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
	case "rest":
		return f.createRestFs(storeInfo)
	case "nats":
		return f.createNatsFs(storeInfo)
	case "redis":
		return f.createRedisFs(storeInfo)
	case "git":
		return f.createGitFs(storeInfo)
	case "tool":
		return f.createToolFs(storeInfo)
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

	// Transform 設定を取得
	transform := getTransformFromArgs(args, "read")

	return GetLocalFs(f.ctx, &LocalFsConfig{
		Root:      root,
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

	config := VaultConfig{
		Addr:      addr,
		Token:     token,
		Namespace: namespace,
		Mount:     mount,
		KvVer:     2,
		Version:   0,
	}

	timeout := resolveTimeout(storeInfo, 10*time.Second)
	ctx, cancel := context.WithTimeout(f.ctx, timeout)
	defer cancel()

	return GetVaultFs(ctx, &config)
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
	endpoint, _ := args["endpoint"].(string)
	accessKeyID, _ := args["access_key_id"].(string)
	secretAccessKey, _ := args["secret_access_key"].(string)
	sessionToken, _ := args["session_token"].(string)

	// Parse boolean with type assertion
	usePathStyle := false
	if val, ok := args["use_path_style"].(bool); ok {
		usePathStyle = val
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
		Endpoint:        endpoint,
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		SessionToken:    sessionToken,
		UsePathStyle:    usePathStyle,
		Transform:       transform,
	}

	timeout := resolveTimeout(storeInfo, 30*time.Second)
	ctx, cancel := context.WithTimeout(f.ctx, timeout)
	defer cancel()

	return NewS3Fs(ctx, config)
}

func (f *FilesystemFactory) createDbFs(storeInfo *StoreInfo) (Filesystem, error) {
	args := storeInfo.Args

	connectionString, _ := args["connection_string"].(string)
	driver, _ := args["driver"].(string)
	query, _ := args["query"].(string)
	namespace, _ := args["namespace"].(string)

	config := DbFsConfig{
		ConnectionString: connectionString,
		Driver:           driver,
		Query:            query,
		Namespace:        namespace,
	}

	timeout := resolveTimeout(storeInfo, 10*time.Second)
	ctx, cancel := context.WithTimeout(f.ctx, timeout)
	defer cancel()

	return NewDbFs(ctx, &config)
}

func (f *FilesystemFactory) createRestFs(storeInfo *StoreInfo) (Filesystem, error) {
	args := storeInfo.Args

	baseURL, _ := args["base_url"].(string)
	authType, _ := args["auth_type"].(string)
	token, _ := args["token"].(string)
	tokenFile, _ := args["token_file"].(string)
	username, _ := args["username"].(string)
	password, _ := args["password"].(string)
	caFile, _ := args["ca_file"].(string)

	insecure := false
	if val, ok := args["insecure"].(bool); ok {
		insecure = val
	}

	config := &RestFsConfig{
		BaseURL:   baseURL,
		AuthType:  authType,
		Token:     token,
		TokenFile: tokenFile,
		Username:  username,
		Password:  password,
		CAFile:    caFile,
		Insecure:  insecure,
	}

	timeout := resolveTimeout(storeInfo, 30*time.Second)
	ctx, cancel := context.WithTimeout(f.ctx, timeout)
	defer cancel()

	return NewRestFs(ctx, config)
}

func (f *FilesystemFactory) createNatsFs(storeInfo *StoreInfo) (Filesystem, error) {
	args := storeInfo.Args

	url, _ := args["url"].(string)
	bucket, _ := args["bucket"].(string)
	token, _ := args["token"].(string)
	user, _ := args["user"].(string)
	password, _ := args["password"].(string)
	credsFile, _ := args["creds_file"].(string)

	config := &NatsFsConfig{
		URL:       url,
		Bucket:    bucket,
		Token:     token,
		User:      user,
		Password:  password,
		CredsFile: credsFile,
	}

	timeout := resolveTimeout(storeInfo, 10*time.Second)
	ctx, cancel := context.WithTimeout(f.ctx, timeout)
	defer cancel()

	return NewNatsFs(ctx, config)
}

func (f *FilesystemFactory) createRedisFs(storeInfo *StoreInfo) (Filesystem, error) {
	args := storeInfo.Args

	addr, _ := args["addr"].(string)
	password, _ := args["password"].(string)
	prefix, _ := args["prefix"].(string)

	db := 0
	if dbVal, ok := args["db"].(int); ok {
		db = dbVal
	}

	config := &RedisFsConfig{
		Addr:     addr,
		Password: password,
		DB:       db,
		Prefix:   prefix,
	}

	timeout := resolveTimeout(storeInfo, 10*time.Second)
	ctx, cancel := context.WithTimeout(f.ctx, timeout)
	defer cancel()

	return NewRedisFs(ctx, config)
}

func (f *FilesystemFactory) createGitFs(storeInfo *StoreInfo) (Filesystem, error) {
	args := storeInfo.Args

	url, _ := args["url"].(string)
	ref, _ := args["ref"].(string)

	config := &GitFsConfig{
		URL: url,
		Ref: ref,
	}

	timeout := resolveTimeout(storeInfo, 60*time.Second)
	ctx, cancel := context.WithTimeout(f.ctx, timeout)
	defer cancel()

	return NewGitFs(ctx, config)
}

func (f *FilesystemFactory) createToolFs(storeInfo *StoreInfo) (Filesystem, error) {
	return NewToolFs(f.ctx, &ToolFsConfig{})
}

// GetFilesystem is a convenience function that creates a filesystem without storing the factory
func GetFilesystem(ctx context.Context, storeInfo *StoreInfo) (Filesystem, error) {
	factory := NewFilesystemFactory(ctx)
	return factory.Create(storeInfo)
}

// resolveTimeout は StoreInfo から timeout を解決する
// context.timeout が指定されていればその値、なければ defaultTimeout
func resolveTimeout(storeInfo *StoreInfo, defaultTimeout time.Duration) time.Duration {
	if storeInfo.Context != nil && storeInfo.Context.Timeout != nil {
		return time.Duration(*storeInfo.Context.Timeout) * time.Second
	}
	return defaultTimeout
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
