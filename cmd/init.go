package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/YewFence/gpg-unlock/backend"
	"github.com/YewFence/gpg-unlock/internal/config"
	"github.com/YewFence/gpg-unlock/internal/example"
)

func runInit() {
	dir := config.Dir()
	if dir == "" {
		fmt.Fprintln(os.Stderr, "错误: 无法确定配置目录（HOME 未设置）")
		os.Exit(1)
	}

	cfgPath := filepath.Join(dir, "config.toml")

	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 创建目录失败: %v\n", err)
		os.Exit(1)
	}

	// 尝试加载现有配置
	var existingBackend string
	existingParams := map[string]map[string]string{}
	if existing, err := config.Load(cfgPath); err == nil {
		existingBackend = existing.Backend
		for name, params := range existing.Backends {
			existingParams[name] = params
		}
		fmt.Printf("已加载现有配置: %s\n", cfgPath)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("=== gpg-unlock 配置向导 ===")
	fmt.Println()

	for {
		names := backend.Names()
		sort.Strings(names)
		fmt.Println("可用后端:")
		for i, name := range names {
			mark := ""
			if _, ok := existingParams[name]; ok {
				mark = " ✓"
			}
			fmt.Printf("  %d. %s%s\n", i+1, name, mark)
		}
		fmt.Print("选择后端 [1]: ")
		input := readLine(reader)
		idx := 0
		if input != "" {
			n, err := fmt.Sscanf(input, "%d", &idx)
			if n != 1 || err != nil || idx < 1 || idx > len(names) {
				fmt.Fprintf(os.Stderr, "错误: 无效选择 %q\n", input)
				fmt.Println()
				continue
			}
			idx--
		}
		chosen := names[idx]
		b, err := backend.Get(chosen)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			fmt.Println()
			continue
		}

		fmt.Println()
		params := map[string]string{}
		for _, f := range b.ConfigFields() {
			for {
				fmt.Printf("%s; 示例： %s", f.Prompt, f.Example)
				if f.DefaultValue != "" {
					fmt.Printf("(默认: %s) ", f.DefaultValue)
				}
				if f.Required {
					fmt.Print("（必填） ")
				}
				fmt.Print("：")
				val := readLine(reader)
				if f.Required && val == "" {
					fmt.Fprintf(os.Stderr, "错误: %s 不能为空\n", f.Key)
					continue
				}
				params[f.Key] = val
				break
			}
		}

		existingParams[chosen] = b.FormatConfig(params)
		if existingBackend == "" {
			existingBackend = chosen
		}

		fmt.Println()
		fmt.Print("继续配置另一个后端？[y/N] ")
		if !confirmYes() {
			break
		}
		fmt.Println()
	}

	// 生成 config.toml
	allNames := make([]string, 0, len(existingParams))
	for name := range existingParams {
		allNames = append(allNames, name)
	}
	sort.Strings(allNames)

	var buf strings.Builder
	fmt.Fprintf(&buf, "backend = %q\n", existingBackend)
	for _, name := range allNames {
		b, err := backend.Get(name)
		fmt.Fprintf(&buf, "\n[backends.%s]\n", name)
		if err != nil {
			for k, v := range existingParams[name] {
				fmt.Fprintf(&buf, "%s = %q\n", k, v)
			}
			continue
		}
		for _, f := range b.ConfigFields() {
			fmt.Fprintf(&buf, "%s = %q\n", f.Key, existingParams[name][f.Key])
		}
	}

	if err := os.WriteFile(cfgPath, []byte(buf.String()), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 写入配置失败: %v\n", err)
		os.Exit(1)
	}

	if err := example.Generate(dir); err != nil {
		fmt.Fprintf(os.Stderr, "警告: 生成示例配置失败: %v\n", err)
	}

	fmt.Println()
	fmt.Printf("配置已写入: %s\n", cfgPath)
	fmt.Printf("示例配置: %s\n", filepath.Join(dir, "config.example.toml"))
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
