#!/bin/bash
# P0-6: Swagger 注解覆盖率检查
# 统计带 @Summary 注释的 handler 数量 / 总 handler 数量 = 覆盖率
# 目标：核心域（/auth / /me / /workspaces）覆盖率 ≥ 90%

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
echo "=== Swagger 注解覆盖率检查 (P0-6) ==="

# 1. 获取路由端点数（从 swagger.yaml 已解析的 Router 标签）
SWAGGER_YAML="$ROOT/docs/swagger/swagger.yaml"
if [ ! -f "$SWAGGER_YAML" ]; then
  echo "⚠️  swagger.yaml 不存在，尝试生成..."
  if command -v swag >/dev/null 2>&1; then
    cd "$ROOT" && swag init -g cmd/api/main.go -o docs/swagger/
  else
    echo "⚠️  swag 命令未安装，跳过路由数统计。安装: go install github.com/swaggo/swag"
  fi
fi

DOCUMENTED=0
if [ -f "$SWAGGER_YAML" ]; then
  DOCUMENTED=$(grep -c '@Router\|paths:' "$SWAGGER_YAML" 2>/dev/null || echo 0)
fi
echo "  已文档化端点（parsed paths in swagger.yaml）: $DOCUMENTED"

# 2. 统计 @Summary 注释数量
SUMMARY_COUNT=$(grep -rn '// @Summary\|//	@Summary' "$ROOT" --include='*.go' 2>/dev/null | wc -l | tr -d ' ')
echo "  @Summary 注解数量: $SUMMARY_COUNT"

# 3. 统计总 handler 数量（近似值：public Go method on *Handler structs）
TOTAL_HANDLERS=$(grep -rn 'func (h \*Handler)\|func (h \*IssueHandler)\|func (h \*Service)\|func.*\*Handler\).*gin\.Context' "$ROOT/internal" --include='*.go' 2>/dev/null | wc -l | tr -d ' ')
echo "  潜在 handler 方法总数（近似）: $TOTAL_HANDLERS"

# 4. 统计每个域的注解覆盖率
echo ""
echo "--- 按包分布 ---"
for pkg_dir in issue sprint version automation webhook dashboard knowledge pages intake workbench workspace auth; do
  count=$(grep -rn '// @Summary\|//	@Summary' "$ROOT/internal/application/$pkg_dir" --include='*.go' 2>/dev/null | wc -l | tr -d ' ')
  total=$(grep -rn 'func.*gin\.Context' "$ROOT/internal/application/$pkg_dir" --include='*.go' 2>/dev/null | wc -l | tr -d ' ')
  if [ "$total" -gt 0 ] || [ "$count" -gt 0 ]; then
    pct=0
    if [ "$total" -gt 0 ]; then
      pct=$((count * 100 / total))
    fi
    echo "  $pkg_dir: $count/$total handlers documented (${pct}%)"
  fi
done

# 5. 覆盖率门禁
TARGET_PCT=${SWAGGER_COVERAGE_TARGET:-80}
ACTUAL_PCT=0
if [ "$TOTAL_HANDLERS" -gt 0 ]; then
  ACTUAL_PCT=$((SUMMARY_COUNT * 100 / TOTAL_HANDLERS))
fi

echo ""
echo "=== 总结 ==="
echo "  总 @Summary: $SUMMARY_COUNT / $TOTAL_HANDLERS 方法"
echo "  总体覆盖率: ${ACTUAL_PCT}%（目标: ${TARGET_PCT}%）"

if [ "$ACTUAL_PCT" -lt "$TARGET_PCT" ]; then
  echo ""
  echo "::warning::Swagger 覆盖率 ${ACTUAL_PCT}% < 目标 ${TARGET_PCT}%。建议在 handler 函数上添加 @Summary/@Description/@Tags/@Router 注解。"
  echo "::warning::详读 docs/architecture/05-API设计规范.md 了解注解规范。"
  # 当前仅警告不阻断（渐进式推行）
  exit 0
else
  echo "✅ Swagger 覆盖率达标（${ACTUAL_PCT}% ≥ ${TARGET_PCT}%）"
fi
