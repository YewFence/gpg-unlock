package backend

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func init() {
	Register(&Infisical{})
}

// Infisical 通过 Infisical CLI 获取密码短语
type Infisical struct{}

func (i *Infisical) Name() string { return "infisical" }

func (i *Infisical) ConfigFields() []ConfigField {
	return []ConfigField{
		{Key: "command", Prompt: "CLI 命令 (command, 默认 infisical)", Required: false, Example: "infisical", Comment: "Infisical CLI 命令"},
		{Key: "secret_name", Prompt: "Secret 名称 (secret_name)", Required: true, Example: "GPG_PASSPHRASE", Comment: "Secret 名称（必填）"},
		{Key: "project_dir", Prompt: "项目目录路径 (project_dir, 含 .infisical.json)", Required: false, Example: "/path/to/project", Comment: "含 .infisical.json 的项目目录（与 project_id 二选一）"},
		{Key: "project_id", Prompt: "项目 ID (project_id, 与 project_dir 二选一)", Required: false, Example: "", Comment: "项目 ID（与 project_dir 二选一）"},
		{Key: "environment", Prompt: "环境 (environment, 如 dev/staging/prod)", Required: false, Example: "dev", Comment: "环境，如 dev/staging/prod"},
		{Key: "secret_path", Prompt: "Secret 路径 (secret_path, 默认 /)", Required: false, Example: "/", Comment: "Secret 路径，默认 /"},
		{Key: "domain", Prompt: "自架实例域名 (domain, 如 https://infisical.example.com)", Required: false, Example: "https://infisical.example.com", Comment: "自架实例域名"},
		{Key: "token", Prompt: "Access Token (token, 用于 CI/CD)", Required: false, Example: "", Comment: "Access Token，用于 CI/CD"},
	}
}

func (i *Infisical) GetPassphrase(params map[string]string) (string, error) {
	secretName := params["secret_name"]
	if secretName == "" {
		return "", fmt.Errorf("infisical 后端需要 secret_name 参数")
	}

	bin := params["command"]
	if bin == "" {
		bin = "infisical"
	}

	args := []string{"secrets", "get", secretName, "--plain"}

	if v := params["domain"]; v != "" {
		args = append(args, "--domain", v)
	}
	if v := params["project_id"]; v != "" {
		args = append(args, "--projectId", v)
	}
	if v := params["environment"]; v != "" {
		args = append(args, "--env", v)
	}
	if v := params["secret_path"]; v != "" {
		args = append(args, "--path", v)
	}
	if v := params["token"]; v != "" {
		args = append(args, "--token", v)
	}

	cmd := exec.Command(bin, args...)
	cmd.Stderr = os.Stderr

	if dir := params["project_dir"]; dir != "" {
		cmd.Dir = dir
	}

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("%s secrets get %q 失败: %w", bin, secretName, err)
	}

	passphrase := strings.TrimSpace(string(out))
	if passphrase == "" {
		return "", fmt.Errorf("secret %q 值为空", secretName)
	}

	return passphrase, nil
}
