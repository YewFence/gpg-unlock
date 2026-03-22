package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/YewFence/gpg-unlock/internal/config"
)

func runEdit() {
	dir := config.Dir()
	if dir == "" {
		fmt.Fprintln(os.Stderr, "错误: 无法确定配置目录（HOME 未设置）")
		os.Exit(1)
	}

	cfgPath := filepath.Join(dir, "config.toml")

	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "错误: 配置文件不存在: %s\n", cfgPath)
		fmt.Fprintln(os.Stderr, "请先运行 gpg-unlock init 创建配置")
		os.Exit(1)
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		if _, err := exec.LookPath("vi"); err == nil {
			editor = "vi"
		} else if _, err := exec.LookPath("notepad"); err == nil {
			editor = "notepad"
		} else {
			fmt.Fprintln(os.Stderr, "错误: 未设置 EDITOR 环境变量，且未找到 vi 或 notepad")
			os.Exit(1)
		}
	}

	parts := strings.Fields(editor)
	if len(parts) == 0 {
		fmt.Fprintln(os.Stderr, "错误: EDITOR 环境变量为空")
		os.Exit(1)
	}

	args := append(parts[1:], cfgPath)
	c := exec.Command(parts[0], args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 启动编辑器失败: %v\n", err)
		os.Exit(1)
	}
}
