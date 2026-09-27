---
title: 待测试
description: 当前版本已实现但仍需人工验证的变更项
---

# 待测试

## 桌面端发布

- Tauri 2 内置 Node 22、Next standalone、public 和静态资源；需在未安装 Node 的电脑上确认安装、启动和动态画布路由。
- 首次启动解压到可写缓存目录；需确认二次启动、升级后启动、目录包含中文或空格时的表现。
- 固定本地端口与单实例；需确认重启后未登录项目、素材、模型配置仍在，重复打开聚焦原窗口。
- 正常退出和意外结束时清理 Node 服务；需确认关闭后端口释放、端口占用时有明确错误、运行日志可定位。
- Windows 关闭 Tauri 原生拖放接管，保留画布 HTML5 拖放；需确认图片拖入、素材拖放、导出、下载和剪贴板。
- 复用项目标识生成 PNG、ICO、ICNS 图标；需确认各系统任务栏、安装器和应用图标。
- VERSION 同步前端、Cargo 与 Tauri；GitHub Actions 按标签构建 Windows x64、macOS Intel / Apple Silicon、Linux x64，全部成功后上传 Release。
- 桌面包不含 Go 后端；需连接自己的后端验证登录、云端同步、代理、OAuth 跳转等在线功能。前端直连接口仍依赖上游允许 CORS。
- 当前 Windows 未配置开发者证书，macOS 使用临时签名且未公证；安装确认流程需实机验证。
