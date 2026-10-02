---
title: 接口响应约定
description: 业务接口统一响应结构与前端处理约定
---

# 接口响应约定

后端业务接口统一返回 JSON：

```json
{
  "code": 0,
  "data": {},
  "msg": "ok"
}
```

- `code`: 业务状态码，`0` 表示成功，非 `0` 表示失败。
- `data`: 业务数据。失败时通常为 `null`。
- `msg`: 响应消息。成功默认为 `ok`，失败时放错误原因。

前端请求逻辑以 `code` 判断业务是否成功。当前后端业务失败也会返回 HTTP 200，前端不要只依赖 HTTP 状态码判断结果。

接口连接失败、服务不可达、返回体不是约定 JSON 时，前端按网络或接口异常处理。

## 设置与工作区

应用接口无需登录或应用 Token。设置使用 `GET/POST /api/settings/config`，渠道读取和测试使用 `/api/settings/config/channel-models` 与 `/api/settings/config/channel-test`。提示词、Skill、素材及 AI 日志使用 `/api/settings/*`。

工作区模型与存储配置使用 `/api/v1/config`、`/api/v1/config/model` 和 `/api/v1/config/storage`；同步数据使用 `/api/v1/data/*`、画布项目和生成记录接口。后端使用唯一工作区 `default`，返回 `syncCapabilities.workspaceData` 等同步能力。上游模型 API Key 与 ComfyUI Bridge 的专用 Token 继续使用原协议。

## AutoDL 渠道

- `POST /api/ai/autodl/workflows` 使用上述业务响应结构。请求为 `{ baseUrl, workflowId? }`，不传上游 API Key；不带工作流 ID 返回列表，带 ID 返回详情。元数据包含 `uuid`、`name`、`kind` 和详情的 `input_rules`，不返回上游内部节点映射。
- AutoDL 视频继续使用现有 `/api/v1/videos` 创建、查询及任务响应结构。
- `/api/v1/audio/speech` 是 AI 协议代理入口。AutoDL 成功时返回 `{ provider: "autodl", audio_url, mime_type }`；失败仍使用现有 `Fail` 业务结构。其他语音协议的响应保持原样。
- 现有 `/api/v1/canvas/audio-tasks` 消费 AutoDL 的上游音频 URL，返回既有任务结构，不新增结果下载或转存。

## RunningHub 画布接口

以下接口无需账号，沿用 `{ code, data, msg }` 响应结构：

| 接口 | 请求 | `data` |
| --- | --- | --- |
| `GET /api/v1/runninghub/collection` | 查询参数 `scope=all`、`scope=personal` 或 `scope=system` | 集合数组，每项包含 `ref`、`channelName` 和条目摘要 `entry` |
| `POST /api/v1/runninghub/entry` | `WorkflowRef`：`scope`、`channelId`、`kind`、`workflowId` | 条目名称、用途及启用的字段配置；不返回渠道密钥、完整工作流 JSON 或图结构 |
| `POST /api/v1/runninghub/tasks` | `ref`、`expectedCapability`、`prompt`、`fieldValues`、参考素材及画布任务关联字段 | `id`、`status`、`progress`，完成后包含 `urls`，失败后包含 `error` |
| `GET /api/v1/runninghub/tasks/:id` | 共享工作区任务 ID | 同上述任务结构 |

本地集合读取已同步到共享工作区的 RunningHub 渠道与条目，内部 scope 保留 `personal`。系统集合遵循条目启停和公开工作流名单；`scope=all` 合并本地与系统条目，不依赖普通模型的渠道模式。集合只返回摘要，选择后通过详情接口获取字段。提交与查询接口拒绝非 RunningHub 条目或任务，任务记录沿用现有工作流服务；项目不再扣减积分。

画布普通模型选择器不再把 RunningHub 条目列为模型；工具栏文字按钮和图片、视频、音频节点的模型菜单均提供 RunningHub 集合入口，选择后插入独立参数节点。普通模型列表为空时，画布菜单仍可进入集合，不会直接跳到配置。`fieldValues` 使用 `field:<nodeId>:<fieldName>` 标识单个字段；数值范围、枚举、必填与随机值由服务端按保存配置校验。素材按配置的序号绑定，上传后的 `fileName` 作为对应节点字段值，不能将其直接当作外链。

RunningHub 上游接口依据[官方文档](https://www.runninghub.cn/runninghub-api-doc-cn/)使用：

- AI 应用字段：`GET /api/webapp/apiCallDemo`，解析 `nodeInfoList` 的描述、`fieldData` 枚举与数值范围。
- 工作流字段：`POST /api/openapi/getJsonApiFormat`，从 API JSON 中提取静态输入，由渠道配置决定启用字段。
- 应用运行：`POST /task/openapi/ai-app/run`；工作流运行：`POST /task/openapi/create`。
- 素材上传：`POST /openapi/v2/media/upload/binary`，使用 Bearer 鉴权和 multipart `file`；与参数拉取、任务提交及查询共用渠道的唯一 API Key。
- 结果查询：`POST /openapi/v2/query`，使用 Bearer 鉴权，请求体仅含 `taskId`，处理 `QUEUED`、`RUNNING`、`SUCCESS`、`FAILED` 与 `results[].url`。
