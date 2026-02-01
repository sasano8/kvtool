/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"os"

	"github.com/sasano8/kvtool/cmd/cmd_convert"
	"github.com/sasano8/kvtool/cmd/cmd_load"
	"github.com/sasano8/kvtool/cmd/cmd_store"
	"github.com/sasano8/kvtool/exceptions"
	"github.com/spf13/cobra"
)

// init は Go言語の仕様で、パッケージインポート時に各モジュールの init が自動で呼び出される
func init() {
	belongsTo := func(root *cobra.Command, subs map[*cobra.Command]map[string]any) *cobra.Command {
		// 順序は保障されない
		for k, v := range subs {
			impl := true
			if v, ok := v["implemented"].(bool); ok {
				impl = v
			}

			if !impl {
				k.Short = "[Not Implemented]" + k.Short
				k.Long = "[Not Implemented]" + k.Long
			}
			root.AddCommand(k)
		}
		return root
	}

	loads := map[*cobra.Command]map[string]any{
		cmd_load.CmdLoadFromAuto:   {"implemented": false},
		cmd_load.CmdLoadFromEnv:    {"implemented": true},
		cmd_load.CmdLoadFromDotenv: {"implemented": false},
		cmd_load.CmdLoadFromJson:   {"implemented": false},
		cmd_load.CmdLoadFromJsonl:  {"implemented": false},
		cmd_load.CmdLoadFromHcl:    {"implemented": false},
		cmd_load.CmdLoadFromVault:  {"implemented": false},
		cmd_load.CmdLoadFromYaml:   {"implemented": false},
		cmd_load.CmdLoadFromToml:   {"implemented": false},
	}

	converts := map[*cobra.Command]map[string]any{
		cmd_convert.CmdConvertToDotenv: {"implemented": false},
		cmd_convert.CmdConvertToYaml:   {"implemented": false},
	}

	stores := map[*cobra.Command]map[string]any{
		cmd_store.CmdInit:  {"implemented": true},
		cmd_store.CmdServe: {"implemented": false},
		cmd_store.CmdGet:   {"implemented": true},
	}

	CmdRoot.AddCommand(belongsTo(CmdLoad, loads))
	CmdRoot.AddCommand(belongsTo(CmdConvert, converts))
	CmdRoot.AddCommand(belongsTo(CmdStore, stores))
}

// CmdRoot represents the base command when called without any subcommands
var CmdRoot = &cobra.Command{
	Use:   "kvtool",
	Short: "A brief description of your application",
	Long:  "",
}

// Execute CLI
func Execute() {
	err := CmdRoot.Execute()
	if err == nil {
		os.Exit(0)
	}

	if errors.Is(err, exceptions.ErrSysNotImplemented) {
	}

	os.Exit(1)
}
