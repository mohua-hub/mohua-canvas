---
title: 数据库说明
description: 当前后端主要数据表与字段说明
---

# 数据库说明

本文档记录当前后端数据结构。应用无需账号，业务数据使用共享工作区 `default`；不再创建用户或积分表。既有数据库不执行旧表删除或数据迁移。

## 数据库

后端使用 GORM 管理数据库连接和表结构迁移。

仅使用本地 SQLite。源码运行时默认文件为 `data/infinite-canvas.db`，可通过 `DATABASE_DSN` 指定本地文件路径；桌面端保存在系统应用数据目录的 `infinite-canvas.db`。

当前启动时执行 `AutoMigrate`，自动维护以下表：

- `prompts`
- `agent_skills`
- `agent_skill_files`
- `assets`
- `settings`
- `video_tasks`
- `image_generation_logs`
- `canvas_image_tasks`
- `canvas_audio_tasks`
- `comfy_bridges`
- `comfy_bridge_requests`
- `canvas_projects`
- `workspace_configs`
- `storage_objects`

后续新增表时再同步补充本文档，未实际使用的规划表不提前写入。

### workspace_configs

共享工作区配置和同步数据表。工作区 ID 固定为 `default`，模型配置、自定义存储及同步数据保存在 text 字段中。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `workspace_id` | string | 工作区 ID，主键 |
| `model_config` | 大文本 | 模型与偏好配置 JSON，本地工作流的 `workflowChannels` 完整条目也暂存在这里；S3/R2 和 WebDAV 的自动同步开关分别为 `syncStorageConfig`、`syncWebDAVStorageConfig` |
| `storage_provider` | text | 自定义存储配置 JSON，内部结构为 `{ "s3": {...}, "webdav": {...} }`，两类配置可保留但不能同时启用 |
| `image_history` | text | 工作区图片历史同步数据 |
| `asset_data` | text | 工作区素材同步数据 |
| `created_at` | string | 创建时间 |
| `updated_at` | string | 更新时间 |

`storage_provider.s3` 保存 Endpoint、Region、Bucket、Access Key、Secret、公开域名和路径前缀；`storage_provider.webdav` 保存 WebDAV 地址、远程目录、用户名和密码/应用密码。自动同步开关不重复写入 Provider；后端下载和删除旧媒体时仍会读取已保存但已停用的 Provider。

### storage_objects

S3/R2 与 WebDAV 共用的媒体文件索引表，不保存画布、素材列表或生成记录。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | string | 文件 ID，前端存储 key 使用 `server:<id>` |
| `provider_id` | string | 创建文件时使用的 S3/R2 或 WebDAV Provider ID |
| `bucket` | string | S3/R2 Bucket；WebDAV 为空 |
| `object_key` | string | Provider 内相对对象路径，唯一索引 |
| `public_url` | string | S3/R2 可选公开地址；WebDAV 为空并通过 `/api/files/:id/content` 读取 |
| `mime_type` | string | 媒体 MIME 类型 |
| `bytes` | number | 文件字节数 |
| `width` | number | 预留字段，当前上传链路未写入，默认 `0` |
| `height` | number | 预留字段，当前上传链路未写入，默认 `0` |
| `sha256` | string | 文件内容摘要 |
| `direct` | boolean | 是否由浏览器直接上传至 WebDAV |
| `created_by` | string | 创建工作区 ID |
| `created_at` | string | 创建时间 |
| `deleted_at` | string | 预留字段；当前删除链路直接删除索引记录 |

### prompts

提示词表。用于保存公开提示词、内置 GitHub 系统提示词、分类和预览内容。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | string | 主键 |
| `title` | string | 标题 |
| `cover_url` | string | 封面图 |
| `prompt` | string | 提示词内容 |
| `tags` | json | 标签列表 |
| `category` | string | 分类标识 |
| `preview` | text | Markdown 展示内容，可包含文本、图片、视频链接等 |
| `created_at` | string | 创建时间 |
| `updated_at` | string | 更新时间 |

`github_url` 仅用于接口返回，不写入数据库。

### agent_skills

