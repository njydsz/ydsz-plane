#!/usr/bin/env bash
# ============================================================
# macOS 签名与公证脚本
# ============================================================
# 用法：
#   ./scripts/sign-macos.sh <app-path> [--notary]
#
# 前置：
#   - Xcode Command Line Tools
#   - Developer ID Application 证书（Keychain）
#   - notarytool（Xcode 自带）
# ============================================================
set -euo pipefail

APP_PATH="${1:-}"
NOTARY="${2:-}"

if [[ -z "$APP_PATH" ]]; then
  echo "用法: $0 <app-path> [--notary]"
  exit 1
fi

if [[ -f ".signing.env" ]]; then
  # shellcheck disable=SC1091
  source .signing.env
fi

SIGN_IDENTITY="${MACOS_SIGN_IDENTITY:-Developer ID Application}"
TEAM_ID="${MACOS_TEAM_ID:-}"
NOTARY_ENABLED="${MACOS_NOTARY_ENABLED:-false}"

if [[ "${MACOS_SIGN_ENABLED:-false}" != "true" ]]; then
  echo "macOS 签名未启用（MACOS_SIGN_ENABLED != true），跳过"
  exit 0
fi

echo "签名 macOS 应用: $APP_PATH"

# 1. 深度签名
codesign \
  --sign "$SIGN_IDENTITY" \
  --force \
  --deep \
  --options runtime \
  --entitlements "build/entitlements.plist" \
  --timestamp \
  "$APP_PATH"

# 2. 验证签名
echo "验证签名..."
codesign --verify --deep --strict --verbose=2 "$APP_PATH"

# 3. 显示签名信息
codesign -dv --verbose=4 "$APP_PATH" 2>&1 | head -10

# 4. 公证
if [[ "$NOTARY" == "--notary" || "$NOTARY_ENABLED" == "true" ]]; then
  echo "提交公证..."
  if [[ -z "${MACOS_NOTARY_APPLE_ID:-}" || -z "${MACOS_NOTARY_PASSWORD:-}" ]]; then
    echo "缺少公证凭证（MACOS_NOTARY_APPLE_ID / MACOS_NOTARY_PASSWORD）"
    exit 1
  fi

  xcrun notarytool submit "$APP_PATH" \
    --apple-id "$MACOS_NOTARY_APPLE_ID" \
    --password "$MACOS_NOTARY_PASSWORD" \
    --team-id "${MACOS_NOTARY_TEAM_ID:-$TEAM_ID}" \
    --wait

  echo "公证完成，staple..."
  xcrun stapler staple "$APP_PATH"

  echo "验证 staple..."
  spctl --assess --type execute --verbose=2 "$APP_PATH"
fi

echo "macOS 签名完成"
