package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/YewFence/gpg-unlock/internal/config"
)

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "重置配置（删除配置目录）",
	Long:  "删除 gpg-unlock 配置目录及其中所有文件，需要确认。",
	Args:  cobra.NoArgs,
	RunE:  runReset,
}

func init() {
	rootCmd.AddCommand(resetCmd)
}

func runReset(cmd *cobra.Command, args []string) error {
	dir := config.Dir()
	if dir == "" {
		return fmt.Errorf("无法确定配置目录（HOME 未设置）")
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Println("配置目录不存在，无需清理")
		return nil
	}

	fmt.Printf("将删除配置目录: %s\n", dir)
	fmt.Print("确认删除？[y/N] ")
	reader := bufio.NewReader(os.Stdin)
	if !confirmYes(reader) {
		fmt.Println("已取消")
		return nil
	}

	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("删除配置目录失败: %w", err)
	}

	fmt.Println("配置已清除")
	return nil
}
