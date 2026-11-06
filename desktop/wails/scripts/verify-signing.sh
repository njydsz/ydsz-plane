#!/usr/bin/env bash
# ============================================================
# 签名验证脚本（跨平台）
# ============================================================
# 用法：
#   ./scripts/verify-signing.sh <path>
#
# 支持：
#   - macOS：codesign / spctl / stapler
#   - Windows：Get-AuthenticodeSignature（通过 PowerShell）
#   - Linux：GPG（可选）
# ============================================================
set -euo pipefail

TARGET="${1:-}"

if [[ -z "$TARGET" ]]; then
  echo "用法: $0 <path>"
  exit 1
fi

OS="$(uname -s)"
FAILED=0

case "$OS" in
  Darwin)
    echo "=== macOS 签名验证 ==="
    if [[ -d "$TARGET" ]]; then
      # .app 包
      codesign --verify --deep --strict --verbose=2 "$TARGET" || FAILED=1
      echo "--- 签名详情 ---"
      codesign -dv --verbose=4 "$TARGET" 2>&1 | head -15
      echo "--- Gatekeeper 评估 ---"
      spctl --assess --type execute --verbose=2 "$TARGET" || FAILED=1
    else
      echo "非 .app 目录，跳过深度验证"
    fi
    ;;

  Linux)
    echo "=== Linux 签名验证 ==="
    if command -v gpg &>/dev/null && [[ -f "$TARGET.sig" ]]; then
      gpg --verify "$TARGET.sig" "$TARGET" || FAILED=1
    else
      echo "无 GPG 签名（Linux 通常不做二进制签名）"
    fi
    ;;

  MINGW*|MSYS*|CYGWIN*)
    echo "=== Windows 签名验证 ==="
    powershell -NoProfile -Command "
      \$s = Get-AuthenticodeSignature '$TARGET'
      Write-Host ('状态: ' + \$s.Status)
      Write-Host ('签名者: ' + \$s.SignerCertificate.Subject)
      Write-Host ('颁发者: ' + \$s.SignerCertificate.Issuer)
      Write-Host ('时间戳: ' + \$s.TimeStamperCertificate.Subject)
      if (\$s.Status -ne 'Valid') { exit 1 }
    " || FAILED=1
    ;;

  *)
    echo "未知平台: $OS"
    exit 1
    ;;
esac

if [[ $FAILED -eq 0 ]]; then
  echo ""
  echo "签名验证通过"
  exit 0
else
  echo ""
  echo "签名验证失败"
  exit 1
fi
