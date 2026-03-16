package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
)

// Config 应用配置
type Config struct {
	Backend  string                       `toml:"backend"`
	Backends map[string]map[string]string `toml:"backends"`
}

// loadConfig 从指定路径或默认位置加载配置
func loadConfig(path string) (*Config, error) {
	if path == "" {
		var err error
		path, err = findConfigFile()
		if err != nil {
			return nil, err
		}
	}

	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("读取配置文件 %q 失败: %w", path, err)
	}

	if cfg.Backend == "" {
		return nil, fmt.Errorf("配置文件中未指定 backend")
	}

	return &cfg, nil
}

// findConfigFile 按优先级查找配置文件
func findConfigFile() (string, error) {
	candidates := configPaths()
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("未找到配置文件，请创建以下任一文件:\n  %s", candidates[0])
}

// configPaths 返回配置文件候选路径
func configPaths() []string {
	var paths []string

	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		paths = append(paths, filepath.Join(xdg, "gpg-unlock", "config.toml"))
	}

	if runtime.GOOS == "windows" {
		if appdata := os.Getenv("APPDATA"); appdata != "" {
			paths = append(paths, filepath.Join(appdata, "gpg-unlock", "config.toml"))
		}
	} else {
		if home, err := os.UserHomeDir(); err == nil {
			paths = append(paths, filepath.Join(home, ".config", "gpg-unlock", "config.toml"))
		}
	}

	return paths
}
