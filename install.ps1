param(
    [string]$ScriptDir
)

$target = "$ScriptDir\gpg-load-passphrase.ps1"
$line = "function gpgunlock { & '$target' }"
$profileDir = Split-Path $PROFILE
if (!(Test-Path $profileDir)) { New-Item -ItemType Directory -Path $profileDir -Force | Out-Null }
if (!(Test-Path $PROFILE)) { New-Item -ItemType File -Path $PROFILE -Force | Out-Null }
if (Select-String -Path $PROFILE -Pattern 'gpgunlock' -Quiet) {
    Write-Host "已存在于 $PROFILE，跳过"
} else {
    Add-Content -Path $PROFILE -Value ''
    Add-Content -Path $PROFILE -Value '# gpgunlock'
    Add-Content -Path $PROFILE -Value $line
    Write-Host "✓ 已添加到 $PROFILE"
}
Write-Host '重启 PowerShell 后生效'
