package backend

import "fmt"

// Backend 定义密码获取接口
type Backend interface {
	// Name 返回后端名称，用于配置匹配
	Name() string
	// GetPassphrase 从密码管理器获取 GPG 密码短语
	GetPassphrase(params map[string]string) (string, error)
}

var registry = map[string]Backend{}

// Register 注册一个后端实现
func Register(b Backend) {
	registry[b.Name()] = b
}

// Get 根据名称获取后端
func Get(name string) (Backend, error) {
	b, ok := registry[name]
	if !ok {
		names := make([]string, 0, len(registry))
		for k := range registry {
			names = append(names, k)
		}
		return nil, fmt.Errorf("未知后端 %q，可用: %v", name, names)
	}
	return b, nil
}
