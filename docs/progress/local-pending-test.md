# 本地运行调整待测试

项目改为本机运行，移除 Docker、Compose、Render 部署文件及容器专用代码。尚未执行语法检查、构建或测试。

- 按快速开始分别启动 Go 与 Next.js，确认仅通过 `127.0.0.1:8080` 和 `127.0.0.1:3000` 访问本机服务；默认 SQLite 文件为 `data/infinite-canvas.db`，设置、素材、提示词与任务正常读写。
- 根目录 `.env` 使用 `PORT` 和 `DATABASE_DSN` 调整本地端口与数据库路径；前端 `web/.env.local` 的 `API_PORT` 与 Go 端口一致时，页面 API 请求正常。数据库仅使用 SQLite。
- Windows 桌面端始终启动内置 Go 服务，收到随机本地端口后启动 Next.js；确认设置、RunningHub、任务查询和重启恢复正常，关闭应用时两个子进程退出，数据库保留在应用数据目录。
- ComfyUI Bridge 启动命令使用当前本地页面地址；在同一台电脑启动 Bridge，确认设备发现、字段拉取和工作流运行正常。
- 画布项目与“我的素材”继续保存在浏览器或桌面 WebView 本地，本机服务可用时写入本机工作区。确认可选 AI 接口、RunningHub、提示词同步、S3/R2、WebDAV 功能保持可用。
- README、快速开始与文档索引仅提供本地运行、桌面开发和发布入口；桌面打包仍保留 Next.js standalone、静态资源复制及 ComfyUI Bridge 下载产物。

原 `todo.md` 与 `pending-test.md` 已由用户删除，保留删除状态；本文件记录此次实际可测试变更，不代表已通过验证。
