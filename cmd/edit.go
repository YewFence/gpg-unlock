package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/shlex"
	"github.com/spf13/cobra"

	"github.com/YewFence/gpg-unlock/internal/config"
)

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "用编辑器打开配置文件",
	Long:  "使用 $EDITOR 环境变量指定的编辑器打开配置文件进行编辑。",
	Args:  cobra.NoArgs,
	RunE:  runEdit,
}

func init() {
	rootCmd.AddCommand(editCmd)
}

func runEdit(cmd *cobra.Command, args []string) error {
	dir := config.Dir()
	if dir == "" {
		return fmt.Errorf("无法确定配置目录（HOME 未设置）")
	}

	cfgPath := filepath.Join(dir, "config.toml")

	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		return fmt.Errorf("配置文件不存在: %s\n请先运行 gpg-unlock init 创建配置", cfgPath)
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		if _, err := exec.LookPath("vi"); err == nil {
			editor = "vi"
		} else if _, err := exec.LookPath("notepad"); err == nil {
			editor = "notepad"
		} else {
			return fmt.Errorf("未设置 EDITOR 环境变量，且未找到 vi 或 notepad")
		}
	}

	parts, err := shlex.Split(editor)
	if err != nil || len(parts) == 0 {
		return fmt.Errorf("EDITOR 解析失败: %w", err)
	}

	editorArgs := append(parts[1:], cfgPath)
	c := exec.Command(parts[0], editorArgs...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		return fmt.Errorf("启动编辑器失败: %w", err)
	}
	return nil
}