画布 Agent 可选择 Skill 表。系统预设和自定义 Skill 使用同一张表，通过来源区分；自定义 Skill 保存到共享工作区，后端不可用时保存在浏览器 localforage。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | string | 主键 |
| `owner_workspace_id` | string | 自定义 Skill 的工作区 ID；系统预设为空 |
| `source` | string | 来源：`system`、`user` |
| `name` | string | Skill 名称 |
| `description` | string | Skill 简介 |
| `cover_url` | text | Skill 封面图片地址 |
| `cover_storage_key` | string | 现有图片存储系统中的对象标识；直接填写图片链接时为空 |
| `content` | text | 完整 Markdown 或文本内容，最多 20000 字 |
| `enabled` | boolean | 是否启用；停用的系统预设不向画布返回 |
| `sort` | number | 系统预设排序值 |
| `created_at` | string | 创建时间 |
| `updated_at` | string | 更新时间 |

工作区接口读写 `source=user` 的自定义 Skill；设置页面维护 `source=system` 的系统 Skill。Skill 编辑后直接使用最新内容，不保存历史版本。

### agent_skill_files

系统预设 Skill 的附属目录和文本文件表。根 `SKILL.md` 不在本表重复保存，仍以 `agent_skills.content` 为唯一内容来源。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `skill_id` | string | 所属系统 Skill ID，与 `path` 组成联合主键 |
| `path` | string | Skill 包内安全相对路径 |
| `kind` | string | `folder` 或 `file` |
| `content` | text | 文件正文；文件夹为空，每个文件最多 20000 字 |
| `sort` | number | 同一目录内的显示顺序 |
| `created_at` | string | 创建时间 |
| `updated_at` | string | 更新时间 |

后台保存系统 Skill 时会在同一数据库事务中替换该 Skill 的附属目录记录；自定义 Skill 仍然只允许单文件内容，不写入本表。Agent 只在当前激活的系统 Skill 明确引用附属文件时按路径读取，不会把整个目录每轮注入模型上下文。

### assets

素材表。当前用于后台素材库。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | string | 主键 |
| `title` | string | 标题 |
| `type` | string | 素材类型：`text`、`image`、`video` 等 |
| `cover_url` | string | 封面图 |
| `tags` | json | 标签列表 |
| `category` | string | 分类标识 |
| `description` | string | 描述 |
| `content` | text | 文本或 Markdown 内容 |
| `url` | string | 图片、视频等媒体地址 |
| `created_at` | string | 创建时间 |
| `updated_at` | string | 更新时间 |

### video_tasks

视频生成任务表。后端创建视频任务后写入该表，后台轮询器每 5 秒统一查询未完成任务并更新进度、完成地址或失败详情；前端刷新、切换页面或关闭浏览器不会影响后端继续轮询。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | string | 主键，本地任务 ID，优先使用上游 task ID |
| `workspace_id` | string | 工作区 ID |
| `model` | string | 模型名称 |
| `channel_id` | string | 模型渠道 ID |
| `channel_name` | string | 模型渠道名称 |
| `source` | string | 任务来源：`canvas`、`workflow` |
| `source_id` | string | 来源内 ID，画布任务记录画布节点 ID |
| `upstream_task_id` | string | 上游任务 ID |
| `workflow_ref` | text | 仅新工作流任务使用的渠道/条目精确引用；旧视频任务为空 |
| `upstream_video_id` | string | 上游视频 ID，例如 Agnes 的 `video_...` |
| `status` | string | 状态：`queued`、`processing`、`completed`、`failed` |
| `progress` | number | 生成进度，0-100 |
| `seconds` | string | 视频秒数 |
| `size` | string | 视频尺寸 |
| `video_url` | text | 完成后的视频临时 URL |
| `error` | text | 失败摘要 |
| `error_detail` | text | 失败详情或最近一次轮询错误详情 |
| `request_body` | text | 创建任务时的请求摘要 |
| `response_body` | text | 创建任务时的响应摘要 |
| `last_response` | text | 最近一次状态响应摘要 |
| `created_at` | string | 创建时间 |
| `updated_at` | string | 更新时间 |
| `started_at` | string | 上游开始时间 |
| `completed_at` | string | 完成时间 |
| `last_polled_at` | string | 最近轮询时间 |

后台轮询器按 `status + created_at` 查询未完成任务；旧数据库中如果残留废弃列，不再参与代码查询。

### image_generation_logs

