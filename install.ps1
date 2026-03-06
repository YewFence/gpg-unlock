param(
    [string]$ScriptDir
)

$profilePath = [System.IO.Path]::GetFullPath($PROFILE)
$target = "$ScriptDir\gpg-load-passphrase.ps1"
$line = "function gpgunlock { & '$target' }"
$profileDir = Split-Path $profilePath
if (!(Test-Path $profileDir)) { New-Item -ItemType Directory -Path $profileDir -Force | Out-Null }
if (!(Test-Path $profilePath)) { New-Item -ItemType File -Path $profilePath -Force | Out-Null }
if (Select-String -Path $profilePath -Pattern 'gpgunlock' -Quiet) {
    Write-Host "已存在于 $profilePath，跳过"
} else {
    Add-Content -Path $profilePath -Value ''
    Add-Content -Path $profilePath -Value '# gpgunlock'
    Add-Content -Path $profilePath -Value $line
    Write-Host "✓ 已添加到 $profilePath"
}
Write-Host '重启 PowerShell 后生效'
