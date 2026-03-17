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

// getKeygrips 获取所有 GPG 私钥的 keygrip，带重试
func getKeygrips() ([]string, error) {
	for i := 0; i < 3; i++ {
		out, err := exec.Command("gpg", "--with-keygrip", "-K").Output()
		if err == nil {
			var grips []string
			for _, line := range strings.Split(string(out), "\n") {
				line = strings.TrimSpace(line)
				if strings.Contains(line, "Keygrip") {
					parts := strings.SplitN(line, "=", 2)
					if len(parts) == 2 {
						kg := strings.TrimSpace(parts[1])
						if kg != "" {
							grips = append(grips, kg)
						}
					}
				}
			}
			if len(grips) > 0 {
				return grips, nil
			}
		}
		if i < 2 {
			fmt.Printf("gpg-agent 尚未就绪，等待重试（%d/3）...\n", i+1)
			time.Sleep(2 * time.Second)
		}
	}
	return nil, fmt.Errorf("无法获取 GPG 密钥的 keygrip，请确保已创建 GPG 密钥")
}

// presetPassphrase 将密码短语注入 gpg-agent（对所有 keygrip）
func presetPassphrase(keygrips []string, passphrase string) error {
	presetCmd := findPresetCommand()
	if presetCmd != "" {
		for _, kg := range keygrips {
			cmd := exec.Command(presetCmd, "--preset", kg)
			cmd.Stdin = strings.NewReader(passphrase)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("gpg-preset-passphrase 失败 (keygrip %s): %w", kg, err)
			}
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

// checkKeygripCached 检查指定 keygrip 的密码短语是否已缓存
func checkKeygripCached(keygrip string) bool {
	cmd := exec.Command("gpg-connect-agent")
	cmd.Stdin = strings.NewReader(fmt.Sprintf("KEYINFO %s\n", keygrip))
	out, err := cmd.Output()
	if err != nil {
		return false // 检测失败视为未缓存（保守策略）
	}

	// 解析输出：S KEYINFO <keygrip> <type> <serialno> <idstr> <cached> ...
	// 第 7 个字段（索引 6）：1=已缓存，-=未缓存
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "S KEYINFO") {
			fields := strings.Fields(line)
			if len(fields) >= 7 && fields[2] == keygrip {
				return fields[6] == "1"
			}
		}
	}
	return false
}

// checkAllKeygripsCached 检查所有 keygrip 是否都已缓存
// 返回 (已缓存数量, 总数量)
func checkAllKeygripsCached(keygrips []string) (int, int) {
	cached := 0
	for _, kg := range keygrips {
		if checkKeygripCached(kg) {
			cached++
		}
	}
	return cached, len(keygrips)
}