图片生成与创意工作流成果历史表。该表保存完整 JSON，并用独立字段做去重、软删除和查询。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | string | 主键，对应前端生成记录 ID |
| `workspace_id` | string | 工作区 ID，多用户数据隔离 |
| `task_id` | string | 图片任务 ID，可为空 |
| `image_id` | string | 图片结果 ID、存储 key 或 URL |
| `status` | string | 记录状态 |
| `payload_json` | text | 完整成果卡片 JSON。删除记录会清空该字段 |
| `created_at` | string | 创建时间 |
| `updated_at` | string | 更新时间 |
| `deleted_at` | string | 软删除时间，空字符串表示未删除 |

### canvas_image_tasks

画布图片生成任务表，用于画布节点生成恢复。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | string | 主键，本地任务 ID |
| `workspace_id` | string | 工作区 ID |
| `source` | string | 固定为 `canvas` |
| `source_id` | string | 画布来源 ID |
| `node_id` | string | 画布节点 ID |
| `model` | string | 模型名称 |
| `channel_id` | string | 模型渠道 ID |
| `workflow_ref` | text | 仅新工作流任务使用的渠道/条目精确引用；旧图片任务为空 |
| `status` | string | 状态：`queued`、`processing`、`completed`、`failed` |
| `progress` | number | 生成进度 |
| `prompt` | text | 提示词 |
| `generation_type` | string | `generation` 或 `edit` |
| `image_url` | text | 完成后图片 URL或第一张图片 URL |
| `image_urls` | JSON | 完成后全部图片 URL，第一项与 `image_url` 一致 |
| `storage_key` | string | 存储对象 key |
| `error` | text | 失败摘要 |
| `error_detail` | text | 失败详情 |
| `created_at` | string | 创建时间 |
| `updated_at` | string | 更新时间 |
| `started_at` | string | 开始时间 |
| `completed_at` | string | 完成时间 |

索引：`idx_canvas_image_tasks_workspace_source_node (workspace_id, source, source_id, node_id)`

### canvas_audio_tasks

画布音频生成任务表。只用于画布节点生成恢复，不影响原 `/audio/speech` 接口。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | string | 主键，本地任务 ID |
| `workspace_id` | string | 工作区 ID |
| `source` | string | 固定为 `canvas` |
| `source_id` | string | 画布来源 ID |
| `node_id` | string | 画布节点 ID |
| `model` | string | 模型名称 |
| `channel_id` | string | 模型渠道 ID |
| `workflow_ref` | text | 仅新工作流任务使用的渠道/条目精确引用；旧音频任务为空 |
| `status` | string | 状态：`queued`、`processing`、`completed`、`failed` |
| `progress` | number | 生成进度 |
| `prompt` | text | 提示词 |
| `audio_url` | text | 完成后音频 URL |
| `storage_key` | string | 存储对象 key |
| `error` | text | 失败摘要 |
| `error_detail` | text | 失败详情 |
| `created_at` | string | 创建时间 |
| `updated_at` | string | 更新时间 |
| `started_at` | string | 开始时间 |
| `completed_at` | string | 完成时间 |

索引：`idx_canvas_audio_tasks_workspace_source_node (workspace_id, source, source_id, node_id)`

### comfy_bridges

ComfyUI Bridge 设备表。每台可访问一处 ComfyUI 的独立程序注册一条设备；工作流配置仍保存在系统设置或工作区模型配置 JSON 中，不在此表重复保存。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | string | Bridge ID，主键 |
| `owner_scope` | string | `system` 或 `personal` |
| `owner_id` | string | 系统设备为 `system`，本地设备为工作区 ID |
| `name` | string | 设备显示名称 |
| `token_hash` | string | 专用 Token 的 SHA-256 摘要，唯一索引；不保存明文 Token |
| `enabled` | boolean | Token 有效标记；删除设备时整行移除 |
| `last_seen_at` | datetime | 最近一次心跳时间，用于判断在线状态 |
| `capabilities_json` | 大文本 | ComfyUI 地址、工作流目录及发现的工作流 ID/标题清单 |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |

索引：`idx_comfy_bridges_owner (owner_scope, owner_id)`。

### comfy_bridge_requests

