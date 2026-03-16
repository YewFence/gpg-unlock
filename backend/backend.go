package backend

import "fmt"

// ConfigField 描述后端需要的一个配置字段
type ConfigField struct {
	Key      string // TOML 键名
	Prompt   string // 交互提示文本
	Required bool
}

// Backend 定义密码获取接口
type Backend interface {
	// Name 返回后端名称，用于配置匹配
	Name() string
	// ConfigFields 返回该后端所需的配置字段（用于 init 向导）
	ConfigFields() []ConfigField
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
		return nil, fmt.Errorf("未知后端 %q，可用: %v", name, Names())
	}
	return b, nil
}

// Names 返回所有已注册后端的名称
func Names() []string {
	names := make([]string, 0, len(registry))
	for k := range registry {
		names = append(names, k)
	}
	return names
}
