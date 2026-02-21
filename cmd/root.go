/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/sasano8/kvtool/cmd/cmd_file_convert"
	"github.com/sasano8/kvtool/cmd/cmd_file_load"
	"github.com/sasano8/kvtool/cmd/cmd_store"
	"github.com/sasano8/kvtool/cmd/cmd_tool"
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

	// file load subcommands
	fileLoads := map[*cobra.Command]map[string]any{
		cmd_file_load.CmdLoadFromAuto:   {"implemented": false},
		cmd_file_load.CmdLoadFromEnv:    {"implemented": true},
		cmd_file_load.CmdLoadFromDotenv: {"implemented": false},
		cmd_file_load.CmdLoadFromJson:   {"implemented": false},
		cmd_file_load.CmdLoadFromJsonl:  {"implemented": false},
		cmd_file_load.CmdLoadFromHcl:    {"implemented": false},
		cmd_file_load.CmdLoadFromVault:  {"implemented": false},
		cmd_file_load.CmdLoadFromYaml:   {"implemented": false},
		cmd_file_load.CmdLoadFromToml:   {"implemented": false},
	}

	// file convert subcommands
	fileConverts := map[*cobra.Command]map[string]any{
		cmd_file_convert.CmdConvertToDotenv: {"implemented": false},
		cmd_file_convert.CmdConvertToYaml:   {"implemented": false},
	}

	// store subcommands
	stores := map[*cobra.Command]map[string]any{
		cmd_store.CmdInit:    {"implemented": true},
		cmd_store.CmdConnect: {"implemented": true},
		cmd_store.CmdServe:   {"implemented": false},
		cmd_store.CmdLoad:    {"implemented": true},
		cmd_store.CmdSync:    {"implemented": true},
	}

	// Build command hierarchy
	// CmdFile contains: load and convert subcommands
	CmdFileLoad := &cobra.Command{
		Use:   "load",
		Short: "Load data from various sources",
		Long:  "Load data from various sources (env, vault, json, yaml, etc.)",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	CmdFileConvert := &cobra.Command{
		Use:   "convert",
		Short: "Convert between formats",
		Long:  "Convert data between different formats (dotenv, yaml, json, etc.)",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	CmdFile.AddCommand(belongsTo(CmdFileLoad, fileLoads))
	CmdFile.AddCommand(belongsTo(CmdFileConvert, fileConverts))
	CmdRoot.AddCommand(CmdFile)
	CmdRoot.AddCommand(belongsTo(CmdStore, stores))
	CmdRoot.AddCommand(cmd_tool.CmdTool)
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

	os.Exit(1)
}