Bridge 持久化请求队列表。普通执行请求由服务端按设备分配，Bridge 通过短租约领取和续租；ComfyUI 媒体只回传本机结果地址等小型元数据，服务端完成关联业务任务，删除设备或超过一小时硬截止的未完成请求标记失败，Bridge 确认不再需要重传后进入十分钟清理等待期。`inspect_workflow` 检查请求的硬截止为三十秒，调用方读取结果后立即删除；服务异常中断遗留的已完成检查请求由现有后台循环在三十秒后清理。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | string | 请求 ID，主键 |
| `bridge_id` | string | 目标 Bridge ID |
| `owner_scope` | string | `system` 或 `personal` |
| `owner_id` | string | 系统或用户归属 ID |
| `task_id` | string | 对应图片、视频或音频任务 ID；检查请求为空 |
| `kind` | string | 输出用途：`image`、`video`、`audio`，或按需读取工作流的 `inspect_workflow` |
| `status` | string | `pending`、`claimed`、`succeeded`、`failed` |
| `payload_json` | 大文本 | 执行请求的工作流、字段覆盖及媒体输入，或检查请求的工作流 ID/JSON |
| `result_json` | 大文本 | Bridge 回传的本机结果地址、文件名、MIME 等小型媒体元数据，或按需检查得到的工作流 JSON、字段和拓扑 |
| `error` | text | 失败原因 |
| `claimed_at` | datetime | 领取或重新领取时间 |
| `lease_token` | string | 当前领取者的租约令牌；续租、检查点和结果回传必须匹配 |
| `lease_expires_at` | datetime | 短租约到期时间；到期的 `claimed` 请求可以重新领取 |
| `checkpoint_json` | text | Bridge 已持久化的执行检查点，包括提交中状态与 ComfyUI `prompt_id`；提交状态不明确时按失败任务失败处理，不重复提交 |
| `completed_at` | datetime | 完成或失败时间 |
| `cleanup_ready_at` | datetime | 普通执行请求在业务任务完成且 Bridge 已确认结果后写入，超过十分钟由后台循环物理删除；检查请求完成时写入，正常由调用方立即删除，服务异常中断遗留记录在三十秒后由后台循环清理 |
| `expires_at` | datetime | 普通执行请求为一小时硬截止，`inspect_workflow` 为三十秒硬截止，均不因续租延长 |
| `created_at` | datetime | 创建时间 |
| `updated_at` | datetime | 更新时间 |

索引：`idx_comfy_bridge_queue (bridge_id, status)`、`task_id`、`lease_expires_at`、`cleanup_ready_at`、`expires_at`。

### canvas_projects

画布项目表。一条画布项目对应一行，完整项目 JSON 保存在 `project_data`，包含节点、连线、聊天会话、画布设置和视口；不拆节点表。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `workspace_id` | string | 所属工作区，与 `id` 组成主键 |
| `id` | string | 画布项目 ID |
| `project_data` | text | 完整 `CanvasProject` JSON |
| `created_at` | string | 项目创建时间 |
| `updated_at` | string | 项目更新时间 |
| `deleted_at` | string | 软删除时间，空字符串表示未删除；超过 7 天由启动时和每天定时任务物理清理 |

索引：`idx_canvas_projects_workspace_deleted_updated (workspace_id, deleted_at, updated_at)`、`idx_canvas_projects_deleted_at (deleted_at)`


### settings

系统配置表。`public` 放前端可读取的公开配置，`private` 放后端保存并由设置接口读取的渠道与存储配置；`agent-skills-initialized` 是默认 Skill 首次初始化标记，避免在设置中删除后被启动流程重新创建。配置值都用 JSON。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `key` | string | 主键：`public`、`private`、`agent-skills-initialized` |
| `value` | json | 配置内容 |
| `created_at` | string | 创建时间 |
| `updated_at` | string | 更新时间 |

`public.value` 常放前端展示和可公开读取的配置，例如模型列表、系统提示词等。  
`private.value` 常放渠道密钥、存储凭证、同步及日志开关等。
`private.value.storage.autoSyncAllAssets` 为“全部素材云端同步”开关，默认 `false`，控制前端新增媒体和生成结果的自动转存；使用现有配置 JSON 保存，不增加数据表。

当前系统设置接口会按后端结构体序列化和反序列化已知字段；数据库 JSON 中额外存在的旧字段会被忽略。

配置字段见[系统配置数据结构](system-settings.md)，包含模型、工作流、提示词、日志和存储配置，不含账号或积分项。
