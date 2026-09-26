# P0-6: Swagger Annotation Coverage Checker
# PowerShell version for Windows (no bash required)

param(
    [int]$TargetPercent = 80
)

$ErrorActionPreference = "SilentlyContinue"
$Root = Split-Path -Parent $PSScriptRoot

Write-Host "=== Swagger Annotation Coverage Check (P0-6) ===" -ForegroundColor Cyan

# 1. Count @Summary annotations
$summaryCount = 0
$handlerCount = 0

# Get all Go files under internal/
$goFiles = Get-ChildItem -Path "$Root\internal" -Filter "*.go" -Recurse -File -ErrorAction SilentlyContinue

foreach ($file in $goFiles) {
    $content = Get-Content $file.FullName -Raw -ErrorAction SilentlyContinue
    # Count @Summary
    $summaryMatches = [regex]::Matches($content, '//\s*@Summary')
    $summaryCount += $summaryMatches.Count

    # Count handler methods (approximate: func with *Handler receiver and gin.Context param)
    $handlerMatches = [regex]::Matches($content, 'func\s+\(h\s+\*Handler\w*\)')
    $handlerCount += $handlerMatches.Count
}

# Also search httpapi package
$httpapiFiles = Get-ChildItem -Path "$Root\internal\interfaces\http" -Filter "*.go" -File -ErrorAction SilentlyContinue
foreach ($file in $httpapiFiles) {
    $content = Get-Content $file.FullName -Raw -ErrorAction SilentlyContinue
    $summaryMatches = [regex]::Matches($content, '//\s*@Summary')
    $summaryCount += $summaryMatches.Count
    $handlerMatches = [regex]::Matches($content, 'func\s+\(d\s+\*Deps\)')
    $handlerCount += $handlerMatches.Count
}

Write-Host "  @Summary annotations: $summaryCount"
Write-Host "  Handler methods: $handlerCount"

# 2. Per-package breakdown
Write-Host ""
Write-Host "--- Per-Package Coverage ---" -ForegroundColor Yellow
$packages = @("issue", "sprint", "version", "automation", "webhook", "dashboard", "knowledge", "pages", "intake", "workbench", "workspace", "auth")

foreach ($pkg in $packages) {
    $pkgPath = "$Root\internal\application\$pkg"
    if (Test-Path $pkgPath) {
        $pkgSummary = 0
        $pkgFiles = Get-ChildItem -Path $pkgPath -Filter "*.go" -File -ErrorAction SilentlyContinue
        foreach ($f in $pkgFiles) {
            $content = Get-Content $f.FullName -Raw -ErrorAction SilentlyContinue
            $pkgSummary += ([regex]::Matches($content, '//\s*@Summary')).Count
        }
        if ($pkgSummary -gt 0) {
            Write-Host "  $pkg : $pkgSummary @Summary annotations"
        }
    }
}

# 3. Coverage calculation
$coveragePercent = 0
if ($handlerCount -gt 0) {
    $coveragePercent = [math]::Round(($summaryCount / $handlerCount) * 100, 1)
}

Write-Host ""
Write-Host "=== Summary ===" -ForegroundColor Yellow
Write-Host "  Total @Summary: $summaryCount / $handlerCount handlers"
Write-Host "  Coverage: ${coveragePercent}% (target: ${TargetPercent}%)"

if ($coveragePercent -lt $TargetPercent) {
    Write-Host "  ⚠️  Coverage below target (${coveragePercent}% < ${TargetPercent}%)" -ForegroundColor Yellow
    Write-Host "  Add @Summary/@Description/@Tags/@Router to handler functions." -ForegroundColor Yellow
    exit 0  # Don't block, just warn
} else {
    Write-Host "  ✅ Swagger coverage OK (${coveragePercent}% >= ${TargetPercent}%)" -ForegroundColor Green
}
