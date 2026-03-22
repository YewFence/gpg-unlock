# gpg-unlock

从密码管理器获取 GPG 密码短语并注入 gpg-agent 缓存，实现 Git 签名免密。

## 项目结构

- `main.go` — 极简入口，仅调用 `cmd.Execute()`
- `cmd/` — CLI 子命令与主流程
  - `root.go` — cobra rootCmd 定义 + 默认 unlock 流程
  - `init.go` — 交互式配置向导
  - `reset.go` / `edit.go` / `genexample.go` — 各子命令
- `internal/config/` — Config 结构体、TOML 加载、XDG 路径查找
- `internal/gpg/` — GPG 交互：keygrip 获取、preset-passphrase、loopback、缓存检测
- `internal/example/` — 示例配置文件生成
- `backend/` — Backend interface 定义 + registry + 各后端实现
- `compose.yml` — dev 容器 + 跨平台构建容器

## 常用命令

- `just run` — go run 本地运行
- `just build` — 编译当前平台二进制
- `just install` — go install 安装到 $GOBIN
- `just build-all` — docker compose 跨平台编译（输出到 dist/）
- `just fmt` / `just vet` — 格式化 / 静态检查

## 开发约定

- 外部依赖尽量少，目前有 `github.com/BurntSushi/toml` 和 `github.com/spf13/cobra`
- CLI 使用 cobra 框架，子命令在各自文件的 `init()` 中通过 `rootCmd.AddCommand()` 注册
- 新增后端：在 `backend/` 下新建文件，实现 `Backend` interface，`init()` 中调用 `Register()`
- 配置格式为 TOML，结构体在 `internal/config/config.go` 中定义
- 后端参数统一用 `map[string]string`，各后端自行校验所需字段
