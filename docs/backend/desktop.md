---
title: 桌面端开发与发布
description: Tauri 2 与 Next.js standalone 桌面端运行及 GitHub 打包流程
---

# 桌面端开发与发布

## 运行方式

桌面端使用 Tauri 2，内置 Node 22、Next.js standalone 和 Go 后端。安装后无需单独安装 Node 或 Go。首次启动将 Next 资源解压至系统应用缓存目录，服务就绪后在窗口中打开 `http://127.0.0.1:39217`。固定前端来源用于保持 localforage、IndexedDB 和浏览器配置跨次启动连续可用；重复打开会聚焦已有窗口。前端端口被其他程序占用时会报错，不会改用随机端口。

关闭桌面程序会结束内置 Node 和 Go 服务。Tauri 意外退出导致管道关闭时，桌面端启动的服务也会退出。启动或运行失败会显示错误信息，日志保存在应用日志目录中的 `next-server.log`。

桌面端始终启动内置 Go 后端，由 Go 在 `127.0.0.1` 分配随机端口，完成 SQLite 初始化和路由注册后通知桌面程序，再启动 Next.js 并注入 `API_PORT`。Next.js 仅代理到本机该端口。数据库 `infinite-canvas.db` 和 AI 日志保存在应用数据目录，重启和升级继续复用。服务启动超时或任一子进程异常退出会显示错误并清理配套进程。

内置数据库首次为空，与源码开发使用的数据目录独立。直接打开 `/settings` 同步提示词分类并配置渠道。画布项目和“我的素材”保存在桌面 WebView 本地；本地服务连接且同步可用时保存到本机共享工作区。本地直连的 AI API Key 保存在 WebView 本地，由前端请求 OpenAI 兼容接口，仍依赖上游允许 CORS；同步后的本地渠道可由本机后端转译。

桌面安装包会打入 ComfyUI Bridge 的 Windows x64 与 Linux x64 / ARM64 下载文件。Bridge 在应用所在电脑单独启动，通过当前本地页面地址连接本机服务；ComfyUI 地址仍可指向该电脑可访问的 ComfyUI 实例。安装包不包含 canvas-agent 服务。

首页提示词是可选的本机服务内容，读取失败时会在展示区域提示，不阻断本地画布。后端公共配置启动请求失败时沿用现有本地配置。设置与后端任务需要本地 Go 服务正常运行，外部 AI 与可选对象存储按原有配置连接。

桌面版和浏览器版拥有不同的本地数据目录，不会自动共享本地数据。清理 WebView 应用数据会删除本地项目；仅删除 Next 解压缓存不会清空画布存储。

## 开发与本地打包

桌面安装包仅支持 Windows x64。构建需要 Node 22、Bun 1.4.2、Go 1.25、Rust stable、Windows MSVC C++ Build Tools 和 WebView2；macOS/Linux 的桌面开发运行仍需对应的 Tauri 系统依赖，但不支持构建安装包。构建需要联网安装依赖，发布电脑无需安装开发工具。

### 桌面自动更新

应用内更新仅适用于 Windows NSIS `.exe`；Windows MSI 通过 GitHub Release 手动更新。v0.1.4 及更早版本没有更新器，需先手动安装第一个带更新器的 NSIS 版本，此后才可在版本更新窗口下载、安装并重启。

发布前生成 Tauri 更新密钥对，并将私钥文件内容保存到 GitHub 仓库 Settings → Secrets and variables → Actions，Secret 名称为 `TAURI_SIGNING_PRIVATE_KEY`。当前机器生成的私钥保存在 `%LOCALAPPDATA%\MohuaCanvas\tauri-updater.key`；公钥已写入 Tauri 配置，私钥不能提交或上传为 Release 附件。Release workflow 只为 Windows NSIS 包签名并写入 `latest.json`，Windows MSI 仍通过 Release 手动更新。签名密钥丢失或更换后，已有安装无法验证新密钥签署的更新；更换密钥需要用户手动安装新版本。

在 `web` 目录运行：

```bash
bun install --frozen-lockfile
bun run tauri:dev
```

开发窗口访问 `http://127.0.0.1:3000`，开发模式仍需在仓库根目录单独执行 `go run .` 启动后端。本地生成安装包：

```bash
bun run tauri:build
```

构建入口 `scripts/desktop.mjs` 会读取根目录 `VERSION`，同步 `web/package.json`、`src-tauri/Cargo.toml` 和 Tauri 配置，再构建 Go 后端、ComfyUI Bridge 与 Next，复制 `public`、`.next/static` 和启动脚本，将其压缩为 Tauri 资源。Node 和 Go 可执行文件按 Tauri 目标命名，作为 sidecar 一并打包；必须在目标操作系统和架构上使用原生 Node，不能用 x64 Node 打包 ARM64 应用。Go 构建显式设置与 Node 对应的 `GOOS` / `GOARCH` 并关闭 CGO。Next 的 `.env*` 文件与构建缓存不进入资源包。

生成目录：

- `src-tauri/binaries/`：当前平台的 Node 和 Go 后端。
- `src-tauri/resources/`：Next 压缩包和构建标识。
- `src-tauri/target/release/bundle/`：安装包。

这些目录不提交至 Git。图标沿用现有项目标识，源文件在 `desktop/icon.svg`；仓库根目录执行以下命令可以重新生成图标：

```bash
node web/node_modules/@tauri-apps/cli/tauri.js icon desktop/icon.svg --output src-tauri/icons
```

## GitHub 发布

1. 将 `CHANGELOG.md` 的 `Unreleased` 内容整理到新版本标题下，保留空 `Unreleased`。
2. 更新根目录 `VERSION`，执行 `node scripts/desktop.mjs sync` 同步版本。
3. 提交发布内容，创建与 `VERSION` 一致的 `vX.Y.Z` 标签并推送。
4. `Release Desktop` 工作流只构建 Windows x64 的 NSIS 与 MSI 安装包，为 NSIS 包签名并生成 `latest.json`；构建成功后创建 Release 并上传安装包与更新清单。

工作流使用 `web/bun.lock` 和 `src-tauri/Cargo.lock` 锁定依赖，版本必须与标签一致。手动运行时可留空标签，仅构建所选分支并保存 Actions artifacts；指定已存在的标签时才发布 Release，避免把分支名误当成版本发布。同一标签重跑会覆盖同名附件。

| 平台 | 安装包 |
| --- | --- |
| Windows x64 | `.exe`（NSIS）、`.msi` |

Windows 尚未配置开发者证书，系统可能显示来源确认。Tauri 更新签名只验证更新包完整性，不替代平台证书签名。
