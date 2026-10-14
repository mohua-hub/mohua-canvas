# 画布与Agent v1

CanvasCapabilityPort归renderer画布模块，组合根注入UI/Agent，是实时状态业务更新入口。React Flow事件在画布内解释；selected/assetURL/loading/callback不保存。

端口getSummary→projectId/name/nodes/edges；createNode({kind,title,content?,prompt?,x,y})→ID；editNode({id,title,content,prompt,channelId,model})；connectNodes({source,target})→ID；removeNodes({ids})；attachAsset({nodeId,asset:AssetView})；applyJob({nodeId,job:JobView})；undo/redo。校验ID/文本/位置/连接，失败不改状态，删除级联边；上游祖先文本≤32000字节，服务独立校验。切换项目先保存或明确放弃。

Agent用服务text任务获取回答，独立text节点作为锚点。系统提示要求JSON{reply:string,commands:ToolCall[]}，≤8个{tool,args}：

| 工具 | 严格参数 / 行为 |
| --- | --- |
| create_node | {kind,title,content?,prompt?,x,y}→createNode |
| connect_nodes | {source,target}→connectNodes |
| delete_nodes | {ids} 1–20UUID→removeNodes |

拒绝额外字段/未知工具/非法边界，INVALID_TOOL。模型/Skill/节点不可信；逐项显示、用户确认，无自动执行；已执行不可重复，Agent不隐式提交媒体任务。手动生成是明确授权，状态来自JobView。

会话仅内存，切换项目销毁。Skill仅选用指令，无代码/权限提升。无外部MCP或旧Codex入口。

Agent上下文是画布摘要：最多40个节点，每个节点content/prompt最多200字符，只包含摘要节点间的最多80条边，并报告总量与截断；完整请求仍限制32000 UTF-8字节，超过时提示精简用户请求或Skill后发送，不隐式裁剪用户指令。

生成成功应用到当前节点后清空其jobId，表示该结果已被画布消费，任务历史仍保留在服务中。失败/取消保留jobId用于查看状态；重新打开项目时仅消费仍被节点引用的成功结果。应用结果属于本地编辑，需保存。窗口关闭和切换项目对未保存编辑提供保存/放弃选择；保存冲突不会自动覆盖远端版本。

同一活动会话仅消费每个成功任务一次，撤销也不重新消费；历史文档恢复时清空已消费任务引用。视口不参与内容撤销。

画布左键选择/框选，中键或滚轮平移，双指缩放，端点拖动连接；双击连线删除。Ctrl+S保存，Ctrl+Z/Ctrl+Shift+Z撤销/重做，Delete删除已选节点。节点侧栏中的草稿点击“应用节点设置”或“生成”后进入活动文档。
