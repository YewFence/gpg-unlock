package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// getKeygrip 获取第一个 GPG 私钥的 keygrip，带重试
func getKeygrip() (string, error) {
	for i := 0; i < 3; i++ {
		out, err := exec.Command("gpg", "--with-keygrip", "-K").Output()
		if err == nil {
			for _, line := range strings.Split(string(out), "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "Keygrip") || strings.Contains(line, "Keygrip") {
					parts := strings.SplitN(line, "=", 2)
					if len(parts) == 2 {
						kg := strings.TrimSpace(parts[1])
						if kg != "" {
							return kg, nil
						}
					}
				}
			}
		}
		if i < 2 {
			fmt.Printf("gpg-agent 尚未就绪，等待重试（%d/3）...\n", i+1)
			time.Sleep(2 * time.Second)
		}
	}
	return "", fmt.Errorf("无法获取 GPG 密钥的 keygrip，请确保已创建 GPG 密钥")
}

// presetPassphrase 将密码短语注入 gpg-agent
func presetPassphrase(keygrip, passphrase string) error {
	presetCmd := findPresetCommand()
	if presetCmd != "" {
		cmd := exec.Command(presetCmd, "--preset", keygrip)
		cmd.Stdin = strings.NewReader(passphrase)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("gpg-preset-passphrase 失败: %w", err)
		}
		fmt.Println("密码短语已缓存到 gpg-agent")
		return nil
	}

	fmt.Println("未找到 gpg-preset-passphrase，使用 loopback 模式...")
	return loopbackSign(passphrase)
}

// loopbackSign 使用 loopback 模式缓存密码短语
func loopbackSign(passphrase string) error {
	cmd := exec.Command("gpg", "--batch", "--yes", "--passphrase-fd", "0", "--pinentry-mode", "loopback", "-s", "-o", os.DevNull, os.DevNull)
	cmd.Stdin = strings.NewReader(passphrase)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("loopback 模式缓存失败: %w", err)
	}
	fmt.Println("密码短语已缓存（loopback 模式）")
	return nil
}

// findPresetCommand 在 PATH 和常见路径查找 gpg-preset-passphrase
func findPresetCommand() string {
	if p, err := exec.LookPath("gpg-preset-passphrase"); err == nil {
		return p
	}

	var candidates []string
	if runtime.GOOS == "windows" {
		if gpgPath, err := exec.LookPath("gpg"); err == nil {
			dir := filepath.Dir(gpgPath)
			candidates = append(candidates,
				filepath.Join(dir, "gpg-preset-passphrase.exe"),
				filepath.Join(dir, "..", "libexec", "gpg-preset-passphrase.exe"),
			)
		}
	} else {
		candidates = []string{
			"/usr/lib/gnupg/gpg-preset-passphrase",
			"/usr/libexec/gpg-preset-passphrase",
			"/usr/local/libexec/gpg-preset-passphrase",
		}
	}

	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}
