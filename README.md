# 无限画布

支持图片、视频、音频生成及 RunningHub、ComfyUI 工作流的 AI 创作画布。

项目在本机运行，支持 Windows 桌面端和本机浏览器使用。无需注册或登录；画布与素材保存在浏览器或桌面 WebView 本地，本机 Go 服务使用 SQLite 保存工作区、配置和任务。渠道、提示词、Skill、素材、日志和存储在 `/settings` 设置页管理。

桌面端可从 [GitHub Releases](https://github.com/mohua-hub/mohua-canvas/releases) 下载安装包，内置前后端，无需安装开发工具。

源码运行需要 Go 1.25、Node.js 22 和 Bun 1.4.2。在项目根目录启动本地服务：

```powershell
go run .
```

另开终端启动前端：

```powershell
cd web
bun install --frozen-lockfile
bun run dev
```

访问 `http://127.0.0.1:3000`。默认数据库为 `data/infinite-canvas.db`；需要调整本地端口或路径时，复制 `.env.example` 为 `.env`。

[快速开始](docs/overview/quick-start.md) · [本地开发](docs/backend/local-development.md) · [桌面端开发与发布](docs/backend/desktop.md) · [文档索引](docs/index.md)
