# gpg-unlock

从密码管理器获取 GPG 密码短语并注入 gpg-agent 缓存，实现 Git 免密签名。

通过 Backend interface 支持多种密码管理器后端

## 为什么需要这个工具？

虽然 GPG 提供了 `gpg-preset-passphrase` 工具来预设密码短语，但将其与密码管理器集成仍需要手动操作：

| 方案 | 优点 | 缺点 |
|------|------|------|
| **手动输入密码** | 简单直接 | 每次签名都要输入，体验差 |
| **移除密钥密码** | 完全免密 | 安全风险高，私钥裸奔 |
| **pam-gnupg** | 登录时自动解锁 | 依赖 PAM，密码必须与系统登录密码相同 |
| **手写脚本** | 灵活 | 需要自己处理 keygrip 获取、错误处理、多后端支持 |
| **gpg-unlock** | 开箱即用，支持多密码管理器，自动化完整流程 | 需要安装额外工具 |

**gpg-unlock 的独特之处：**
- ✅ 支持多个密码管理器后端（Bitwarden、Infisical），易于扩展
- ✅ 自动获取 keygrip 并注入密码短语到 gpg-agent
- ✅ 智能缓存检测，避免重复调用密码管理器
- ✅ 完整的配置管理和交互式初始化向导
- ✅ 跨平台支持（Linux、macOS、Windows）
- ✅ 单一二进制文件，无运行时依赖

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
npm install -g bitwarden-cli-bio
```

也兼容官方 [Bitwarden CLI](https://bitwarden.com/help/cli/)（`bw`），但需要手动设置 session key。

### Infisical 后端

使用 [Infisical CLI](https://infisical.com/docs/cli/overview) 从 Infisical 平台获取密码短语。

前置要求：
1. 安装 Infisical CLI
2. 通过 `infisical login` 登录（用户登录或 Machine Identity）
3. 项目目录中需有 `.infisical.json`（通过 `infisical init` 生成），或在配置中指定 `project_id`

## 安装

### Homebrew (macOS / Linux)

```bash
brew tap YewFence/tap
brew install gpg-unlock
```

> Homebrew formula 发布到 `YewFence/homebrew-tap`，对应 tap 名称为 `YewFence/tap`。

### 从 Release 下载

前往 [GitHub Releases](https://github.com/YewFence/gpg-unlock/releases/latest) 下载对应平台的压缩包。

```bash
# Linux/macOS（以 linux amd64 为例）
VERSION="$(curl -fsSL https://api.github.com/repos/YewFence/gpg-unlock/releases/latest | grep '"tag_name":' | cut -d'"' -f4 | sed 's/^v//')"
curl -fL "https://github.com/YewFence/gpg-unlock/releases/latest/download/gpg-unlock_${VERSION}_linux_amd64.tar.gz" | tar xz
sudo mv gpg-unlock /usr/local/bin/
```

```powershell
# Windows（PowerShell，以 windows amd64 为例）
$version = (Invoke-RestMethod -Uri "https://api.github.com/repos/YewFence/gpg-unlock/releases/latest").tag_name.TrimStart("v")
Invoke-WebRequest -Uri "https://github.com/YewFence/gpg-unlock/releases/latest/download/gpg-unlock_${version}_windows_amd64.zip" -OutFile "gpg-unlock.zip"
Expand-Archive gpg-unlock.zip -DestinationPath . -Force
Move-Item gpg-unlock.exe "$env:USERPROFILE\bin\gpg-unlock.exe" -Force
```

可用平台：`linux_amd64`、`linux_arm64`、`darwin_amd64`、`darwin_arm64`、`windows_amd64`、`windows_arm64`

### Go

```bash
go install github.com/YewFence/gpg-unlock@latest
```

### 从源码编译

#### Go

```bash
git clone https://github.com/YewFence/gpg-unlock.git
cd gpg-unlock
just install
# 如果没有安装 Just 也可以手动安装
go install .
```

#### Docker

```bash
git clone https://github.com/YewFence/gpg-unlock.git
cd gpg-unlock
docker compose run --rm build
```

构建完成后，二进制文件会生成在 `dist/` 目录下，选择对应平台的文件并手动放入 `PATH` 中：

```bash
# Linux/macOS 示例
sudo cp dist/gpg-unlock_linux_amd64_v1/gpg-unlock /usr/local/bin/gpg-unlock

# Windows 示例（PowerShell）
copy dist\gpg-unlock_windows_amd64_v1\gpg-unlock.exe C:\Windows\System32\gpg-unlock.exe
```

## 配置

```bash
gpg-unlock init
```

交互式向导会引导你选择后端并生成配置文件到 `~/.config/gpg-unlock/config.toml`。

也可以参考具体[配置示例](./config.example.toml)自行配置

## 使用

### 基本用法

```bash
gpg-unlock
```

程序会自动检测 gpg-agent 中是否已缓存密码短语：
- 如果**已缓存**，直接退出，无需调用密码管理器
- 如果**未缓存**或**部分缓存**，从密码管理器获取密码并注入

### 强制重新注入

```bash
gpg-unlock --force
```

跳过缓存检测，强制重新从密码管理器获取密码并注入到 gpg-agent。适用于：
- 密码已更改，需要刷新缓存
- 怀疑缓存状态异常
- 需要确保使用最新密码

### 指定后端

```bash
gpg-unlock --backend bitwarden
# 或使用简写
gpg-unlock -b infisical
```

运行时指定使用的后端，覆盖配置文件中的默认值。适用于：
- 配置了多个后端，需要临时切换
- 测试不同后端的配置

### 编辑配置

```bash
gpg-unlock edit
```

使用默认编辑器打开配置文件进行编辑。编辑器优先级：
1. `$EDITOR` 环境变量指定的编辑器
2. `vi`（Linux/macOS）
3. `notepad`（Windows）

## 贡献新后端

欢迎为其他密码管理器贡献后端实现！本项目的后端接口设计简洁，易于扩展。

### 如何添加新后端

1. 在 `backend/` 目录下创建新文件（如 `backend/yourmanager.go`）
2. 实现 `Backend` interface：具体要求与说明请参考[代码中的注释](backend/backend.go)
3. 在 `init()` 函数中调用 `Register(&YourBackend{})` 注册后端
4. 运行 `gpg-unlock gen-example` 重新生成 `config.example.toml`
5. 更新 README

可以参考现有的 [Bitwarden](./backend/bitwarden.go) 和 [Infisical](./backend/infisical.go) 后端实现。

欢迎提交 Pull Request！

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
