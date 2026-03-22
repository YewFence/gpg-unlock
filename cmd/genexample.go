package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/YewFence/gpg-unlock/internal/example"
)

func runGenExample() {
	dir := "."
	if err := example.Generate(dir); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("示例配置已写入: %s\n", filepath.Join(dir, "config.example.toml"))
}
