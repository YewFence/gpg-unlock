# gpg-unlock

从密码管理器获取 GPG 密码短语并注入 gpg-agent 缓存，实现 Git 签名免密。

通过 Backend interface 支持多种密码管理器后端，当前实现了 Bitwarden 后端。

## 依赖

- GPG
- 密码管理器 CLI（取决于所选后端）

### Bitwarden 后端

推荐使用 [bitwarden-cli-bio](https://github.com/jeanregisser/bitwarden-cli-bio)（`bw-bio`），支持系统生物认证（指纹/面容）解锁 vault，体验远好于官方 CLI 的主密码交互。

安装：

```bash
npm i -g @nicolo-ribaudo/bitwarden-cli-bio
```

也兼容官方 [Bitwarden CLI](https://bitwarden.com/help/cli/)（`bw`），但需要手动输入主密码或管理 session key。

## 安装

```bash
go install github.com/YewFence/gpg-unlock@latest
```

或从 [Releases](https://github.com/YewFence/gpg-unlock/releases) 下载预编译二进制。

## 配置

```bash
gpg-unlock init
```

交互式向导会引导你选择后端并生成配置文件到 `~/.config/gpg-unlock/config.toml`。

配置示例：

```toml
backend = "bitwarden"

[backends.bitwarden]
command = "bw-bio"       # CLI 命令，默认 "bw"
item_name = "GPG"        # Bitwarden 条目名称
field_name = "passphrase" # 密码短语字段名
```

## 使用

```bash
gpg-unlock
```

## 许可证

MIT © YewFence 2026
