# ============================================================
# Windows 签名脚本（Authenticode）
# ============================================================
# 用法：
#   .\scripts\sign-windows.ps1 -TargetPath <path>
#
# 前置：
#   - Windows SDK（含 signtool.exe）
#   - 代码签名证书（.pfx 或来自 Windows 证书存储）
# ============================================================

param(
    [Parameter(Mandatory = $true)]
    [string]$TargetPath
)

$ErrorActionPreference = "Stop"

# 加载签名配置
if (Test-Path ".signing.env") {
    Get-Content ".signing.env" | ForEach-Object {
        if ($_ -match '^\s*([^#][^=]+)=(.*)$') {
            $name = $matches[1].Trim()
            $value = $matches[2].Trim().Trim('"').Trim("'")
            [Environment]::SetEnvironmentVariable($name, $value, "Process")
        }
    }
}

$signEnabled = [Environment]::GetEnvironmentVariable("WINDOWS_SIGN_ENABLED")
if ($signEnabled -ne "true") {
    Write-Host "Windows 签名未启用（WINDOWS_SIGN_ENABLED != true），跳过" -ForegroundColor Yellow
    exit 0
}

$certPath = [Environment]::GetEnvironmentVariable("WINDOWS_SIGN_CERT")
$certPassword = [Environment]::GetEnvironmentVariable("WINDOWS_SIGN_PASSWORD")
$timestamp = [Environment]::GetEnvironmentVariable("WINDOWS_SIGN_TIMESTAMP")
if (-not $timestamp) { $timestamp = "http://timestamp.digicert.com" }

Write-Host "签名 Windows 应用: $TargetPath" -ForegroundColor Cyan

# --- 方式一：signtool（Windows SDK）---
$signtool = Get-ChildItem -Path "C:\Program Files (x86)\Windows Kits\10\bin" -Filter "signtool.exe" -Recurse -ErrorAction SilentlyContinue |
    Sort-Object { $_.DirectoryName } -Descending |
    Select-Object -First 1

if ($signtool) {
    Write-Host "使用 signtool: $($signtool.FullName)"

    if ($certPath -and (Test-Path $certPath)) {
        # PFX 证书文件
        & $signtool sign /f "$certPath" /p "$certPassword" /tr $timestamp /td sha256 /fd sha256 /a "$TargetPath"
    } else {
        # Windows 证书存储（按主题选择）
        & $signtool sign /n "$certPath" /tr $timestamp /td sha256 /fd sha256 /a "$TargetPath"
    }

    if ($LASTEXITCODE -ne 0) {
        Write-Error "signtool 签名失败 (exit $LASTEXITCODE)"
    }
} else {
    # --- 方式二：osslsigncode（跨平台）---
    Write-Host "未找到 signtool，尝试 osslsigncode..." -ForegroundColor Yellow
    $osslCert = [Environment]::GetEnvironmentVariable("WINDOWS_OSSL_CERT")
    $osslKey = [Environment]::GetEnvironmentVariable("WINDOWS_OSSL_KEY")

    if (-not (Get-Command osslsigncode -ErrorAction SilentlyContinue)) {
        Write-Error "未找到 osslsigncode，请安装或配置 signtool"
    }

    & osslsigncode sign `
        -certs "$osslCert" `
        -key "$osslKey" `
        -n "Ydsz Plane" `
        -i "https://ydsz-plane.example.com" `
        -t "$timestamp" `
        -h sha256 `
        -in "$TargetPath" `
        -out "$TargetPath.signed"

    Move-Item -Force "$TargetPath.signed" "$TargetPath"
}

# 验证签名
Write-Host "验证签名..." -ForegroundColor Cyan
$signature = Get-AuthenticodeSignature "$TargetPath"
Write-Host "  状态: $($signature.Status)"
Write-Host "  签名者: $($signature.SignerCertificate.Subject)"

if ($signature.Status -ne "Valid") {
    Write-Warning "签名状态异常: $($signature.Status)"
}

Write-Host "Windows 签名完成" -ForegroundColor Green
