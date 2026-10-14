# HTTP / Electron IPC v1

HTTP归Go，前缀/v1，127.0.0.1:0；所有路由含健康/媒体要求Bearer会话token，恒定时间比较，无CORS。JSON严格解析拒绝额外字段/多段JSON，≤8MiB；导入≤64MiB，X-Asset-Name为percent encoded名称。

成功{ok:true,data:T}，失败{ok:false,error:{code,message}}，固定中文提示，无内部错误。INVALID_INPUT400/UNAUTHORIZED401/NOT_FOUND404/CONFLICT409/UNSUPPORTED422/BUSY与RATE_LIMITED429/UPSTREAM_AUTH与UPSTREAM_FAILED502/TIMEOUT504/INTERNAL500。媒体支持Range。

| operation | /v1路由 | 输入→输出 |
| --- | --- | --- |
| health | GET /health | 无→{version:1} |
| listProjects | GET /projects | {}→ProjectSummary[] |
| createProject | POST /projects | {name}→ProjectView |
| openProject | GET /projects/{id} | {id}→ProjectView |
| saveProject | POST /projects/{id}/save | {name,expectedRevision,document}→ProjectView |
| deleteProject | DELETE /projects/{id} | {id}→null |
| listChannels | GET /channels | {}→ChannelView[] |
| configureChannel | POST /channels | {id?,name,pluginId,baseUrl,models,apiKey?}→ChannelView |
| deleteChannel | DELETE /channels/{id} | {id}→null |
| listPlugins | GET /plugins | {}→PluginView[] |
| listAssets | GET /assets | {}→AssetView[] |
| importAsset | POST /assets/import | 名称+字节→AssetView |
| openAsset | GET /assets/{id}/content | UUID→字节 |
| deleteAsset | DELETE /assets/{id} | {id}→null |
| submitGeneration | POST /jobs | {requestId,projectId,nodeId,channelId,model,kind,prompt}→JobView |
| listJobs | GET /jobs?projectId=UUID | {projectId}→最新100、所有未完成及快照节点引用的 JobView[]，去重 |
| getJob | GET /jobs/{id} | {id}→JobView |
| cancelJob | POST /jobs/{id}/cancel | {}→JobView |
| listLibrary | GET /library | {}→LibraryEntryView[] |
| saveLibraryEntry | POST /library | {id?,kind,name,content}→LibraryEntryView |
| deleteLibraryEntry | DELETE /library/{id} | {id}→null |

ProjectView=id/name/revision/document/createdAt/updatedAt，摘要无document。ChannelView无凭据引用，仅hasCredential。Asset/LibraryView同领域字段；JobView见业务契约；PluginView=id/type/version/apiVersion/capabilities/status，无内部路径。空列表[]。

window.mohua归preload，version1，仅invoke(operation,input)、importAsset({name,bytes:ArrayBuffer})、openRepository()。invoke按操作类型化；查询{}，资源{id}；saveProject包含id，cancelJob IPC包含id/HTTP为{}。import名字≤240字节；openRepository固定https://github.com/mohua-hub/mohua-canvas，无URL输入。

仅mohua:invoke/import-asset/open-repository三个channel；返回与HTTP相同结果。主进程验证操作/字段集合/ID/大小，Go权威校验；仅主窗口主frame可信URL。开发http://127.0.0.1:5173，生产打包renderer/index.html。服务不可用SERVICE_UNAVAILABLE，非法输入INVALID_INPUT，非可信来源FORBIDDEN。

生产HTML及主进程响应同时设置CSP：script仅self、connect none、img/media仅self或mohua-media（图片可data）、禁object/frame/base。开发允许Vite内联preamble及固定热更新WebSocket。生产sidecar工作目录为实际resources目录，不能使用虚拟app.asar路径。

mohua-media://asset/<UUID>仅GET/HEAD，无query/fragment/多余路径；附加token、转发Range、nosniff，无任意网络目标/文件路径/端口暴露。

组合根配置MOHUA_DATA_DIR绝对路径、MOHUA_SERVICE_TOKEN随机32字节hex、MOHUA_VAULT_KEY32字节base64、可选MOHUA_PLUGIN_CONFIG默认config/plugins.json。Go stdout就绪MOHUA_READY=<port>，其他固定日志。开发go run ./cmd/mohua，生产resources/bin/mohua-service；stdin关闭/SIGINT/SIGTERM停止，10秒内释放。
