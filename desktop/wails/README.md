# Ydsz Plane 桌面客户端

基于 **Wails v3** 的原生桌面应用（Windows / macOS / Linux），复用现有 Go 后端与 Vue 3 前端。

## 特点

- **轻量**：不打包 Chromium，使用系统原生 WebView（Windows WebView2 / macOS WKWebView / Linux WebKitGTK）
- **小体积**：安装包约 15-20MB（对比 Electron 的 80-150MB）
- **Go 原生后端**：直接复用项目现有的 Go 代码与服务
- **类型安全绑定**：`wails3 dev/build` 自动生成 TypeScript 绑定
- **系统能力**：原生通知、窗口控制、文件拖放、自动更新
- **代码签名与公证**：支持 Windows Authenticode 与 macOS notarytool

## 目录结构

```
desktop/wails/
├── main.go                    # 应用入口（窗口创建 + 服务注册）
├── common.go                  # 版本号 + 日志工具
├── service_platform.go        # 平台信息服务
├── service_notification.go    # 原生通知服务
├── service_window.go          # 窗口控制服务
├── service_updater.go         # 自动更新服务
├── wails.json                 # Wails v3 配置
├── Taskfile.yml               # 构建任务（含签名）
├── .signing.env.example       # 签名配置模板
├── .gitignore                 # 保护签名凭证
├── frontend/                  # 前端目录（软链到 ../../web）
│   └── dist/                  # 构建产物（嵌入二进制）
├── build/                     # 构建资源
│   ├── entitlements.plist     # macOS 权限声明
│   ├── certs/                 # 签名证书（.gitignore 排除）
│   └── icons/                 # 应用图标
└── scripts/                   # 签名与验证脚本
    ├── sign-macos.sh          # macOS 签名 + 公证
    ├── sign-windows.ps1       # Windows 签名
    └── verify-signing.sh      # 跨平台签名验证
```

## 前置要求

| 平台 | 要求 |
|------|------|
| Windows | WebView2 Runtime（Win10/11 自带）；Windows SDK（签名） |
| macOS | Xcode Command Line Tools（`xcode-select --install`）；Developer ID 证书（签名） |
| Linux | WebKitGTK 6.0（Ubuntu 24.04+ 或 Debian 13+） |

安装 Wails CLI：

```bash
go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0
wails3 doctor   # 检查环境
```

## 开发

```bash
# 1. 链接前端目录（首次）
cd desktop/wails
ln -s ../../web frontend        # macOS/Linux
# Windows: mklink /J frontend ..\..\web

# 2. 启动开发（热重载）
wails3 dev
```

## 构建

```bash
# 当前平台
wails3 build

# 指定平台
wails3 build -platform windows/amd64
wails3 build -platform darwin/universal
wails3 build -platform linux/amd64

# 使用 Taskfile（推荐）
task build:windows
task build:macos
task build:linux
```

输出位于 `bin/`。

---

## 代码签名与公证

### 1. 配置凭证

```bash
cp .signing.env.example .signing.env
# 编辑 .signing.env 填写真实凭证
```

**安全提示**：`.signing.env` 已在 `.gitignore` 中，切勿提交真实凭证。

### 2. macOS 签名与公证

**获取证书**：在 Apple Developer 网站申请 **Developer ID Application** 证书，双击导入 Keychain。

**配置**（`.signing.env`）：

```bash
MACOS_SIGN_ENABLED=true
MACOS_SIGN_IDENTITY="Developer ID Application: Your Name (TEAM_ID)"
MACOS_TEAM_ID="YOUR_TEAM_ID"

MACOS_NOTARY_ENABLED=true
MACOS_NOTARY_APPLE_ID="your@example.com"
MACOS_NOTARY_PASSWORD="@keychain:AC_PASSWORD"   # 推荐：存 Keychain
MACOS_NOTARY_TEAM_ID="YOUR_TEAM_ID"
```

**生成专用密码**（推荐用 Keychain 存储）：

```bash
xcrun notarytool store-credentials "AC_PASSWORD" \
  --apple-id "your@example.com" \
  --team-id "YOUR_TEAM_ID"
```

**执行**：

```bash
task sign:macos
```

**流程**：
1. `codesign --deep --options runtime` 深度签名（含 Hardened Runtime）
2. `notarytool submit --wait` 提交 Apple 服务器公证
3. `stapler staple` 绑定公证票据（离线可用）
4. `spctl --assess` 验证 Gatekeeper 认可

### 3. Windows 签名

**获取证书**：从 CA 购买代码签名证书（如 DigiCert、Sectigo），导出为 `.pfx`。

**配置**（`.signing.env`）：

```bash
WINDOWS_SIGN_ENABLED=true
WINDOWS_SIGN_CERT="build/certs/ydsz-plane.pfx"
WINDOWS_SIGN_PASSWORD="your-password"
WINDOWS_SIGN_TIMESTAMP="http://timestamp.digcert.com"
```

**执行**：

```bash
task sign:windows
```

**流程**：
1. 优先使用 `signtool`（Windows SDK）签名 + 时间戳
2. 无 signtool 时回退 `osslsigncode`（跨平台）
3. `Get-AuthenticodeSignature` 验证

**交叉签名**（Linux/macOS 上构建 Windows）：

```bash
WINDOWS_OSSL_ENABLED=true
WINDOWS_OSSL_CERT="build/certs/ydsz-plane.pem"
WINDOWS_OSSL_KEY="build/certs/ydsz-plane.key"
```

### 4. 验证签名

```bash
task sign:verify
# 或直接
bash scripts/verify-signing.sh bin/ydsz-plane
```

### 5. 一键签名打包

```bash
task package:windows:signed   # 构建 + 签名 + 验证
task package:macos:signed     # 构建 + 签名 + 公证 + 验证
```

---

## 前端桥接

前端通过自动生成的绑定调用 Go 服务（`web/src/lib/desktop.ts`）：

```ts
import { initDesktop, notify, openExternal, isDesktop } from "@/lib/desktop";

await initDesktop();

if (isDesktop.value) {
  await notify("任务已完成", "您有 3 个待处理事项");
  await openExternal("https://github.com/njydsz/ydsz-plane");
}
```

浏览器环境下自动降级为 Web API。

## 故障排查

| 问题 | 解决方案 |
|------|---------|
| `frontend/dist` 不存在 | 先运行 `pnpm build`（在 web/ 目录） |
| WebView2 缺失 | 从 Microsoft 下载安装 |
| Linux 构建失败 | 安装 `libwebkitgtk-6.0-dev` 或用 `-tags gtk3` |
| 绑定未生成 | 重新运行 `wails3 dev` 或 `wails3 build` |
| macOS 签名失败 | 检查证书是否过期、Keychain 是否解锁 |
| macOS 公证超时 | Apple 服务器繁忙，稍后重试；`--wait` 会自动轮询 |
| Windows 签名失败 | 确认 `.pfx` 密码正确；时间戳服务器可达 |
