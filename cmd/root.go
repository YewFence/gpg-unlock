package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/YewFence/gpg-unlock/backend"
	"github.com/YewFence/gpg-unlock/internal/config"
	"github.com/YewFence/gpg-unlock/internal/gpg"
)

// Version 由 ldflags 注入
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:          "gpg-unlock",
	Short:        "从密码管理器获取 GPG 密码短语并注入 gpg-agent 缓存",
	SilenceUsage: true,
	RunE:         runUnlock,
}

func init() {
	names := backend.Names()
	sort.Strings(names)
	var sb strings.Builder
	sb.WriteString("从密码管理器获取 GPG 密码短语并注入 gpg-agent 缓存，实现 Git 签名免密。\n\n")
	sb.WriteString("可用后端:\n")
	for _, name := range names {
		fmt.Fprintf(&sb, "  - %s\n", name)
	}
	rootCmd.Long = sb.String()

	rootCmd.Version = Version
	rootCmd.SetVersionTemplate("gpg-unlock {{.Version}}\n")

	rootCmd.PersistentFlags().String("config", "", "配置文件路径")
	rootCmd.PersistentFlags().StringP("backend", "b", "", "指定后端（覆盖配置文件）")
	rootCmd.Flags().Bool("force", false, "强制重新注入密码短语（即使已缓存）")
}

// Execute 是 CLI 入口
func Execute() {
	rootCmd.SilenceErrors = true
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}

func runUnlock(cmd *cobra.Command, args []string) error {
	configPath, _ := cmd.Flags().GetString("config")
	backendName, _ := cmd.Flags().GetString("backend")
	force, _ := cmd.Flags().GetBool("force")

	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	if backendName != "" {
		cfg.Backend = backendName
	}

	b, err := backend.Get(cfg.Backend)
	if err != nil {
		return err
	}

	params := cfg.Backends[cfg.Backend]
	if params == nil {
		return fmt.Errorf("配置中缺少 [backends.%s] 段", cfg.Backend)
	}

	params = b.FormatConfig(params)
	if errs := b.ValidateConfig(params); len(errs) > 0 {
		fmt.Fprintln(os.Stderr, "配置校验失败:")
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "  - %v\n", e)
		}
		return fmt.Errorf("配置校验失败")
	}

	fmt.Println("=== GPG 密码短语加载器 ===")
	fmt.Printf("后端: %s\n\n", cfg.Backend)

	fmt.Println("正在获取 GPG 密钥信息...")
	keygrips, err := gpg.GetKeygrips()
	if err != nil {
		return err
	}
	fmt.Printf("发现 %d 个密钥 Keygrip\n\n", len(keygrips))

	if !force {
		fmt.Println("正在检查缓存状态...")
		cached, total := gpg.CheckAllKeygripsCached(keygrips)

		if cached == total {
			fmt.Printf("所有密钥（%d/%d）的密码短语已缓存\n", cached, total)
			fmt.Println("无需重新注入（使用 --force 可强制重新注入）")
			fmt.Println()
			fmt.Println("完成！现在可以无感签名了")
			return nil
		}

		if cached > 0 {
			fmt.Printf("部分密钥（%d/%d）已缓存，将注入剩余密钥\n\n", cached, total)
		} else {
			fmt.Println("密码短语未缓存，需要注入")
		}
	}

	fmt.Println()
	fmt.Println("正在获取密码短语...")
	passphrase, err := b.GetPassphrase(params)
	if err != nil {
		return err
	}
	fmt.Println("已获取密码短语")
	fmt.Println()

	fmt.Println("正在缓存密码短语到 gpg-agent...")
	if err := gpg.PresetPassphrase(keygrips, passphrase); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("完成！现在可以无感签名了")
	return nil
}
