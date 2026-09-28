---
title: 桌面端开发与发布
description: Tauri 2 与 Next.js standalone 桌面端运行及 GitHub 打包流程
---

# 桌面端开发与发布

## 运行方式

桌面端使用 Tauri 2，内置 Node 22、Next.js standalone 和 Go 后端。安装后无需单独安装 Node 或 Go。首次启动将 Next 资源解压至系统应用缓存目录，服务就绪后在窗口中打开 `http://127.0.0.1:39217`。固定前端来源用于保持 localforage、IndexedDB 和浏览器配置跨次启动连续可用；重复打开会聚焦已有窗口。前端端口被其他程序占用时会报错，不会改用随机端口。

关闭桌面程序会结束 Node 服务和由桌面端启动的 Go 服务；复用外部后端时不会停止外部进程。Tauri 意外退出导致管道关闭时，桌面端启动的服务也会退出。启动或运行失败会显示错误信息，日志保存在应用日志目录中的 `next-server.log`。

默认启动内置 Go 后端，由 Go 在 `127.0.0.1` 分配随机端口，完成数据库初始化和路由注册后通知桌面程序，再启动 Next.js 并注入 `API_BASE_URL`。数据库 `infinite-canvas.db`、自动生成的登录密钥 `.jwt-secret` 和 AI 日志保存在应用数据目录，重启和升级继续复用。服务启动超时或任一子进程异常退出会显示错误并清理配套进程。

如需使用已有的本地或远程后端，在启动桌面进程前设置 `API_BASE_URL`（例如 `http://127.0.0.1:8080`），桌面端将只连接指定后端。构建时的环境变量和仓库 `.env` 不会自动成为安装后应用的配置；需要从设置了环境变量的终端启动应用。

内置数据库首次为空，不会包含开发环境或其他服务器的提示词和渠道。默认管理员为 `admin` / `infinite-canvas`，可通过启动环境中的 `ADMIN_USERNAME`、`ADMIN_PASSWORD` 指定首次管理员；登录管理后台后同步提示词分类并配置渠道。未登录的画布项目和“我的素材”仍保存在桌面 WebView 本地；登录并启用账号同步后保存在当前后端，内置后端的数据仍在本机，跨设备同步需连接同一个远程后端。未登录直连的 AI API Key 保存在 WebView 本地，由前端请求 OpenAI 兼容接口，仍依赖上游允许 CORS；登录后的渠道同步及转译沿用项目现有规则。安装包不包含 canvas-agent。

首页提示词是可选的服务端内容，读取失败时会在展示区域提示，不阻断本地画布。后台公共配置启动请求失败时沿用现有本地配置。登录和云端功能仍需要真实可用的后端。

桌面版和浏览器版拥有不同的本地数据目录，不会自动共享未登录数据。清理 WebView 应用数据会删除本地项目；仅删除 Next 解压缓存不会清空画布存储。

## 开发与本地打包

需要 Node 22、Bun 1.4.2、Go 1.25、Rust stable，以及 Tauri 所需系统开发依赖：Windows 的 MSVC C++ Build Tools 和 WebView2，macOS 的 Xcode Command Line Tools，Linux 的 WebKitGTK 4.1 等。构建需要联网安装依赖，发布电脑无需安装开发工具。

在 `web` 目录运行：

```bash
bun install --frozen-lockfile
bun run tauri:dev
```

开发窗口访问 `http://127.0.0.1:3000`，开发模式仍需在仓库根目录单独执行 `go run .` 启动后端。本地生成安装包：

```bash
bun run tauri:build
```

构建入口 `scripts/desktop.mjs` 会读取根目录 `VERSION`，同步 `web/package.json`、`src-tauri/Cargo.toml` 和 Tauri 配置，再构建 Next、Go 后端，复制 `public`、`.next/static` 和启动脚本，将其压缩为 Tauri 资源。Node 和 Go 可执行文件按 Tauri 目标命名，作为 sidecar 一并打包；必须在目标操作系统和架构上使用原生 Node，不能用 x64 Node 打包 ARM64 应用。Go 构建显式设置与 Node 对应的 `GOOS` / `GOARCH` 并关闭 CGO。Next 的 `.env*` 文件与构建缓存不进入资源包。

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
4. `Release Desktop` 工作流构建全部平台；全部成功后创建 Release 并上传安装包，任一平台失败则暂停发布。

工作流使用 `web/bun.lock` 和 `src-tauri/Cargo.lock` 锁定依赖，版本必须与标签一致。手动运行时可留空标签，仅构建所选分支并保存 Actions artifacts；指定已存在的标签时才发布 Release，避免把分支名误当成版本发布。同一标签重跑会覆盖同名附件。

| 平台 | 安装包 |
| --- | --- |
| Windows x64 | `.exe`（NSIS）、`.msi` |
| macOS Intel x64 | `.dmg` |
| macOS Apple Silicon ARM64 | `.dmg` |
| Linux x64 | `.AppImage`、`.deb` |

当前未配置 Windows 开发者证书及 macOS Developer ID、公证，macOS 使用临时签名。系统可能要求用户确认运行来源；当前不会承诺已通过平台信任验证，也未接入自动更新。
