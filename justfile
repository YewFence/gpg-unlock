# gpgunlock - 从 Bitwarden 加载 GPG 密码短语到 gpg-agent

# 运行脚本（自动检测系统）
run:
    #!/usr/bin/env bash
    if [ "$OS" = "Windows_NT" ]; then
        powershell -ExecutionPolicy Bypass -File "{{justfile_directory()}}/gpg-load-passphrase.ps1"
    else
        bash "{{justfile_directory()}}/gpg-load-passphrase.sh"
    fi

# 安装 gpgunlock 别名到 bash/zsh 初始化文件
install:
    #!/usr/bin/env bash
    SCRIPT_DIR="{{justfile_directory()}}"
    ALIAS_LINE="alias gpgunlock='bash \"$SCRIPT_DIR/gpg-load-passphrase.sh\"'"
    INSTALLED=0

    for rc in "$HOME/.bashrc" "$HOME/.zshrc"; do
        [ -f "$rc" ] || continue
        if grep -q "gpgunlock" "$rc" 2>/dev/null; then
            echo "已存在于 $rc，跳过"
        else
            printf '\n# gpgunlock\n%s\n' "$ALIAS_LINE" >> "$rc"
            echo "✓ 已添加到 $rc"
        fi
        INSTALLED=1
    done

    if [ "$INSTALLED" -eq 0 ]; then
        echo "未找到 ~/.bashrc 或 ~/.zshrc，请手动添加以下内容："
        echo "$ALIAS_LINE"
        exit 1
    fi

    echo "重启终端或 source 对应文件后生效"

# 从 bash/zsh 初始化文件中移除 gpgunlock 别名
uninstall:
    #!/usr/bin/env bash
    REMOVED=0

    for rc in "$HOME/.bashrc" "$HOME/.zshrc"; do
        [ -f "$rc" ] || continue
        if grep -q "gpgunlock" "$rc" 2>/dev/null; then
            sed -i '/^# gpgunlock$/,/gpgunlock/d' "$rc"
            echo "✓ 已从 $rc 移除"
            REMOVED=1
        else
            echo "未在 $rc 中找到 gpgunlock，跳过"
        fi
    done

    if [ "$REMOVED" -eq 0 ]; then
        echo "未找到任何需要清理的配置"
    else
        echo "重启终端或 source 对应文件后生效"
    fi

# 从 PowerShell Profile 中移除 gpgunlock 函数
uninstall-ps:
    @pwsh -NoProfile -ExecutionPolicy Bypass -File "{{justfile_directory()}}/uninstall.ps1"

# 安装 gpgunlock 函数到 PowerShell Profile
install-ps:
    @pwsh -NoProfile -ExecutionPolicy Bypass -File "{{justfile_directory()}}/install.ps1" -ScriptDir "{{justfile_directory()}}"
