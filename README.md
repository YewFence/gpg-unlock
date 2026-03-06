# gpgunlock

从 Bitwarden 自动加载 GPG 密码短语到 gpg-agent，实现无感 Git 提交签名。

## 依赖

- [Bitwarden CLI](https://bitwarden.com/help/cli/) (`bw`)
- [jq](https://jqlang.github.io/jq/)
- GPG

## 配置

```bash
cp config.example config
# 编辑 config 填写你的 Bitwarden 条目信息
```

`config` 内容格式：

```ini
BW_ITEM_NAME=你的Bitwarden条目名称
BW_FIELD_NAME=密码短语字段名
```

## 使用

```bash
just run        # 自动检测系统并运行一次
just install    # 安装 gpgunlock 别名到 ~/.bashrc / ~/.zshrc
just install-ps # 安装 gpgunlock 函数到 PowerShell Profile
```

安装后直接在终端运行：

```bash
gpgunlock
```

## 许可证

MIT © YewFence 2026
