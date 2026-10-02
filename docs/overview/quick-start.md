---
title: 快速开始
description: 用最少步骤把无限画布跑起来
---

# 快速开始

项目在本机运行，可使用 Windows 桌面安装包，也可从源码启动本机浏览器界面。

## 桌面端使用

从 [GitHub Releases](https://github.com/mohua-hub/mohua-canvas/releases) 下载 Windows x64 安装包并安装。应用内置 Node、Go 和 SQLite，打开后自动启动本地服务，无需另外安装开发工具。

首次启动后，进入 `/settings` 配置渠道、提示词和存储。画布项目和素材保存在本机。

## 源码本地运行

准备 Go 1.25、Node.js 22 和 Bun 1.4.2。在项目根目录执行：

```powershell
go run .
```

另开终端，在项目根目录执行：

```powershell
cd web
bun install --frozen-lockfile
bun run dev
```

访问 `http://127.0.0.1:3000`，直接打开 `/canvas` 使用画布，进入 `/settings` 管理设置，无需账号。

默认 Go 服务地址为 `http://127.0.0.1:8080`，SQLite 数据库为 `data/infinite-canvas.db`。需要调整本地端口或路径时，将 `.env.example` 复制为 `.env`；前端对应端口配置见[本地开发](../backend/local-development.md)。

## 首次使用建议

- 先打开右上角配置弹窗，填入自己的 `Base URL`、`API Key` 和模型名。
- 使用系统渠道时，进入 `/settings` 配置本地服务保存的模型与渠道；RunningHub 仅需一个 API Key。
- 如果需要提示词仓库内容，可进入 `/settings/prompts` 拉取或同步。

## 说明

- 画布项目和“我的素材”保存在浏览器或桌面 WebView 本地；本地服务连接且同步可用时，同时保存到本机工作区。浏览器版与桌面版的本地数据目录各自独立。
- AI 接口、RunningHub、提示词仓库及可选 S3/R2、WebDAV 存储需要连接相应外部服务；应用自身在本机运行。
- 本地直连模式下，AI API Key 保存在浏览器本地，并由前端直接请求 OpenAI 兼容接口。
