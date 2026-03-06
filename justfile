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

# 安装 gpgunlock 函数到 PowerShell Profile
install-ps:
    @powershell -ExecutionPolicy Bypass -Command \
        "$target = '{{justfile_directory()}}\\gpg-load-passphrase.ps1'; \
        $line = \"function gpgunlock { & '$target' }\"; \
        $profileDir = Split-Path $PROFILE; \
        if (!(Test-Path $profileDir)) { New-Item -ItemType Directory -Path $profileDir -Force | Out-Null }; \
        if (!(Test-Path $PROFILE)) { New-Item -ItemType File -Path $PROFILE -Force | Out-Null }; \
        if (Select-String -Path $PROFILE -Pattern 'gpgunlock' -Quiet) { \
            Write-Host '已存在于 ' + $PROFILE + '，跳过' \
        } else { \
            Add-Content -Path $PROFILE -Value ''; \
            Add-Content -Path $PROFILE -Value '# gpgunlock'; \
            Add-Content -Path $PROFILE -Value $line; \
            Write-Host ('✓ 已添加到 ' + $PROFILE) \
        }; \
        Write-Host '重启 PowerShell 后生效'"
