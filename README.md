# 墨华画布

本地 AI 无限画布，使用 Electron、React 和 Go。支持文本、图片、视频、音频节点，连接上游文本生成作品，并管理项目、素材、模型渠道、提示词与 Agent Skill。

采用六边形架构和受信插件清单。活动画布由渲染进程管理，业务与持久化由本机 Go 服务负责，供应商调用通过独立适配器接入。项目与素材保存于本机工作空间，渠道密钥受系统安全存储保护。

需要 Node.js 22.12+、npm、Go 1.25+：

```powershell
npm ci
npm run dev
```

```powershell
npm run build
npm run desktop:package
```

Windows x64 安装包输出至 release/。本次重构使用新的 workspace-v1 数据目录，不读取旧项目数据。

[文档索引](docs/index.md) · [架构](docs/architecture/design.md) · [开发与打包](docs/development.md)
