package cmd_file_load

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/sasano8/kvtool/pkg/filesystems"
	"github.com/sasano8/kvtool/pkg/encoders"
	"github.com/spf13/cobra"
)

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func catchPanic[T any](fn func() (T, error)) (v T, err error) {
	defer func() {
		if r := recover(); r != nil {
			// そのまま err に格納
			err = fmt.Errorf("panic: %v\n%s", r, debug.Stack())
		}
	}()
	return fn()
}

func init() {
	cmd := CmdLoadFromVault
	cmd.Flags().StringP("addr", "a", "", "adress")
	cmd.Flags().StringP("namespace", "n", "admin", "Vault namespace (default: admin) (Enterprise)")
	cmd.Flags().StringP("mount", "m", "", "KV mount path (e.g. secret)")
	cmd.Flags().IntP("kv", "k", 2, "KV engine version: 1 or 2")
	cmd.Flags().IntP("version", "v", 0, "KV v2 version (0=latest)")
	cmd.Flags().StringP("field", "f", "", "output only this key from secret (optional)")
	cmd.Flags().StringP("token", "t", "", "output only this key from secret (optional)")
	cmd.Flags().BoolP("pretty", "p", true, "pretty print JSON")
	cmd.Flags().DurationP("timeout", "b", 10*time.Second, "request timeout")
	cmd.Flags().StringP("output", "o", "@", "")
	cmd.MarkFlagRequired("addr")
	cmd.MarkFlagRequired("mount")
	cmd.MarkFlagRequired("token")
}

var CmdLoadFromVault = &cobra.Command{
	Use:   "vault",
	Short: "A brief description of your command",
	Long:  "Env: VAULT_ADDR, VAULT_TOKEN, VAULT_NAMESPACE, VAULT_* TLS vars are supported by vault/api config.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path_out, _ := cmd.Flags().GetString("output")
		path_file := args[0]

		out := filesystems.FsLocalFileCli{Path: path_out, NoMkdir: false}
		w, err := out.OpenWriter()
		if err != nil {
			return err
		}
		defer w.Close()

		config, err := catchPanic(func() (filesystems.VaultConfig, error) {
			config := filesystems.VaultConfig{
				Addr:      must(cmd.Flags().GetString("addr")),
				Namespace: must(cmd.Flags().GetString("namespace")),
				Mount:     must(cmd.Flags().GetString("mount")),
				KvVer:     must(cmd.Flags().GetInt("kv")),
				Version:   must(cmd.Flags().GetInt("version")),
				Field:     must(cmd.Flags().GetString("field")),
				Pretty:    must(cmd.Flags().GetBool("pretty")),
				Timeout:   must(cmd.Flags().GetDuration("timeout")),
				Token:     must(cmd.Flags().GetString("token")),
			}
			return config, nil
		})
		if err != nil {
			return err
		}

		loadAsBytes := func(ctx context.Context, config filesystems.VaultConfig, path string) ([]byte, error) {
			parent := context.Background()
			fs, err := filesystems.GetVaultFs(parent, &config)
			if err != nil {
				return nil, err
			}

			file, err := fs.GetFile(path_file)
			if err != nil {
				return nil, err
			}

			result, err := file.LoadAsJson()
			if err != nil {
				return nil, err
			}

			encoder := encoders.NewJsonEncoder()
			bytes, err := encoder.Marshal(result)

			return bytes, nil
		}

		parent := context.Background()
		bytes, err := loadAsBytes(parent, config, path_file)
		if err != nil {
			return err
		}

		_, err = w.Write(bytes)
		if err != nil {
			return err
		}
		return nil
	},
}
