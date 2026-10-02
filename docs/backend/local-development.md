---
title: 本地开发
description: 前后端分开启动时的本地开发方式
---

# 本地开发

准备 Go 1.25、Node.js 22 和 Bun 1.4.2，前后端分别在本机启动。

## 1. 准备环境变量

默认配置可以直接启动；需要调整本地端口或数据库路径时，在项目根目录执行：

```powershell
Copy-Item .env.example .env
```

默认配置下：

- 后端仅监听 `127.0.0.1:8080`
- 前端仅监听 `127.0.0.1:3000`
- SQLite 数据库是 `data/infinite-canvas.db`

## 2. 启动后端

在仓库根目录执行：

```powershell
go run .
```

后端会读取根目录 `.env`，使用本地 SQLite，并监听：

```text
http://127.0.0.1:8080
```

## 3. 启动前端

在 `web` 目录执行：

```powershell
bun install --frozen-lockfile
bun run dev
```

前端默认访问：

```text
http://127.0.0.1:3000
```

前端 `/api/*` 请求固定代理到本机 Go 服务，默认端口为 `8080`。如需修改端口，在根目录 `.env` 设置 `PORT=8081`，并在 `web/.env.local` 设置 `API_PORT=8081`，然后重启前后端。无需配置后端 URL。

页面入口：`/canvas` 为画布，`/settings` 为设置；无需注册或登录。

## 常见场景

- 改画布、页面和交互：主要看 `web/`
- 改接口、业务逻辑和数据库：主要看仓库根目录下的 Go 代码
- 改文档内容：主要看 `docs/`
