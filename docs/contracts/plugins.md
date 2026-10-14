# 插件v1

manifest=id/type(storage,media,provider)/version语义版本/apiVersion("1")/entry受信工厂标识/provides(storage,credentials,media,model集合)/dependencies插件ID数组/capabilities(text,image,video,audio)。config/plugins.json={version:1,plugins:[manifest]}是装配事实来源。

重复ID/能力/依赖、未知工厂、身份不符、缺失或循环依赖、不支持版本/provides均终止启动。工厂仅在bootstrap，无init隐式注册/磁盘代码加载。一个storage和media，多provider按channel.pluginId选，不猜模型名。实现ID/能力和清单一致。

registered→activating→active，失败failed且逆序回滚。关闭先取消等待任务，再逆拓扑释放。插件实例无共享可变状态/互相import；只注入声明的dataDir/masterKey/凭据端口。listPlugins无内部context。

storage.sqlite提供storage/credentials；media.local提供media；provider.openai提供model及text/image/video/audio，provider.gemini提供model及text/image，均依赖storage.sqlite凭据端口。版本1.0.0/API1。

渲染节点插件{id,version,apiVersion,kind,label,component}，ID/kind唯一/API1；四个类型各一个。不能自行扩持久化字段/访问桌面服务。无热卸载/OSS/MCP，不声明未实现能力。
