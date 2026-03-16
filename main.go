package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/YewFence/gpg-unlock/backend"
)

var version = "dev"

func main() {
	configPath := flag.String("config", "", "配置文件路径")
	showVersion := flag.Bool("version", false, "显示版本")
	flag.Parse()

	if *showVersion {
		fmt.Println("gpg-unlock", version)
		return
	}

	cfg, err := loadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
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

	fmt.Println("=== GPG 密码短语加载器 ===")
	fmt.Printf("后端: %s\n\n", cfg.Backend)

	fmt.Println("正在获取密码短语...")
	passphrase, err := b.GetPassphrase(params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("已获取密码短语")
	fmt.Println()

	fmt.Println("正在获取 GPG 密钥信息...")
	keygrip, err := getKeygrip()
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("密钥 Keygrip: %s\n\n", keygrip)

	fmt.Println("正在缓存密码短语到 gpg-agent...")
	if err := presetPassphrase(keygrip, passphrase); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("完成！现在可以无感签名了")
}
