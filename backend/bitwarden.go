package backend

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

func init() {
	Register(&Bitwarden{})
}

// Bitwarden 通过 Bitwarden CLI 获取密码短语
type Bitwarden struct{}

func (b *Bitwarden) Name() string { return "bitwarden" }

func (b *Bitwarden) GetPassphrase(params map[string]string) (string, error) {
	itemName := params["item_name"]
	fieldName := params["field_name"]
	if itemName == "" || fieldName == "" {
		return "", fmt.Errorf("bitwarden 后端需要 item_name 和 field_name 参数")
	}

	out, err := exec.Command("bw", "get", "item", itemName).Output()
	if err != nil {
		return "", fmt.Errorf("bw get item %q 失败: %w", itemName, err)
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
