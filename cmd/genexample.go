package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/YewFence/gpg-unlock/internal/example"
)

var genExampleCmd = &cobra.Command{
	Use:    "gen-example",
	Short:  "生成示例配置文件",
	Long:   "在当前目录下生成 config.example.toml 示例配置文件。",
	Hidden: true,
	Args:   cobra.NoArgs,
	RunE:   runGenExample,
}

func init() {
	rootCmd.AddCommand(genExampleCmd)
}

func runGenExample(cmd *cobra.Command, args []string) error {
	dir := "."
	if err := example.Generate(dir); err != nil {
		return err
	}
	fmt.Printf("示例配置已写入: %s\n", filepath.Join(dir, "config.example.toml"))
	return nil
}
