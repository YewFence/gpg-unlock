package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/YewFence/gpg-unlock/backend"
)

var version = "dev"

func main() {
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
		case "version":
			fmt.Println("gpg-unlock", version)
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
		fmt.Println("gpg-unlock", version)
		return
	}

	cfg, err := loadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	// 如果指定了 --backend/-b，覆盖配置文件中的后端
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

	fmt.Println("=== GPG 密码短语加载器 ===")
	fmt.Printf("后端: %s\n\n", cfg.Backend)

	fmt.Println("正在获取 GPG 密钥信息...")
	keygrips, err := getKeygrips()
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("发现 %d 个密钥 Keygrip\n\n", len(keygrips))

	// 缓存检测逻辑
	if !*force {
		fmt.Println("正在检查缓存状态...")
		cached, total := checkAllKeygripsCached(keygrips)

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
	if err := presetPassphrase(keygrips, passphrase); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("完成！现在可以无感签名了")
}

func runInit() {
	dir := configDir()
	if dir == "" {
		fmt.Fprintln(os.Stderr, "错误: 无法确定配置目录（HOME 未设置）")
		os.Exit(1)
	}

	cfgPath := filepath.Join(dir, "config.toml")

	if _, err := os.Stat(cfgPath); err == nil {
		fmt.Printf("配置文件已存在: %s\n", cfgPath)
		fmt.Print("要覆盖吗？[y/N] ")
		if !confirmYes() {
			fmt.Println("已取消")
			return
		}
	}

	reader := bufio.NewReader(os.Stdin)

	names := backend.Names()
	fmt.Println("=== gpg-unlock 配置向导 ===")
	fmt.Println()
	fmt.Println("可用后端:")
	for i, name := range names {
		fmt.Printf("  %d. %s\n", i+1, name)
	}
	fmt.Printf("选择后端 [1]: ")
	input := readLine(reader)
	idx := 0
	if input != "" {
		n, err := fmt.Sscanf(input, "%d", &idx)
		if n != 1 || err != nil || idx < 1 || idx > len(names) {
			fmt.Fprintf(os.Stderr, "错误: 无效选择 %q\n", input)
			os.Exit(1)
		}
		idx--
	}
	chosen := names[idx]
	b, err := backend.Get(chosen)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	params := map[string]string{}
	for _, f := range b.ConfigFields() {
		fmt.Printf("%s: ", f.Prompt)
		val := readLine(reader)
		if f.Required && val == "" {
			fmt.Fprintf(os.Stderr, "错误: %s 不能为空\n", f.Key)
			os.Exit(1)
		}
		params[f.Key] = val
	}

	var buf strings.Builder
	fmt.Fprintf(&buf, "backend = %q\n\n[backends.%s]\n", chosen, chosen)
	for _, f := range b.ConfigFields() {
		fmt.Fprintf(&buf, "%s = %q\n", f.Key, params[f.Key])
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 创建目录失败: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(cfgPath, []byte(buf.String()), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 写入配置失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("配置已写入: %s\n", cfgPath)
	fmt.Println("现在可以运行 gpg-unlock 来加载密码短语了")
}

func readLine(r *bufio.Reader) string {
	line, _ := r.ReadString('\n')
	return strings.TrimSpace(line)
}

func confirmYes() bool {
	reader := bufio.NewReader(os.Stdin)
	line := readLine(reader)
	return strings.EqualFold(line, "y") || strings.EqualFold(line, "yes")
}

func runReset() {
	dir := configDir()
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

func runEdit() {
	dir := configDir()
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

	// 解析编辑器命令（可能包含参数，如 "zed --wait"）
	parts := strings.Fields(editor)
	if len(parts) == 0 {
		fmt.Fprintln(os.Stderr, "错误: EDITOR 环境变量为空")
		os.Exit(1)
	}

	// 将配置文件路径追加到参数列表
	args := append(parts[1:], cfgPath)
	cmd := exec.Command(parts[0], args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 启动编辑器失败: %v\n", err)
		os.Exit(1)
	}
}
