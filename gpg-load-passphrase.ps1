# GPG 密码短语自动加载脚本（从 Bitwarden）- PowerShell 版本
# 使用方法：运行此脚本解锁 Bitwarden 并缓存 GPG 密码短语

$ErrorActionPreference = "Stop"

# 加载配置
$ConfigFile = "$env:USERPROFILE\.gnupg\bw-gpg-config"
if (Test-Path $ConfigFile) {
    Get-Content $ConfigFile | ForEach-Object {
        if ($_ -match '^([^=]+)=(.+)$') {
            $name = $matches[1].Trim()
            $value = $matches[2].Trim().Trim('"')
            Set-Variable -Name $name -Value $value -Scope Script
        }
    }
} else {
    Write-Host "错误：配置文件不存在：$ConfigFile" -ForegroundColor Red
    exit 1
}

# 检查必要的工具
$requiredCommands = @('bw', 'jq', 'gpg')
foreach ($cmd in $requiredCommands) {
    if (-not (Get-Command $cmd -ErrorAction SilentlyContinue)) {
        Write-Host "错误：未找到命令 $cmd" -ForegroundColor Red
        exit 1
    }
}

Write-Host "=== Bitwarden GPG 密码短语加载器 ===" -ForegroundColor Cyan
Write-Host

# 从 Bitwarden 获取 GPG 密码短语（如需认证会自动提示）
Write-Host "正在从 Bitwarden 获取密码短语..."
Write-Host "（如需认证，请在桌面应用中确认校验码）" -ForegroundColor Yellow
Write-Host

$itemJson = bw get item $BW_ITEM_NAME

if ($LASTEXITCODE -ne 0) {
    Write-Host "错误：无法获取项目 '$BW_ITEM_NAME'" -ForegroundColor Red
    exit 1
}

# 解析 JSON 并提取密码短语字段
$item = $itemJson | ConvertFrom-Json
$passphraseField = $item.fields | Where-Object { $_.name -eq $BW_FIELD_NAME } | Select-Object -First 1
$passphrase = $passphraseField.value

if ([string]::IsNullOrEmpty($passphrase)) {
    Write-Host "错误：无法找到字段 '$BW_FIELD_NAME'" -ForegroundColor Red
    exit 1
}

Write-Host "✓ 已获取密码短语" -ForegroundColor Green
Write-Host

# 获取 GPG 密钥的 keygrip
Write-Host "正在获取 GPG 密钥信息..."
$gpgOutput = gpg --with-keygrip -K 2>$null | Out-String
$keygrip = ($gpgOutput -split "`n" | Select-String -Pattern "Keygrip" | Select-Object -First 1) -replace '.*Keygrip\s*=\s*(\S+).*', '$1'

if ([string]::IsNullOrEmpty($keygrip)) {
    Write-Host "错误：无法获取 GPG 密钥的 keygrip" -ForegroundColor Red
    Write-Host "请确保你已经创建了 GPG 密钥" -ForegroundColor Yellow
    exit 1
}

Write-Host "✓ 密钥 Keygrip: $keygrip" -ForegroundColor Green
Write-Host

# 检查 gpg-preset-passphrase 工具
$gpgPath = (Get-Command gpg -ErrorAction SilentlyContinue).Source
if ($gpgPath) {
    $gpgDir = Split-Path $gpgPath -Parent
    $presetPaths = @(
        "$gpgDir\gpg-preset-passphrase.exe",
        "$gpgDir\..\libexec\gpg-preset-passphrase.exe",
        "/usr/lib/gnupg/gpg-preset-passphrase",
        "/usr/libexec/gpg-preset-passphrase",
        "/usr/local/libexec/gpg-preset-passphrase"
    )
} else {
    $presetPaths = @()
}

$presetCmd = $null
foreach ($path in $presetPaths) {
    if (Test-Path $path) {
        $presetCmd = $path
        break
    }
}

if ($null -eq $presetCmd) {
    Write-Host "警告：未找到 gpg-preset-passphrase，尝试使用 pinentry-mode loopback" -ForegroundColor Yellow
    # 备用方案：使用 loopback 模式测试密码短语
    $tempFile = [System.IO.Path]::GetTempFileName()
    try {
        $passphrase | gpg --batch --yes --passphrase-fd 0 --pinentry-mode loopback -s -o $tempFile $tempFile 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✓ GPG 密码短语已缓存（loopback 模式）" -ForegroundColor Green
        } else {
            Write-Host "✗ 缓存密码短语失败" -ForegroundColor Red
            exit 1
        }
    } finally {
        if (Test-Path $tempFile) {
            Remove-Item $tempFile -Force
        }
    }
} else {
    # 使用 gpg-preset-passphrase 缓存密码短语
    Write-Host "正在缓存密码短语到 gpg-agent..."
    $passphrase | & $presetCmd -c $keygrip

    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ GPG 密码短语已缓存到 gpg-agent" -ForegroundColor Green
    } else {
        Write-Host "✗ 缓存密码短语失败" -ForegroundColor Red
        exit 1
    }
}

Write-Host
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host "✓ 完成！现在可以无感签名了" -ForegroundColor Green
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
