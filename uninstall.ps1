if (!(Test-Path $PROFILE)) { Write-Host '未找到 PowerShell Profile，跳过'; exit 0 }
$content = Get-Content $PROFILE
$filtered = $content | Where-Object { $_ -notmatch 'gpgunlock' -and $_ -notmatch '^# gpgunlock$' }
if ($filtered.Count -eq $content.Count) {
    Write-Host '未在 Profile 中找到 gpgunlock，跳过'
} else {
    $filtered | Set-Content $PROFILE
    Write-Host "✓ 已从 $PROFILE 移除"
}
Write-Host '重启 PowerShell 后生效'
