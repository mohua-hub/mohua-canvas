# 开发与打包

需要Node.js22.12+、npm、Go1.25+，不需要Rust/Bun/Next.js。

```powershell
npm ci
npm run dev
```

Vite就绪后启动Electron及Go，无独立浏览器模式。数据userData/workspace-v1。

```powershell
npm run build
npm run desktop:package
```

build编译renderer和Go，desktop:package生成Windows x64 NSIS。package.json是版本唯一来源，重构保留当前版本为Unreleased，不创建tag/Release。签名用electron-builder的CSC_LINK/CSC_KEY_PASSWORD，无证书为未签名。无自动更新，旧Tauri更新文件删除。

Desktop package工作流仅手工构建上传artifact，不发Release；Review运行契约测试。本次按AGENTS.md不在本地执行语法检查、测试或构建。发布另行授权。

验证入口：npm test（Go用例、SQLite事务、HTTP鉴权、插件清单及渲染端命令行为）；npm run test:renderer只验证渲染端与桥接。手工验收包括新建/保存/重开项目、保存冲突保留编辑、节点连线/撤销、素材导入/播放、渠道密钥更新、四类生成/取消、Agent逐项确认、未保存关闭提示及打包后的sidecar启动。供应商调用需要自行配置有效渠道；自动测试不发起有成本调用。

入口cmd/mohua、electron/main/index.cjs、renderer/main.tsx、config/plugins.json。safeStorage不可用阻止启动，日志无key/token/供应商响应。
