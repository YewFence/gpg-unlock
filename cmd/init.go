package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/YewFence/gpg-unlock/backend"
	"github.com/YewFence/gpg-unlock/internal/config"
	"github.com/YewFence/gpg-unlock/internal/example"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "交互式配置向导",
	Long:  "通过交互式问答创建或更新 gpg-unlock 配置文件。",
	Args:  cobra.NoArgs,
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	dir := config.Dir()
	if dir == "" {
		return fmt.Errorf("无法确定配置目录（HOME 未设置）")
	}

	cfgPath := filepath.Join(dir, "config.toml")

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
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

		// 检查现有配置中是否存在未注册的后端
		var unknown []string
		for name := range existing.Backends {
			if _, err := backend.Get(name); err != nil {
				unknown = append(unknown, name)
			}
		}
		if len(unknown) > 0 {
			sort.Strings(unknown)
			return fmt.Errorf(
				"配置文件中存在未注册的后端: %v\n"+
					"这可能是版本升级导致的不兼容变更，请手动编辑配置文件: %s",
				unknown, cfgPath,
			)
		}
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
		current := existingParams[chosen]
		if current == nil {
			current = map[string]string{}
		}
		var formatted map[string]string
		for {
			params := map[string]string{}
			for _, f := range b.ConfigFields() {
				for {
					suggestion := current[f.Key]
					if suggestion == "" {
						suggestion = f.DefaultValue
					}
					fmt.Printf("%s; 示例： %s", f.Prompt, f.Example)
					if suggestion != "" {
						fmt.Printf("(默认: %s) ", suggestion)
					}
					if f.Required {
						fmt.Print("（必填） ")
					}
					fmt.Print("：")
					val := readLine(reader)
					if val == "" {
						fmt.Printf("使用默认值：%s\n", suggestion)
						val = suggestion
					}
					// 此处默认值和必填的语义会有些歧义，需要注意
					// 理论上不应该有字段同时为必填又有默认值，因为默认值会自动填充此时必填校验永远不会触发
					if f.Required && val == "" {
						fmt.Fprintf(os.Stderr, "错误: %s 不能为空\n", f.Key)
						continue
					}
					params[f.Key] = val
					break
				}
			}
			formatted = b.FormatConfig(params)
			errs := b.ValidateConfig(formatted)
			if len(errs) == 0 {
				break
			}
			fmt.Fprintln(os.Stderr, "配置校验失败:")
			for _, e := range errs {
				fmt.Fprintf(os.Stderr, "  - %v\n", e)
			}
			fmt.Println("请重新输入配置。")
			fmt.Println()
			current = params
		}
		existingParams[chosen] = formatted
		if existingBackend == "" {
			existingBackend = chosen
		}

		fmt.Println()
		fmt.Print("继续配置另一个后端？[y/N] ")
		if !confirmYes(reader) {
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
		b, _ := backend.Get(name) // 前面已校验，此处不会失败
		fmt.Fprintf(&buf, "\n[backends.%s]\n", name)
		for _, f := range b.ConfigFields() {
			fmt.Fprintf(&buf, "%s = %q\n", f.Key, existingParams[name][f.Key])
		}
	}

	if err := os.WriteFile(cfgPath, []byte(buf.String()), 0o600); err != nil {
		return fmt.Errorf("写入配置失败: %w", err)
	}

	if err := example.Generate(dir); err != nil {
		fmt.Fprintf(os.Stderr, "警告: 生成示例配置失败: %v\n", err)
	}

	fmt.Println()
	fmt.Printf("配置已写入: %s\n", cfgPath)
	fmt.Printf("示例配置: %s\n", filepath.Join(dir, "config.example.toml"))
	fmt.Println("现在可以运行 gpg-unlock 来加载密码短语了")
	return nil
}
