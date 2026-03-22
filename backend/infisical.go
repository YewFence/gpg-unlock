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
		{Key: "command", Prompt: "CLI 命令（command）", Required: false, Example: "infisical", Comment: "Infisical CLI 命令（与 docker_image 二选一）", DefaultValue: "infisical"},
		{Key: "docker_image", Prompt: "Docker 镜像名称 (docker_image, 非空则通过 docker run 调用)", Required: false, Example: "infisical/cli", Comment: "设置后用 docker run <image> 代替本地 infisical 命令"},
		{Key: "secret_name", Prompt: "Secret 名称 (secret_name)", Required: true, Example: "GPG_PASSPHRASE", Comment: "Secret 名称"},
		{Key: "project_dir", Prompt: "项目目录路径 (project_dir，与 project_id 二选一)", Required: false, Example: "/path/to/project", Comment: "含 .infisical.json 的项目目录（与 project_id 二选一）"},
		{Key: "project_id", Prompt: "项目 ID (project_id, 与 project_dir 二选一)", Required: false, Example: "fcxxxxx-xxxx-xxxx-xxxx-xxxxxxxx", Comment: "项目 ID（与 project_dir 二选一）"},
		{Key: "environment", Prompt: "环境 (environment, 如 dev/staging/prod)", Required: false, Example: "dev", Comment: "环境，如 dev/staging/prod", DefaultValue: "dev"},
		{Key: "secret_path", Prompt: "Secret 路径", Required: false, Example: "/", Comment: "Secret 路径", DefaultValue: "/"},
		{Key: "domain", Prompt: "Infisical 实例域名 (domain)", Required: false, Example: "https://app.infisical.com", Comment: "Infisical 实例域名", DefaultValue: "https://app.infisical.com"},
		{Key: "token", Prompt: "Access Token (token, 用于 CI/CD)", Required: false, Example: "", Comment: "Access Token，用于 CI/CD"},
	}
}

func buildInfisicalArgs(secretName string, params map[string]string) []string {
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
	return args
}

func (i *Infisical) GetPassphrase(params map[string]string) (string, error) {
	secretName := params["secret_name"]
	if secretName == "" {
		return "", fmt.Errorf("infisical 后端需要 secret_name 参数")
	}

	infisicalArgs := buildInfisicalArgs(secretName, params)

	var cmd *exec.Cmd
	if dockerImage := params["docker_image"]; dockerImage != "" {
		dockerArgs := []string{"run", "--rm"}
		if projectDir := params["project_dir"]; projectDir != "" {
			dockerArgs = append(dockerArgs, "-v", projectDir+":"+projectDir, "-w", projectDir)
		}
		dockerArgs = append(dockerArgs, dockerImage)
		dockerArgs = append(dockerArgs, infisicalArgs...)
		cmd = exec.Command("docker", dockerArgs...)
	} else {
		bin := params["command"]
		if bin == "" {
			bin = "infisical"
		}
		cmd = exec.Command(bin, infisicalArgs...)
		if dir := params["project_dir"]; dir != "" {
			cmd.Dir = dir
		}
	}

	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("infisical secrets get %q 失败: %w", secretName, err)
	}

	passphrase := strings.TrimSpace(string(out))
	if passphrase == "" {
		return "", fmt.Errorf("secret %q 值为空", secretName)
	}

	return passphrase, nil
}
