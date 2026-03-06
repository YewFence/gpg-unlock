#!/usr/bin/env bash
# GPG 密码短语自动加载脚本（从 Bitwarden）
# 使用方法：运行此脚本解锁 Bitwarden 并缓存 GPG 密码短语

set -e

# 加载配置
SCRIPT_DIR=$(dirname "$(realpath "$0")")
CONFIG_FILE="$SCRIPT_DIR/config"
if [ -f "$CONFIG_FILE" ]; then
    source "$CONFIG_FILE"
else
    echo "错误：配置文件不存在：$CONFIG_FILE"
    echo "请复制 config.example 为 config 并填写你的配置"
    exit 1
fi

# 检查必要的工具
for cmd in bw jq gpg; do
    if ! command -v $cmd &> /dev/null; then
        echo "错误：未找到命令 $cmd"
        exit 1
    fi
done

echo "=== Bitwarden GPG 密码短语加载器 ==="
echo

# 从 Bitwarden 获取 GPG 密码短语（会自动触发登录流程）
echo "正在从 Bitwarden 获取密码短语..."
echo "（如需认证，请在桌面应用中确认校验码）"
echo

ITEM_JSON=$(bw get item "$BW_ITEM_NAME" 2>&1)

if [ $? -ne 0 ]; then
    echo "错误：无法获取项目 '$BW_ITEM_NAME'"
    exit 1
fi

# 解析密码短语字段
PASSPHRASE=$(echo "$ITEM_JSON" | jq -r ".fields[] | select(.name==\"$BW_FIELD_NAME\") | .value")

if [ -z "$PASSPHRASE" ] || [ "$PASSPHRASE" = "null" ]; then
    echo "错误：无法找到字段 '$BW_FIELD_NAME'"
    exit 1
fi

echo "✓ 已获取密码短语"
echo

# 获取 GPG 密钥的 keygrip
echo "正在获取 GPG 密钥信息..."
KEYGRIP=$(gpg --with-keygrip -K 2>/dev/null | grep "Keygrip" | head -1 | awk '{print $3}')

if [ -z "$KEYGRIP" ]; then
    echo "错误：无法获取 GPG 密钥的 keygrip"
    echo "请确保你已经创建了 GPG 密钥"
    exit 1
fi

echo "✓ 密钥 Keygrip: $KEYGRIP"
echo

# 检查 gpg-preset-passphrase 工具
PRESET_CMD=""
for path in /usr/lib/gnupg/gpg-preset-passphrase /usr/libexec/gpg-preset-passphrase /usr/local/libexec/gpg-preset-passphrase; do
    if [ -x "$path" ]; then
        PRESET_CMD="$path"
        break
    fi
done

if [ -z "$PRESET_CMD" ]; then
    echo "警告：未找到 gpg-preset-passphrase，尝试使用 pinentry-mode loopback"
    # 备用方案：使用 loopback 模式
    echo "$PASSPHRASE" | gpg --batch --yes --passphrase-fd 0 --pinentry-mode loopback -s -o /dev/null /dev/null 2>/dev/null
    if [ $? -eq 0 ]; then
        echo "✓ GPG 密码短语已缓存（loopback 模式）"
    else
        echo "✗ 缓存密码短语失败"
        exit 1
    fi
else
    # 使用 gpg-preset-passphrase 缓存密码短语
    echo "正在缓存密码短语到 gpg-agent..."
    echo "$PASSPHRASE" | "$PRESET_CMD" --preset "$KEYGRIP"

    if [ $? -eq 0 ]; then
        echo "✓ GPG 密码短语已缓存到 gpg-agent"
    else
        echo "✗ 缓存密码短语失败"
        exit 1
    fi
fi

echo
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✓ 完成！现在可以无感签名了"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
