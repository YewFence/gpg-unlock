# gpg-unlock

从密码管理器获取 GPG 密码短语并注入 gpg-agent 缓存，实现 Git 签名免密。

通过 Backend interface 支持多种密码管理器后端

## 依赖

- GPG
- 密码管理器 CLI（取决于所选后端）

### 可选的依赖
- [just](https://github.com/casey/just)
> 用于自动化流程如安装/卸载

### 可选的后端
- [Bitwarden CLI](https://github.com/bitwarden/clients)
- [bitwarden-cli-bio](https://github.com/jeanregisser/bitwarden-cli-bio)
- [Infisical CLI](https://infisical.com/docs/cli/overview)

### Bitwarden 后端

推荐使用 [bitwarden-cli-bio](https://github.com/jeanregisser/bitwarden-cli-bio)（`bwbio`），支持系统生物认证（指纹/面容）解锁 vault，若使用官方 CLI 需要先手动解锁并配置 session key 环境变量

安装：

```bash
npm i -g @nicolo-ribaudo/bitwarden-cli-bio
```

也兼容官方 [Bitwarden CLI](https://bitwarden.com/help/cli/)（`bw`），但需要手动输入主密码或管理 session key。

### Infisical 后端

使用 [Infisical CLI](https://infisical.com/docs/cli/overview) 从 Infisical 平台获取密码短语。

前置要求：
1. 安装 Infisical CLI
2. 通过 `infisical login` 登录（用户登录或 Machine Identity）
3. 项目目录中需有 `.infisical.json`（通过 `infisical init` 生成），或在配置中指定 `project_id`

## 安装

### 使用 Go

```bash
go install github.com/YewFence/gpg-unlock@latest
```

### 从源码编译

```bash
git clone https://github.com/YewFence/gpg-unlock.Git
just install
# 如果没有安装 Just 也可以手动安装
go install .
```

## 配置

```bash
gpg-unlock init
```

交互式向导会引导你选择后端并生成配置文件到 `~/.config/gpg-unlock/config.toml`。

也可以参考具体[配置示例](./config.example.toml)自行配置

## 使用

```bash
gpg-unlock
```

## 卸载

### 使用 Justfile

```bash
just uninstall
# 这会删除配置文件与可执行文件，实现干净卸载
```

### 手动卸载

```bash
gpg-unlock reset
go clean -i .
```

## 许可证

MIT © YewFence 2026
