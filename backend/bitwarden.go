package backend

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func init() {
	Register(&Bitwarden{})
}

// Bitwarden 通过 Bitwarden CLI 获取密码短语
type Bitwarden struct{}

func (b *Bitwarden) Name() string { return "bitwarden" }

func (b *Bitwarden) ConfigFields() []ConfigField {
	return []ConfigField{
		{Key: "command", Prompt: "CLI 命令 (command)", Required: false, Example: "bw", Comment: "Bitwarden CLI 命令，支持生物认证可改为 bwbio", DefaultValue: "bw"},
		{Key: "item_name", Prompt: "Bitwarden 项目名称 (item_name)", Required: true, Example: "GPG", Comment: "Bitwarden 中存储 GPG 密码短语的项目名称"},
		{Key: "field_name", Prompt: "字段名称 (field_name)", Required: true, Example: "passphrase", Comment: "项目中存储密码短语的自定义字段名"},
	}
}

func (b *Bitwarden) GetPassphrase(params map[string]string) (string, error) {
	itemName := params["item_name"]
	fieldName := params["field_name"]
	if itemName == "" || fieldName == "" {
		return "", fmt.Errorf("bitwarden 后端需要 item_name 和 field_name 参数")
	}

	bin := params["command"]
	if bin == "" {
		bin = "bw"
	}

	cmd := exec.Command(bin, "get", "item", itemName)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("%s get item %q 失败: %w", bin, itemName, err)
	}

	var item struct {
		Fields []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"fields"`
	}
	if err := json.Unmarshal(out, &item); err != nil {
		return "", fmt.Errorf("解析 bw 输出失败: %w", err)
	}

	for _, f := range item.Fields {
		if f.Name == fieldName {
			v := strings.TrimSpace(f.Value)
			if v == "" {
				return "", fmt.Errorf("字段 %q 值为空", fieldName)
			}
			return v, nil
		}
	}
	return "", fmt.Errorf("未找到字段 %q", fieldName)
}
