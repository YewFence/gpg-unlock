package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/YewFence/gpg-unlock/backend"
	"github.com/YewFence/gpg-unlock/internal/config"
	"github.com/YewFence/gpg-unlock/internal/gpg"
)

// Version 由 ldflags 注入
var Version = "dev"

// Execute 是 CLI 入口，解析子命令和 flag
func Execute() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "init":
			runInit()
			return
		case "reset":
			runReset()
			return
		case "edit":
			runEdit()
			return
		case "gen-example":
			runGenExample()
			return
		case "version":
			fmt.Println("gpg-unlock", Version)
			return
		}
	}

	configPath := flag.String("config", "", "配置文件路径")
	backendName := flag.String("backend", "", "指定后端（覆盖配置文件）")
	flag.StringVar(backendName, "b", "", "指定后端（简写）")
	showVersion := flag.Bool("version", false, "显示版本")
	force := flag.Bool("force", false, "强制重新注入密码短语（即使已缓存）")
	flag.Parse()

	if *showVersion {
		fmt.Println("gpg-unlock", Version)
		return
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	if *backendName != "" {
		cfg.Backend = *backendName
	}

	b, err := backend.Get(cfg.Backend)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	params := cfg.Backends[cfg.Backend]
	if params == nil {
		fmt.Fprintf(os.Stderr, "错误: 配置中缺少 [backends.%s] 段\n", cfg.Backend)
		os.Exit(1)
	}

	params = b.FormatConfig(params)
	if errs := b.ValidateConfig(params); len(errs) > 0 {
		fmt.Fprintln(os.Stderr, "配置校验失败:")
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "  - %v\n", e)
		}
		os.Exit(1)
	}

	fmt.Println("=== GPG 密码短语加载器 ===")
	fmt.Printf("后端: %s\n\n", cfg.Backend)

	fmt.Println("正在获取 GPG 密钥信息...")
	keygrips, err := gpg.GetKeygrips()
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("发现 %d 个密钥 Keygrip\n\n", len(keygrips))

	if !*force {
		fmt.Println("正在检查缓存状态...")
		cached, total := gpg.CheckAllKeygripsCached(keygrips)

		if cached == total {
			fmt.Printf("所有密钥（%d/%d）的密码短语已缓存\n", cached, total)
			fmt.Println("无需重新注入（使用 --force 可强制重新注入）")
			fmt.Println()
			fmt.Println("完成！现在可以无感签名了")
			return
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
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("已获取密码短语")
	fmt.Println()

	fmt.Println("正在缓存密码短语到 gpg-agent...")
	if err := gpg.PresetPassphrase(keygrips, passphrase); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("完成！现在可以无感签名了")
}
