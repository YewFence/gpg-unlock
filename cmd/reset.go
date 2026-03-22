package cmd

import (
	"fmt"
	"os"

	"github.com/YewFence/gpg-unlock/internal/config"
)

func runReset() {
	dir := config.Dir()
	if dir == "" {
		fmt.Fprintln(os.Stderr, "错误: 无法确定配置目录（HOME 未设置）")
		os.Exit(1)
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Println("配置目录不存在，无需清理")
		return
	}

	fmt.Printf("将删除配置目录: %s\n", dir)
	fmt.Print("确认删除？[y/N] ")
	if !confirmYes() {
		fmt.Println("已取消")
		return
	}

	if err := os.RemoveAll(dir); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 删除配置目录失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("配置已清除")
}
