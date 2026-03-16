# gpg-unlock

从密码管理器获取 GPG 密码短语并注入 gpg-agent 缓存，实现 Git 签名免密。

## 项目结构

- `main.go` — 入口，CLI flag 解析，串联流程
- `config.go` — TOML 配置加载，XDG 路径查找
- `gpg.go` — GPG 交互：keygrip 获取、gpg-preset-passphrase、loopback fallback
- `backend/backend.go` — Backend interface 定义 + registry（init 自注册模式）
- `backend/bitwarden.go` — Bitwarden CLI 后端实现
- `compose.yml` — dev 容器（golang:1.24）+ 跨平台构建容器

## 常用命令

- `just run` — go run 本地运行
- `just build` — 编译当前平台二进制
- `just install` — go install 安装到 $GOBIN
- `just build-all` — docker compose 跨平台编译（输出到 dist/）
- `just fmt` / `just vet` — 格式化 / 静态检查

## 开发约定

- 外部依赖尽量少，目前仅 `github.com/BurntSushi/toml`
- CLI 用标准库 `flag`，不引入 cobra 等
- 新增后端：在 `backend/` 下新建文件，实现 `Backend` interface，`init()` 中调用 `Register()`
- 配置格式为 TOML，结构体在 `config.go` 中定义
- 后端参数统一用 `map[string]string`，各后端自行校验所需字段
