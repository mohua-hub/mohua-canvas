# 业务与存储契约v1

ID为UUID，时间服务生成UTC RFC3339。名称1–120 UTF-8字节，模型/插件ID≤200，prompt/content/Skill≤32000。拒绝未知字段、非有限数值和不支持版本/枚举。Domain无JSON/SQL标签；Application输入输出、HTTP DTO、SQL record、供应商DTO分别定义并映射，无用户/租户字段。

## 项目

Project=id/name/revision/document/createdAt/updatedAt，revision从1开始。CanvasDocument=version(1)/nodes/edges/viewport；≤2000节点/4000边、JSON≤8MiB。节点=id/kind(text,image,video,audio)/x/y/title/content/prompt/channelId/model/assetId/jobId，引用可空，坐标绝对值≤1e6。边=id/source/target，不许自环/悬空/重复端点对/环。viewport=x/y/zoom(0.1–4)。引用必须存在，任务属于项目/节点，素材类别匹配kind，content为纯文本。

ProjectStore/ListProjects摘要，CreateProject(name)空文档，OpenProject(id)快照，SaveProject(id,name,expectedRevision,document)权威校验+CAS，冲突CONFLICT且保留本地编辑。DeleteProject拒绝未完成任务，删除项目及任务历史，不级联媒体。SQL适配器执行事务。

## 渠道与凭据

Channel=id/name/pluginId/baseUrl/models/credentialRef；models≤50个{name,kind}，无重复pair。baseUrl绝对http/https，无userinfo/query/fragment；HTTP仅允许localhost/loopback，HTTPS可使用供应商地址；插件决定能力。

ConfigureChannel输入id可空/name/pluginId/baseUrl/models/apiKey可空；更新空key保留，新建key1–8192字节。ChannelView无key/credentialRef，仅hasCredential。配置用例串行；新凭据引用写入后保存渠道，失败清新引用，成功清旧引用；事务拒绝修改存在未完成任务的渠道，删除拒绝项目/未完成任务引用。

ChannelStore=List/Get/Save/Delete；CredentialVault=Put(ref,value)/Get(ref)/Delete(ref)。模型经vault获取key，用例仅传credentialRef。

## 素材

Asset=id/name/mimeType/bytes/createdAt。嗅探PNG/JPEG/WebP/GIF/MP4/WebM/MP3/WAV/OGG/FLAC，≤64MiB。名称仅显示，不接受路径或权威客户端MIME。MediaStore=Put(id,bytes)/Open(id)/Remove(id)，UUID键，目录组合根注入。

ImportAsset校验→落盘→元信息，数据库失败回滚文件。List/Open按元信息/ID。Delete拒绝项目/任务引用。供应商媒体同样落盘，远程URL不进公共DTO。

## 生成任务

Job=id/requestId/inputHash/projectId/nodeId/channelId/model/kind/prompt/status/text/assetId/errorCode/createdAt/updatedAt；status=queued/running/succeeded/failed/cancelled。Submit输入requestId/projectId/nodeId/channelId/model/kind/prompt；UUID幂等键，内容不同CONFLICT。服务校验项目/节点/模型能力/prompt，持久化后返回queued。并行2，排队32，超额BUSY，10分钟超时，无自动有成本重试。

ModelProvider.Generate(GenerationRequest,Channel)→GenerationResult{Text,Media,MediaName}，Media字节。分类错误UPSTREAM_AUTH/RATE_LIMITED/UPSTREAM_FAILED/TIMEOUT/UNSUPPORTED。取消持久化后成功，晚到结果不覆盖cancelled；写入失败不宣称完成。重启queued/running→failed/INTERRUPTED。

GeneratedAssets端口归生成用例：ImportGenerated(name,bytes)→Asset，Delete(id)→无；素材用例实现，同ImportAsset验证、落盘和回滚边界。生成模块仅依赖此端口，不依赖素材服务具体类型。

JobStore=GetByRequest(requestId)/CreateIdempotent(job,expectedChannel,admit)/Get/List/SetStatus(expectedStatuses)/InterruptPending，状态CAS。GetByRequest先返回已有幂等回执，不要求已完成任务的节点仍存在；内容不符CONFLICT。CreateIdempotent在事务中优先回放已有回执，再判断admit容量许可，BUSY不创建记录；并校验项目节点及渠道快照仍与expectedChannel一致，不符CONFLICT。Application串行保留最多34个未完成任务（2执行+32等待），不对供应商重试。JobView只含id/projectId/nodeId/kind/status/text/assetId/errorCode/createdAt/updatedAt，无prompt/requestId/hash/上游ID。

OpenAI兼容：text→chat/completions，image→images/generations，audio→audio/speech，video→videos multipart/轮询/videos/{id}/content。Gemini：text/image→generateContent。插件无业务任务写入。结果URL仅HTTPS，无重定向，DNS拒绝非公网地址；显式供应商endpoint受信，返回下载URL不受信。

供应商普通JSON响应≤4MiB，图片base64响应≤90MiB，解码后的媒体仍受64MiB及类型校验。正文不进入日志或公共错误。

## 创作库

LibraryEntry=id/kind(prompt,skill)/name/content/createdAt/updatedAt，LibraryStore/List/Save→已保存LibraryEntry/Delete。更新保留createdAt，返回数据库确认的时间。prompt填节点，skill只补充Agent指令，无代码执行或权限提升；无远程抓取/scope。

## SQLite v1

workspace.db user_version=1；非空版本不符拒绝，无迁移。WAL/foreign_keys/busy_timeout5000、单连接。表：

- projects(id,name,revision,document_json,created_at,updated_at)
- channels(id,name,plugin_id,base_url,models_json,credential_ref)
- credentials(ref,nonce,ciphertext)
- assets(id,name,mime_type,bytes,created_at)
- jobs(id,request_id UNIQUE,input_hash,project_id,node_id,channel_id,model,kind,prompt,status,text,asset_id,error_code,created_at,updated_at)
- library_entries(id,kind,name,content,created_at,updated_at)

内部JSON record不得直接返回API。标准库AES-GCM随机nonce；Electron safeStorage加密32字节主密钥到vault-key.bin，环境传Go，仅内存。safeStorage不可用阻止启动。媒体media/<UUID>，退出不删数据。
