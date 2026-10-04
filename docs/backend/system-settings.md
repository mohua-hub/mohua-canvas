---
title: 系统配置数据结构
description: settings 表中 public 和 private 配置结构说明
---

# 系统配置数据结构

配置保存在 `settings` 表中的 `public`、`private` 两行，值为 JSON。设置入口为 `/settings`，读写接口为 `GET/POST /api/settings/config`；应用无需账号或积分。

`GET /api/settings` 返回通用配置与渠道摘要，不包含上游密钥。设置接口返回完整配置结构，但已保存的渠道密钥、存储密码等敏感值以空值返回；编辑时留空沿用已保存值。

## public.value

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `modelChannel.availableModels` | string[] | 可用模型列表 |
| `modelChannel.availableWorkflows` | string[] | 可用系统工作流标识 |
| `modelChannel.channels` | object[] | 渠道摘要，不含 API Key 或完整工作流 |
| `modelChannel.defaultModel` | string | 默认模型 |
| `modelChannel.defaultImageModel` | string | 默认图片模型 |
| `modelChannel.defaultVideoModel` | string | 默认视频模型 |
| `modelChannel.defaultTextModel` | string | 默认文本模型 |
| `modelChannel.systemPrompt` | string | 系统提示词 |
| `modelChannel.systemPrompts` | object | image、video、text、workflow、workflowAgent 的系统提示词 |
| `storage.mode` | string | 根据存储配置识别的模式 |
| `storage.allowCustomProvider` | boolean | 是否启用自定义 S3/R2 或 WebDAV |
| `storage.useGlobalProvider` | boolean | 是否启用全局存储 |

通用配置不再包含账号认证、用户权限或计费项。AI 渠道和密钥保存在本机 Go 服务的 `settings.private.channels`；所有 AI 请求均经本机 `/api/v1/*` 接口由 Go 服务转发，不支持浏览器直连渠道。

## private.value

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `channels` | object[] | 系统模型和工作流渠道 |
| `promptSync.enabled` | boolean | 定时同步远程提示词，默认开启 |
| `promptSync.cron` | string | 默认 `0 0 * * *` |
| `aiLog.cleanup` | object | enabled、retentionDays、cron；默认保留 14 天 |
| `storage.allowCustomProvider` | boolean | 自定义存储开关 |
| `storage.useGlobalProvider` | boolean | 全局存储开关 |
| `storage.autoSyncAllAssets` | boolean | 全部素材云端同步，默认关闭 |
| `storage.providers` | object[] | 全局 S3/R2 或 WebDAV 提供商 |
| `storage.capacityCheck` | object | enabled、cron；默认每 6 小时检查 |
| `storage.capacityLimitBytes` | number | 容量上限，默认 9 GiB |

渠道包含 id、protocol、name、baseUrl、apiKey、models、weight、timeout、enabled、remark。支持项目已有模型协议及 RunningHub、ComfyUI 工作流；同一模型的多个可用渠道按权重选择。

`workflows` 保存应用或工作流 ID、标题、用途、启用状态及字段映射。RunningHub 仅配置一个 `apiKey`，参数拉取、素材上传、任务提交和查询均使用该密钥；ComfyUI 使用 bridgeId、comfyUrl、workflowDir。Bridge 继续使用专用 Token。

全局存储保留 endpoint、region、bucket、accessKeyId、secretAccessKey、publicBaseUrl、pathPrefix，以及 WebDAV 的 username、password 等连接凭证。S3/R2 与 WebDAV 不能同时启用。

## 工作区配置

模型选择偏好、工作流集合及自定义存储保存在 `workspace_configs`，唯一工作区 ID 为 `default`。使用 `/api/v1/config` 读取，`/api/v1/config/model` 和 `/api/v1/config/storage` 保存。工作区配置不保存渠道凭据。

个人工作流集合使用 `localforage` 保存，连接后端时同步到共享工作区，并关联本机设置中已配置的渠道。S3/R2 和 WebDAV 的配置同步开关分别为 `syncStorageConfig`、`syncWebDAVStorageConfig`。AI 渠道密钥只存放在本机 Go 服务设置中；WebDAV 与 S3 的鉴权凭证继续按各自配置使用。
