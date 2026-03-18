package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/YewFence/gpg-unlock/backend"
)

func runGenExample() {
	// 生成到当前工作目录（项目根目录）
	dir := "."
	if err := generateExampleConfig(dir); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("示例配置已写入: %s\n", filepath.Join(dir, "config.example.toml"))
}

func generateExampleConfig(dir string) error {
	names := backend.Names()
	sort.Strings(names)

	var buf strings.Builder
	buf.WriteString("# 此文件由 gpg-unlock 自动生成，仅供参考，请勿直接使用。\n")
	buf.WriteString("# 复制为 config.toml 并按需修改：\n")
	buf.WriteString("#   Linux/macOS: ~/.config/gpg-unlock/config.toml\n")
	buf.WriteString("#   Windows:     %APPDATA%\\gpg-unlock\\config.toml\n")
	buf.WriteString("\n")

	if len(names) > 0 {
		fmt.Fprintf(&buf, "backend = %q\n", names[0])
	}

	for _, name := range names {
		b, err := backend.Get(name)
		if err != nil {
			continue
		}
		buf.WriteString("\n")
		fmt.Fprintf(&buf, "[backends.%s]\n", name)
		for _, f := range b.ConfigFields() {
			if f.Comment != "" {
				fmt.Fprintf(&buf, "# %s\n", f.Comment)
			}
			if f.Required {
				fmt.Fprintf(&buf, "%s = %q\n", f.Key, f.Example)
			} else {
				fmt.Fprintf(&buf, "# %s = %q\n", f.Key, f.Example)
			}
		}
	}

	examplePath := filepath.Join(dir, "config.example.toml")
	return os.WriteFile(examplePath, []byte(buf.String()), 0o644)
}
