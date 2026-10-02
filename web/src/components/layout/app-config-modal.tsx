"use client";

import { App, Button, Form, Input, Modal, Segmented, Select, Switch } from "antd";
import dynamic from "next/dynamic";
import { useEffect, useRef, useState } from "react";

import { ChannelModelSelectorModal } from "@/components/channel-model-selector-modal";
import type { WorkflowChannelSettings } from "@/components/workflow/workflow-channel-pane";
import { GrokTtsVoiceSelect } from "@/components/grok-tts-voice-select";
import { ModelPicker } from "@/components/model-picker";
import { fetchImageModels } from "@/services/api/image";
import { fetchWorkspaceConfig, measureCustomStorageProvider, syncWorkspaceModelConfig, syncCustomStorageProvider } from "@/services/api/workspace-config";
import { clearStorageConfigCache as clearFileStorageCache } from "@/services/file-storage";
import { clearStorageConfigCache as clearImageStorageCache, defaultCustomStorageProvider, defaultCustomWebDAVStorageProvider, loadStorageConfig, loadCustomS3StorageProvider, loadCustomWebDAVStorageProvider, saveCustomStorageProvider, saveCustomWebDAVStorageProvider, type CustomStorageProvider } from "@/services/image-storage";
import { audioFormatOptions, audioVoiceOptions, glmTtsFormatOptions, glmTtsVoiceOptions, isGlmTtsModel, normalizeAudioSpeedValue, normalizeGlmTtsFormat, normalizeGlmTtsSpeed, normalizeGlmTtsVoice } from "@/lib/audio-generation";
import { grokTtsFormatOptions, grokTtsLanguageOptions, isGrok2APITtsConfig, normalizeGrokTtsFormat, normalizeGrokTtsLanguage, normalizeGrokTtsSpeed } from "@/lib/grok-tts";
import { isGeminiConfig, isGeminiTtsModel } from "@/lib/gemini";
import { geminiTtsVoiceOptions, normalizeGeminiTtsVoice } from "@/lib/gemini-tts";
import { isMimoPresetTtsModel, isMimoTtsModel, isMimoVoiceCloneModel, isMimoVoiceDesignModel, mimoTtsFormatOptions, mimoTtsVoiceOptions } from "@/lib/mimo-tts";
import { isWorkflowProtocol, modelChannelApiKeyUrls, modelChannelDefaultBaseUrls, modelChannelProtocolOptions } from "@/lib/model-channel";
import type { WorkflowChannelData, WorkflowEntry } from "@/lib/workflow-channel";
import { listWorkflowChannels, readWorkflowChannel, replaceWorkflowChannels, saveWorkflowChannel } from "@/services/workflow-channel-storage";
import { filterChannelModelsByCapability, normalizeLocalChannels, useConfigStore, useEffectiveConfig, type AiConfig, type LocalModelChannel, type ModelCapability } from "@/stores/use-config-store";
import { useBackendStore } from "@/stores/use-backend-store";

type ModelGroup = {
    capability: ModelCapability;
    modelKey: "imageModel" | "videoModel" | "textModel" | "audioModel";
    workflowKey?: "imageWorkflowRef" | "videoWorkflowRef" | "audioWorkflowRef";
    channelKey: "imageChannelId" | "videoChannelId" | "textChannelId" | "audioChannelId";
    modelsKey: "imageModels" | "videoModels" | "textModels" | "audioModels";
    defaultLabel: string;
    optionsLabel: string;
};

const modelGroups: ModelGroup[] = [
    { capability: "image", modelKey: "imageModel", workflowKey: "imageWorkflowRef", channelKey: "imageChannelId", modelsKey: "imageModels", defaultLabel: "默认生图模型", optionsLabel: "生图模型可选项" },
    { capability: "video", modelKey: "videoModel", workflowKey: "videoWorkflowRef", channelKey: "videoChannelId", modelsKey: "videoModels", defaultLabel: "默认视频模型", optionsLabel: "视频模型可选项" },
    { capability: "text", modelKey: "textModel", channelKey: "textChannelId", modelsKey: "textModels", defaultLabel: "默认文本模型", optionsLabel: "文本模型可选项" },
    { capability: "audio", modelKey: "audioModel", workflowKey: "audioWorkflowRef", channelKey: "audioChannelId", modelsKey: "audioModels", defaultLabel: "默认音频模型", optionsLabel: "音频模型可选项" },
];

const WorkflowChannelPane = dynamic(() => import("@/components/workflow/workflow-channel-pane").then((module) => module.WorkflowChannelPane), { ssr: false });

export function AppConfigModal() {
    const { message } = App.useApp();
    const [loadingModels, setLoadingModels] = useState(false);
    const [savingConfig, setSavingConfig] = useState(false);
    const [modelSelectChannelId, setModelSelectChannelId] = useState("");
    const [workflowEntries, setWorkflowEntries] = useState<WorkflowEntry[]>([]);
    const workspaceConfigRef = useRef<{ ready: boolean; workflowChannels?: WorkflowChannelData[] }>({ ready: false });
    const [remoteStorageSyncEnabled, setRemoteStorageSyncEnabled] = useState(false);
    const [remoteWebDAVStorageSyncEnabled, setRemoteWebDAVStorageSyncEnabled] = useState(false);
    const [allowCustomStorageProvider, setAllowCustomStorageProvider] = useState(false);
    const [customStorage, setCustomStorage] = useState(() => defaultCustomStorageProvider());
    const [customWebDAVStorage, setCustomWebDAVStorage] = useState(() => defaultCustomWebDAVStorageProvider());
    const [measuringStorageType, setMeasuringStorageType] = useState<"s3" | "webdav" | null>(null);
    const [storageUsageText, setStorageUsageText] = useState("");
    const [webDAVStorageUsageText, setWebDAVStorageUsageText] = useState("");
    const config = useConfigStore((state) => state.config);
    const updateConfig = useConfigStore((state) => state.updateConfig);
    const isConfigOpen = useConfigStore((state) => state.isConfigOpen);
    const shouldPromptContinue = useConfigStore((state) => state.shouldPromptContinue);
    const setConfigDialogOpen = useConfigStore((state) => state.setConfigDialogOpen);
    const clearPromptContinue = useConfigStore((state) => state.clearPromptContinue);
    const publicSettings = useConfigStore((state) => state.publicSettings);
    const backendConnected = useBackendStore((state) => state.available);
    const effectiveConfig = useEffectiveConfig();
    const modelChannel = publicSettings?.modelChannel;
    const backendAvailable = Boolean(backendConnected);
    const canUseRemoteChannel = Boolean(backendConnected);
    const allowCustomChannel = backendAvailable && modelChannel?.allowCustomChannel === true;
    const effectiveMode = canUseRemoteChannel ? (allowCustomChannel ? config.channelMode : "remote") : "local";
    const localModelConfig: AiConfig = effectiveMode === "local" && config.channelMode !== "local" ? { ...config, channelMode: "local" } : config;
    const modelConfig = effectiveMode === "remote" ? effectiveConfig : localModelConfig;
    const canUseCustomStorageProvider = allowCustomStorageProvider;
    const glmTts = isGlmTtsModel(config.audioModel);
    const grokTts = isGrok2APITtsConfig({ ...modelConfig, model: config.audioModel, audioModel: config.audioModel }, config.audioModel);
    const geminiTts = isGeminiTtsModel(config.audioModel) && isGeminiConfig({ ...modelConfig, model: config.audioModel, audioModel: config.audioModel }, config.audioModel);
    const modelSelectChannel = normalizeLocalChannels(config).find((channel) => channel.id === modelSelectChannelId);

    useEffect(() => {
        setWorkflowEntries([]);
        if (!modelSelectChannel || !isWorkflowProtocol(modelSelectChannel.protocol)) return;
        const protocol = modelSelectChannel.protocol;
        let canceled = false;
		void readWorkflowChannel(protocol, modelSelectChannel.id)
            .then((items) => { if (!canceled) setWorkflowEntries(items); })
            .catch((error) => { if (!canceled) message.error(error instanceof Error ? error.message : "读取工作流配置失败"); });
        return () => { canceled = true; };
    }, [modelSelectChannelId, modelSelectChannel?.protocol]);

    useEffect(() => {
        setCustomStorage(loadCustomS3StorageProvider() || defaultCustomStorageProvider());
        setCustomWebDAVStorage(loadCustomWebDAVStorageProvider() || defaultCustomWebDAVStorageProvider());
        workspaceConfigRef.current = { ready: false };
        if (!isConfigOpen || !backendConnected) return;
        let canceled = false;
        void fetchWorkspaceConfig()
            .then(async (payload) => {
                if (canceled || !useBackendStore.getState().available) return;
                const remoteConfig = payload.modelConfig;
                const remoteWorkflowChannels = remoteConfig?.workflowChannels;
                const syncS3 = remoteConfig?.syncStorageConfig === true;
                const syncWebDAV = remoteConfig?.syncWebDAVStorageConfig === true;
                setRemoteStorageSyncEnabled(syncS3);
                setRemoteWebDAVStorageSyncEnabled(syncWebDAV);
                if (remoteConfig) {
                    const { workflowChannels, ...modelFields } = remoteConfig;
                    delete modelFields.workflowSyncTouched;
                    if (workflowChannels !== undefined) {
                        try {
                            await replaceWorkflowChannels(workflowChannels);
                            if (canceled || !useBackendStore.getState().available) return;
                            updateConfig("workflowSyncTouched", true);
                        } catch {
                            if (canceled || !useBackendStore.getState().available) return;
                            updateConfig("workflowSyncTouched", false);
                        }
                    }
                    if (canceled || !useBackendStore.getState().available) return;
                    Object.entries(modelFields)
                        .forEach(([key, value]) => updateConfig(key as keyof AiConfig, value as never));
                }
                workspaceConfigRef.current = { ready: true, workflowChannels: remoteWorkflowChannels };
                if (remoteWorkflowChannels === undefined) updateConfig("workflowSyncTouched", true);
                updateConfig("syncStorageConfig", syncS3);
                updateConfig("syncWebDAVStorageConfig", syncWebDAV);
                if (syncS3 && payload.storageProvider?.s3) {
                    const next = { ...defaultCustomStorageProvider(), ...payload.storageProvider.s3, type: "s3" as const };
                    setCustomStorage(next);
                    saveCustomStorageProvider(next);
                }
                if (syncWebDAV && payload.storageProvider?.webdav) {
                    const next = { ...defaultCustomWebDAVStorageProvider(), ...payload.storageProvider.webdav, type: "webdav" as const };
                    setCustomWebDAVStorage(next);
                    saveCustomWebDAVStorageProvider(next);
                }
            })
            .catch(() => { });
        return () => {
            canceled = true;
        };
    }, [isConfigOpen, backendConnected, updateConfig]);

    useEffect(() => {
        if (!isConfigOpen) return;
        let canceled = false;
        void loadStorageConfig()
            .then((storage) => {
                if (!canceled) setAllowCustomStorageProvider(storage.allowCustomProvider === true);
            })
            .catch(() => {
                if (!canceled) setAllowCustomStorageProvider(false);
            });
        return () => {
            canceled = true;
        };
    }, [isConfigOpen]);

    const finishConfig = async () => {
        const localIncomplete = effectiveMode === "local" && normalizeLocalChannels(config).filter((channel) => !isWorkflowProtocol(channel.protocol)).some((channel) => !channel.baseUrl.trim() || !channel.apiKey.trim());
        const modelIncomplete = !modelConfig.imageModel.trim() || !modelConfig.videoModel.trim() || !modelConfig.textModel.trim();
		if (customStorage.enabled && customWebDAVStorage.enabled) {
			message.error("S3/R2 与 WebDAV 不能同时启用");
			return;
		}
		if (backendConnected && !workspaceConfigRef.current.ready) {
			message.warning("工作区配置仍在加载，请稍后再保存");
			return;
		}
        if (!canUseRemoteChannel && config.channelMode !== "local") updateConfig("channelMode", "local");
        else if (canUseRemoteChannel && !allowCustomChannel && config.channelMode !== "remote") updateConfig("channelMode", "remote");
        if (canUseCustomStorageProvider) {
            saveCustomStorageProvider(customStorage);
            saveCustomWebDAVStorageProvider(customWebDAVStorage);
        }
        setSavingConfig(true);
		try {
			if (backendConnected) {
                const configToSave = effectiveMode === "local" && config.channelMode !== "local" ? { ...config, channelMode: "local" as const } : config;
                const workflowChannels = normalizeLocalChannels(config).filter((channel) => isWorkflowProtocol(channel.protocol));
                let workflowData = workspaceConfigRef.current.workflowChannels;
                if (config.workflowSyncTouched) {
                    const stored = await listWorkflowChannels();
                    const activeKeys = new Set(workflowChannels.map((channel) => `${channel.protocol}:${channel.id}`));
                    workflowData = stored.filter((channel) => activeKeys.has(`${channel.protocol}:${channel.channelId}`));
                }
                await syncWorkspaceModelConfig(configToSave, workflowData);
            }
            const providers = {
                ...(config.syncStorageConfig || remoteStorageSyncEnabled ? { s3: config.syncStorageConfig ? customStorage : { ...customStorage, enabled: false, endpoint: "", bucket: "", accessKeyId: "", secretAccessKey: "" } } : {}),
                ...(config.syncWebDAVStorageConfig || remoteWebDAVStorageSyncEnabled ? { webdav: config.syncWebDAVStorageConfig ? customWebDAVStorage : { ...customWebDAVStorage, enabled: false, endpoint: "", username: "", password: "" } } : {}),
            };
            if (backendConnected && canUseCustomStorageProvider && Object.keys(providers).length) {
                await syncCustomStorageProvider(providers);
                setRemoteStorageSyncEnabled(config.syncStorageConfig);
                setRemoteWebDAVStorageSyncEnabled(config.syncWebDAVStorageConfig);
            }
            clearImageStorageCache();
            clearFileStorageCache();
            setConfigDialogOpen(false);
            if ((config.syncStorageConfig || config.syncWebDAVStorageConfig) && !backendConnected) message.warning("请连接后端后再同步配置");
            else if (localIncomplete || modelIncomplete) message.warning("部分模型或本地渠道密钥尚未配置完整，配置已保存");
            else message.success(shouldPromptContinue ? "配置已保存，请继续刚才的请求" : "配置已保存");
            clearPromptContinue();
        } catch (error) {
            message.error(error instanceof Error ? "同步配置失败：" + error.message : "同步配置失败");
        } finally {
            setSavingConfig(false);
        }
    };

    const refreshModels = async () => {
        if (effectiveMode === "remote") return;
        const allChannels = normalizeLocalChannels(config);
        const channels = allChannels.filter((channel) => !isWorkflowProtocol(channel.protocol));
        if (channels.some((channel) => !channel.baseUrl.trim() || !channel.apiKey.trim())) {
            message.error("请先填写所有本地渠道的 Base URL 和 API Key");
            return;
        }
        setLoadingModels(true);
        try {
            const results = await Promise.allSettled(channels.map(async (channel) => fetchImageModels(configForLocalChannel(config, channel))));
            const updated = new Map(channels.map((channel, index) => [channel.id, results[index]] as const));
            updateLocalChannels(allChannels.map((channel) => {
                const result = updated.get(channel.id);
                return result?.status === "fulfilled" ? { ...channel, models: result.value } : channel;
            }));
            const failedCount = results.filter((result) => result.status === "rejected").length;
            if (failedCount) message.warning(`${failedCount} 个渠道拉取失败，已保留原有模型，可在“选择”中手动增加模型`);
            else message.success("模型列表已更新");
        } finally {
            setLoadingModels(false);
        }
    };

	const updateLocalChannels = (channels: LocalModelChannel[]) => {
		const normalized = channels.length ? channels : normalizeLocalChannels({ baseUrl: config.baseUrl, apiKey: config.apiKey, models: config.models });
		const modelChannels = normalized.filter((channel) => !isWorkflowProtocol(channel.protocol));
        const models = uniqueModels(modelChannels.flatMap((channel) => channel.models));
        const nextImageModels = filterChannelModelsByCapability(modelChannels, "image");
        const nextVideoModels = filterChannelModelsByCapability(modelChannels, "video");
        const nextTextModels = filterChannelModelsByCapability(modelChannels, "text");
        const nextAudioModels = filterChannelModelsByCapability(modelChannels, "audio");
        const imageModel = nextImageModels.includes(config.imageModel) ? config.imageModel : nextImageModels[0] || "";
        const videoModel = nextVideoModels.includes(config.videoModel) ? config.videoModel : nextVideoModels[0] || "";
        const textModel = nextTextModels.includes(config.textModel) ? config.textModel : nextTextModels[0] || "";
        const audioModel = nextAudioModels.includes(config.audioModel) ? config.audioModel : nextAudioModels[0] || "";
        updateConfig("localChannels", normalized);
        updateConfig("models", models);
        updateConfig("imageModels", nextImageModels);
        updateConfig("videoModels", nextVideoModels);
        updateConfig("textModels", nextTextModels);
        updateConfig("audioModels", nextAudioModels);
        updateConfig("imageModel", imageModel);
        updateConfig("videoModel", videoModel);
        updateConfig("textModel", textModel);
        updateConfig("audioModel", audioModel);
        updateConfig("imageChannelId", channelIdForLocalModel(modelChannels, imageModel, config.imageChannelId));
        updateConfig("videoChannelId", channelIdForLocalModel(modelChannels, videoModel, config.videoChannelId));
        updateConfig("textChannelId", channelIdForLocalModel(modelChannels, textModel, config.textChannelId));
        updateConfig("audioChannelId", channelIdForLocalModel(modelChannels, audioModel, config.audioChannelId));
        updateConfig("baseUrl", modelChannels[0]?.baseUrl || config.baseUrl);
        updateConfig("apiKey", modelChannels[0]?.apiKey || config.apiKey);
    };

    const patchLocalChannel = (id: string, patch: Partial<LocalModelChannel>) => {
        updateLocalChannels(normalizeLocalChannels(config).map((channel) => (channel.id === id ? { ...channel, ...patch } : channel)));
    };

    const addLocalChannel = () => {
        updateLocalChannels([...normalizeLocalChannels(config), { id: "local-" + Date.now(), protocol: "openai", name: "新渠道", baseUrl: modelChannelDefaultBaseUrls.openai, apiKey: "", models: [] }]);
    };

    const removeLocalChannel = (id: string) => {
        updateLocalChannels(normalizeLocalChannels(config).filter((channel) => channel.id !== id));
    };

    const openLocalModelSelector = (channel: LocalModelChannel) => setModelSelectChannelId(channel.id);

    const closeLocalModelSelector = () => setModelSelectChannelId("");

    const confirmLocalModelSelector = (models: string[]) => {
        if (!modelSelectChannelId) return;
        patchLocalChannel(modelSelectChannelId, { models });
        closeLocalModelSelector();
    };

    const changePersonalWorkflows = (items: WorkflowEntry[]) => {
        if (!modelSelectChannel || !isWorkflowProtocol(modelSelectChannel.protocol)) return;
        const protocol = modelSelectChannel.protocol;
        setWorkflowEntries(items);
        patchLocalChannel(modelSelectChannel.id, { workflowSummaries: items.map(({ provider, kind, workflowId, title, capability, enabled }) => ({ provider, kind, workflowId, title, capability, enabled })) });
		void saveWorkflowChannel(protocol, modelSelectChannel.id, items).catch((error) => {
			message.error(error instanceof Error ? error.message : "保存工作流配置失败");
        });
    };

    const syncPersonalWorkflows = async () => {
        if (!backendConnected || !modelSelectChannel || !isWorkflowProtocol(modelSelectChannel.protocol)) {
            throw new Error("请先连接后端服务并选择工作流渠道");
        }
        const channelId = modelSelectChannel.id;
        const protocol = modelSelectChannel.protocol;
        const entries = workflowEntries;
        if (!useBackendStore.getState().available) {
            throw new Error("后端连接已变化");
		}
		const current = useConfigStore.getState().config;
		if (!current.workflowSyncTouched) throw new Error("工作区配置仍在加载");
		await saveWorkflowChannel(protocol, channelId, entries);
        if (!useBackendStore.getState().available) {
            throw new Error("后端连接已变化");
        }
        const channels = normalizeLocalChannels(current).filter((item) => isWorkflowProtocol(item.protocol));
        const keys = new Set(channels.map((item) => `${item.protocol}:${item.id}`));
        if (!keys.has(`${protocol}:${channelId}`)) throw new Error("工作流渠道已变化");
        const saved = (await listWorkflowChannels()).filter((item) => keys.has(`${item.protocol}:${item.channelId}`));
        if (!useBackendStore.getState().available) {
            throw new Error("后端连接已变化");
        }
        await syncWorkspaceModelConfig(current, saved);
        if (!useBackendStore.getState().available) {
            throw new Error("后端连接已变化");
        }
        return channelId;
    };

    const finishWorkflowChannel = () => {
        if (!backendConnected) {
            closeLocalModelSelector();
            return;
        }
        void syncPersonalWorkflows()
            .then(closeLocalModelSelector)
            .catch((error) => message.error(error instanceof Error ? error.message : "同步工作流配置失败"));
    };

    const fetchLocalModelList = async () => {
        if (!modelSelectChannel) return;
        if (!modelSelectChannel.baseUrl.trim() || !modelSelectChannel.apiKey.trim()) {
            message.error("请先填写该渠道的 Base URL 和 API Key");
            return;
        }
        return uniqueModels(await fetchImageModels(configForLocalChannel(config, modelSelectChannel)));
    };


    const measureStorage = async (provider: CustomStorageProvider) => {
        if (!backendConnected) {
            message.warning("请先连接后端服务后再统计容量");
            return;
        }
        setMeasuringStorageType(provider.type);
        try {
            const result = await measureCustomStorageProvider(provider);
            const usageText = formatBytes(result.bytes) + " / " + formatBytes(result.limitBytes) + (result.overLimit ? "，已达到上限" : "");
            if (provider.type === "webdav") {
                setWebDAVStorageUsageText(usageText);
                if (result.overLimit) {
                    const next = { ...customWebDAVStorage, enabled: false };
                    setCustomWebDAVStorage(next);
                    saveCustomWebDAVStorageProvider(next);
                }
            } else {
                setStorageUsageText(usageText);
                if (result.overLimit) {
                    const next = { ...customStorage, enabled: false };
                    setCustomStorage(next);
                    saveCustomStorageProvider(next);
                }
            }
            message.success("容量统计完成");
        } catch (error) {
            message.error(error instanceof Error ? error.message : "容量统计失败");
        } finally {
            setMeasuringStorageType(null);
        }
    };

    return (
        <>
            <Modal
            title={
                <div>
                    <div className="text-lg font-semibold">配置与偏好</div>
                    <div className="mt-1 text-xs font-normal text-stone-500">模型、渠道和画布默认行为</div>
                </div>
            }
            open={isConfigOpen}
            width={960}
            centered
            onCancel={() => setConfigDialogOpen(false)}
            styles={{ body: { maxHeight: "72vh", overflowY: "auto", paddingRight: 18 } }}
            footer={
                <Button type="primary" loading={savingConfig} onClick={() => void finishConfig()}>
                    完成
                </Button>
            }
        >
            <div className="pt-1">
                <Form layout="vertical" requiredMark={false}>
                    {allowCustomChannel && canUseRemoteChannel ? (
                        <Form.Item label="渠道模式" className="mb-5">
                            <Segmented
                                block
                                size="middle"
                                value={effectiveMode}
                                onChange={(value) => updateConfig("channelMode", value as AiConfig["channelMode"])}
                                options={[
                                    { label: "本地直连", value: "local" },
                                    { label: "云端渠道", value: "remote" },
                                ]}
                            />
                        </Form.Item>
                    ) : null}
                    {effectiveMode === "local" ? (
                        <>
                            <div className="mb-5 space-y-3 rounded-lg border border-stone-200 p-3 dark:border-stone-800">
                                <div className="flex items-center justify-between gap-3">
                                    <div>
                                        <div className="text-sm font-medium">本地模型渠道</div>
                                        <div className="mt-1 text-xs text-stone-500">可为生图、视频、文本、音频分别选择不同渠道的模型。</div>
                                    </div>
                                    <Button size="small" onClick={addLocalChannel}>
                                        新增渠道
                                    </Button>
                                </div>
                                {normalizeLocalChannels(config).map((channel, index) => (
                                    <div key={channel.id} className="space-y-2 rounded-md bg-stone-50 p-2 dark:bg-stone-900">
                                        <div className="grid gap-2 md:grid-cols-[130px_150px_minmax(0,1fr)_minmax(0,1fr)_auto]">
                                            <Input value={channel.name} placeholder="渠道名称" onChange={(event) => patchLocalChannel(channel.id, { name: event.target.value })} />
                                            <Select
                                                value={channel.protocol}
                                                options={modelChannelProtocolOptions}
                                                onChange={(protocol: LocalModelChannel["protocol"]) => patchLocalChannel(channel.id, { protocol, baseUrl: modelChannelDefaultBaseUrls[protocol] })}
                                            />
                                            {isWorkflowProtocol(channel.protocol) ? <div className="flex h-8 items-center justify-center rounded-md border border-dashed border-[var(--ant-color-border)] px-3 text-xs text-[var(--ant-color-text-secondary)] md:col-span-2">连接信息和工作流在「选择」里配置</div> : <><Input value={channel.baseUrl} placeholder="Base URL" onChange={(event) => patchLocalChannel(channel.id, { baseUrl: event.target.value })} /><Input.Password value={channel.apiKey} placeholder="API Key" onChange={(event) => patchLocalChannel(channel.id, { apiKey: event.target.value })} /></>}
                                            <div className="relative flex flex-wrap gap-2 md:flex-nowrap">
                                                <Button size="small" onClick={() => openLocalModelSelector(channel)}>
                                                    选择
                                                </Button>
                                                <Button size="small" danger disabled={index === 0 && normalizeLocalChannels(config).length === 1} onClick={() => removeLocalChannel(channel.id)}>
                                                    删除
                                                </Button>
                                                {modelChannelApiKeyUrls[channel.protocol] ? (
                                                    <div className="w-full md:absolute md:left-0 md:top-8">
                                                        <Button block type="primary" size="small" href={modelChannelApiKeyUrls[channel.protocol]} target="_blank">
                                                            获取 API Key
                                                        </Button>
                                                    </div>
                                                ) : null}
                                            </div>
                                        </div>
                                        <div className="text-xs text-stone-500">{isWorkflowProtocol(channel.protocol) ? `已保存 ${channel.workflowSummaries?.length || 0} 条工作流` : `已保存 ${channel.models.length} 个模型`}</div>
                                    </div>
                                ))}
                            </div>
                            <div className="mb-5 flex items-center justify-between gap-3 rounded-lg border border-stone-200 px-3 py-2 dark:border-stone-800">
                                <div className="min-w-0">
                                    <div className="text-sm font-medium">模型列表</div>
                                    <div className="mt-1 text-xs text-stone-500">当前已保存 {config.models.length} 个模型</div>
                                </div>
                                <div className="flex shrink-0 flex-wrap items-center justify-end gap-2">
                                    <Button size="small" loading={loadingModels} onClick={() => void refreshModels()}>
                                        拉取全部渠道
                                    </Button>
                                </div>
                            </div>
                        </>
                    ) : (
                        <div className="mb-5 rounded-lg border border-stone-200 p-3 text-sm text-stone-500 dark:border-stone-800">
                            <div className="font-medium text-stone-900 dark:text-stone-100">云端渠道</div>
                            <div className="mt-1">由系统渠道转发请求，当前可用 {modelChannel?.availableModels.length || 0} 个模型。</div>
                        </div>
                    )}
                    <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
                        {modelGroups.map((group) => (
                            <Form.Item key={group.modelKey} label={group.defaultLabel} className="mb-4">
                                <ModelPicker config={modelConfig} value={modelConfig[group.modelKey]} channelId={modelConfig[group.channelKey]} workflowRef={group.workflowKey ? modelConfig[group.workflowKey] : undefined} onWorkflowChange={group.workflowKey ? (ref) => updateConfig(group.workflowKey!, ref) : undefined} onChange={(model, channelId) => { updateConfig(group.modelKey, model); if (channelId) updateConfig(group.channelKey, channelId); }} capability={group.capability} fullWidth />
                            </Form.Item>
                        ))}
                    </div>
                    <div className="grid gap-4 md:grid-cols-4">
                        <Form.Item label="画布默认生图张数" extra="新建画布生图和配置节点默认使用，单个节点仍可单独覆盖。" className="mb-4">
                            <Input
                                type="number"
                                min={1}
                                max={15}
                                value={config.canvasImageCount}
                                onChange={(event) => updateConfig("canvasImageCount", event.target.value)}
                                onBlur={(event) => updateConfig("canvasImageCount", normalizeImageCount(event.target.value))}
                            />
                        </Form.Item>
                        {geminiTts ? (
                            <Form.Item label="默认 Gemini 音色" className="mb-4">
                                <Select showSearch optionFilterProp="label" value={normalizeGeminiTtsVoice(config.geminiTtsVoice)} options={geminiTtsVoiceOptions} onChange={(value) => updateConfig("geminiTtsVoice", value)} />
                            </Form.Item>
                        ) : isMimoPresetTtsModel(config.audioModel) ? (
                            <Form.Item label="默认 MiMo 音色" className="mb-4">
                                <Select value={config.mimoTtsVoice} options={[...mimoTtsVoiceOptions]} onChange={(value) => updateConfig("mimoTtsVoice", value)} />
                            </Form.Item>
                        ) : isMimoVoiceDesignModel(config.audioModel) ? (
                            <Form.Item label="默认音色描述" className="mb-4">
                                <Input value={config.mimoVoiceDesignPrompt} placeholder="例如：年轻女性，声音清亮自然，有亲和力。" onChange={(event) => updateConfig("mimoVoiceDesignPrompt", event.target.value)} />
                            </Form.Item>
                        ) : isMimoTtsModel(config.audioModel) ? null : (
                            <Form.Item label="默认音频声音" className="mb-4">
                                {grokTts ? <GrokTtsVoiceSelect config={modelConfig} model={config.audioModel} value={config.grokTtsVoice} enabled={isConfigOpen} onChange={(value) => updateConfig("grokTtsVoice", value)} /> : <Select value={glmTts ? normalizeGlmTtsVoice(config.glmTtsVoice) : config.audioVoice} options={glmTts ? glmTtsVoiceOptions : audioVoiceOptions} onChange={(value) => updateConfig(glmTts ? "glmTtsVoice" : "audioVoice", value)} />}
                            </Form.Item>
                        )}
                        {grokTts ? (
                            <Form.Item label="默认音频语言" className="mb-4">
                                <Select value={normalizeGrokTtsLanguage(config.grokTtsLanguage)} options={grokTtsLanguageOptions} showSearch optionFilterProp="label" onChange={(value) => updateConfig("grokTtsLanguage", value)} />
                            </Form.Item>
                        ) : null}
                        {!geminiTts ? (
                            <Form.Item label="默认音频格式" className="mb-4">
                                <Select value={isMimoTtsModel(config.audioModel) ? config.mimoTtsFormat : glmTts ? normalizeGlmTtsFormat(config.glmTtsFormat) : grokTts ? normalizeGrokTtsFormat(config.grokTtsFormat) : config.audioFormat} options={isMimoTtsModel(config.audioModel) ? [...mimoTtsFormatOptions] : glmTts ? glmTtsFormatOptions : grokTts ? grokTtsFormatOptions : audioFormatOptions} onChange={(value) => isMimoTtsModel(config.audioModel) ? updateConfig("mimoTtsFormat", value) : updateConfig(glmTts ? "glmTtsFormat" : grokTts ? "grokTtsFormat" : "audioFormat", value)} />
                            </Form.Item>
                        ) : null}
                        {!geminiTts && !isMimoTtsModel(config.audioModel) ? (
                            <Form.Item label="默认音频语速" className="mb-4">
                                <Input
                                    type="number"
                                    min={glmTts ? 0.5 : grokTts ? 0.7 : 0.25}
                                    max={glmTts ? 2 : grokTts ? 1.5 : 4}
                                    step={0.05}
                                    value={glmTts ? config.glmTtsSpeed : grokTts ? config.grokTtsSpeed : config.audioSpeed}
                                    onChange={(event) => updateConfig(glmTts ? "glmTtsSpeed" : grokTts ? "grokTtsSpeed" : "audioSpeed", event.target.value)}
                                    onBlur={(event) => updateConfig(glmTts ? "glmTtsSpeed" : grokTts ? "grokTtsSpeed" : "audioSpeed", glmTts ? normalizeGlmTtsSpeed(event.target.value) : grokTts ? normalizeGrokTtsSpeed(event.target.value) : normalizeAudioSpeedValue(event.target.value))}
                                />
                            </Form.Item>
                        ) : null}
                    </div>
                    <div className="mb-4 grid gap-3 md:grid-cols-3">
                        <FeatureSwitch title="流式传输" description="开启后请求中追加 stream，支持读取中间图片事件并避免长时间无数据。" checked={Boolean(config.streamImages)} onChange={(checked) => updateConfig("streamImages", checked ? "1" : "")} />
                        <FeatureSwitch title="返回 Base64 图片数据" description="开启后 Image API 请求会追加 response_format: b64_json。" checked={Boolean(config.responseFormatB64Json)} onChange={(checked) => updateConfig("responseFormatB64Json", checked ? "1" : "")} />
                        <FeatureSwitch title="Codex CLI 兼容模式" description="开启后减少不兼容参数，并追加防提示词改写前缀。" checked={Boolean(config.codexCli)} onChange={(checked) => updateConfig("codexCli", checked ? "1" : "")} />
                    </div>
                    {canUseCustomStorageProvider ? (
                        <>
                            <section className="mb-5 mt-4 rounded-xl border border-stone-200 bg-stone-50/70 p-3 dark:border-stone-800 dark:bg-stone-900/50">
                                <div className="flex items-center justify-between gap-3">
                                    <div>
                                        <div className="text-sm font-medium">用户 S3/R2 存储</div>
                                        <div className="mt-1 text-xs text-stone-500">
                                            开启后，新生成图片和媒体文件会优先保存到你的 S3 兼容对象存储。
                                            {storageUsageText ? <>当前容量：{storageUsageText}</> : null}
                                        </div>
                                    </div>
                                    <div className="flex shrink-0 flex-wrap items-center justify-end gap-2">
                                        <Button size="small" loading={measuringStorageType === "s3"} onClick={() => void measureStorage(customStorage)}>
                                            统计容量
                                        </Button>
                                        <span className="text-xs text-stone-500">自动同步</span>
                                        <Switch size="small" checked={config.syncStorageConfig} onChange={(checked) => updateConfig("syncStorageConfig", checked)} />
                                        <Switch checked={customStorage.enabled} disabled={customWebDAVStorage.enabled} onChange={(enabled) => setCustomStorage((value) => ({ ...value, enabled }))} />
                                    </div>
                                </div>
                                {customStorage.enabled ? (
                                    <div className="mt-3 grid gap-3 md:grid-cols-2">
                                        <Input value={customStorage.name} placeholder="配置名称" onChange={(event) => setCustomStorage((value) => ({ ...value, name: event.target.value }))} />
                                        <Input value={customStorage.endpoint} placeholder="Endpoint，例如 https://<account>.r2.cloudflarestorage.com" onChange={(event) => setCustomStorage((value) => ({ ...value, endpoint: event.target.value }))} />
                                        <Input value={customStorage.region} placeholder="Region，R2 通常为 auto" onChange={(event) => setCustomStorage((value) => ({ ...value, region: event.target.value }))} />
                                        <Input value={customStorage.bucket} placeholder="Bucket 名称" onChange={(event) => setCustomStorage((value) => ({ ...value, bucket: event.target.value }))} />
                                        <Input value={customStorage.accessKeyId} placeholder="Access Key ID" onChange={(event) => setCustomStorage((value) => ({ ...value, accessKeyId: event.target.value }))} />
                                        <Input.Password value={customStorage.secretAccessKey} placeholder="Secret Access Key" onChange={(event) => setCustomStorage((value) => ({ ...value, secretAccessKey: event.target.value }))} />
                                        <Input value={customStorage.publicBaseUrl} placeholder="公开访问地址，例如 https://pub-xxx.r2.dev" onChange={(event) => setCustomStorage((value) => ({ ...value, publicBaseUrl: event.target.value }))} />
                                        <Input value={customStorage.pathPrefix} placeholder="保存路径前缀，例如 images" onChange={(event) => setCustomStorage((value) => ({ ...value, pathPrefix: event.target.value }))} />
                                    </div>
                                ) : null}
                            </section>
                            <section className="mb-5 mt-4 rounded-xl border border-stone-200 bg-stone-50/70 p-3 dark:border-stone-800 dark:bg-stone-900/50">
                                <div className="flex items-center justify-between gap-3">
                                    <div>
                                        <div className="text-sm font-medium">WebDAV 存储</div>
                                        <div className="mt-1 text-xs text-stone-500">
                                            开启后，新生成图片和媒体文件会优先保存到你的 WebDAV。
                                            {webDAVStorageUsageText ? <>当前容量：{webDAVStorageUsageText}</> : null}
                                        </div>
                                    </div>
                                    <div className="flex shrink-0 flex-wrap items-center justify-end gap-2">
                                        <Button size="small" loading={measuringStorageType === "webdav"} onClick={() => void measureStorage(customWebDAVStorage)}>
                                            统计容量
                                        </Button>
                                        <span className="text-xs text-stone-500">自动同步</span>
                                        <Switch size="small" checked={config.syncWebDAVStorageConfig} onChange={(checked) => updateConfig("syncWebDAVStorageConfig", checked)} />
                                        <Switch checked={customWebDAVStorage.enabled} disabled={customStorage.enabled} onChange={(enabled) => setCustomWebDAVStorage((value) => ({ ...value, enabled }))} />
                                    </div>
                                </div>
                                {customWebDAVStorage.enabled ? (
                                    <div className="mt-3 grid gap-3 md:grid-cols-2">
                                        <Input value={customWebDAVStorage.name} placeholder="配置名称" onChange={(event) => setCustomWebDAVStorage((value) => ({ ...value, name: event.target.value }))} />
                                        <Input value={customWebDAVStorage.endpoint} placeholder="WebDAV 地址" onChange={(event) => setCustomWebDAVStorage((value) => ({ ...value, endpoint: event.target.value }))} />
                                        <Input value={customWebDAVStorage.pathPrefix} placeholder="远程目录" onChange={(event) => setCustomWebDAVStorage((value) => ({ ...value, pathPrefix: event.target.value }))} />
                                        <Input value={customWebDAVStorage.username} placeholder="用户名" onChange={(event) => setCustomWebDAVStorage((value) => ({ ...value, username: event.target.value }))} />
                                        <Input.Password value={customWebDAVStorage.password} placeholder="密码 / 应用密码" onChange={(event) => setCustomWebDAVStorage((value) => ({ ...value, password: event.target.value }))} />
                                    </div>
                                ) : null}
                            </section>
                        </>
                    ) : null}
                    {(!isMimoTtsModel(config.audioModel) || isMimoPresetTtsModel(config.audioModel) || isMimoVoiceCloneModel(config.audioModel)) && !glmTts && !grokTts ? (
                        <Form.Item label="默认音频指令" className="mb-4">
                            <Input.TextArea rows={2} value={config.audioInstructions} placeholder="例如：自然、温暖、适合旁白。" onChange={(event) => updateConfig("audioInstructions", event.target.value)} />
                        </Form.Item>
                    ) : null}
                    {effectiveMode === "local" ? (
                        <Form.Item label="系统提示词" className="mb-0">
                            <Input.TextArea rows={3} value={config.systemPrompt} placeholder="例如：你是一位擅长电影感写实摄影的视觉导演。" onChange={(event) => updateConfig("systemPrompt", event.target.value)} />
                        </Form.Item>
                    ) : null}
                </Form>
            </div>
            </Modal>
            {modelSelectChannel && isWorkflowProtocol(modelSelectChannel.protocol) ? (
                <Modal title={`${modelSelectChannel.name || "工作流渠道"} · ${modelSelectChannel.protocol === "runninghub" ? "RunningHub" : "ComfyUI"} 工作流`} open width="75vw" centered onCancel={finishWorkflowChannel} footer={<Button type="primary" onClick={finishWorkflowChannel}>完成</Button>} styles={{ body: { maxHeight: "76vh", overflowY: "auto" } }}>
                    <WorkflowChannelPane key={`${"default"}:${modelSelectChannel.protocol}:${modelSelectChannel.id}`} channel={modelSelectChannel as WorkflowChannelSettings} workflows={workflowEntries}  onChannelChange={(patch) => patchLocalChannel(modelSelectChannel.id, patch)} onBridgeDeleted={(bridgeId) => {
                        const current = useConfigStore.getState().config;
                        updateLocalChannels(normalizeLocalChannels(current).map((channel) => channel.protocol === "comfyui" && channel.bridgeId === bridgeId ? { ...channel, bridgeId: "" } : channel));
                    }} onWorkflowsChange={changePersonalWorkflows} onBeforeTest={syncPersonalWorkflows} />
                </Modal>
            ) : modelSelectChannel ? (
                <ChannelModelSelectorModal
                    channel={modelSelectChannel}
                    models={modelSelectChannel.models}
                    onCancel={closeLocalModelSelector}
                    onConfirm={confirmLocalModelSelector}
                    onFetchModels={fetchLocalModelList}
                />
            ) : null}
        </>
    );
}

function FeatureSwitch({ title, description, checked, onChange }: { title: string; description: string; checked: boolean; onChange: (checked: boolean) => void }) {
    return (
        <div className="rounded-lg border border-stone-200 px-3 py-2 dark:border-stone-800">
            <div className="flex items-center justify-between gap-3">
                <div className="text-sm font-medium">{title}</div>
                <Switch checked={checked} onChange={onChange} />
            </div>
            <div className="mt-1 text-xs leading-5 text-stone-500">{description}</div>
        </div>
    );
}

function configForLocalChannel(config: AiConfig, channel: LocalModelChannel): AiConfig {
    return {
        ...config,
        channelMode: "local",
        baseUrl: channel.baseUrl,
        apiKey: channel.apiKey,
        localChannels: [{ ...channel }],
        imageChannelId: channel.id,
        videoChannelId: channel.id,
        textChannelId: channel.id,
        audioChannelId: channel.id,
        model: channel.models[0] || config.model,
    };
}

function channelIdForLocalModel(channels: LocalModelChannel[], model: string, currentId: string) {
    if (!channels.length) return "";
    if (channels.some((channel) => channel.id === currentId && (!model || channel.models.includes(model)))) return currentId;
    return channels.find((channel) => model && channel.models.includes(model))?.id || channels[0].id;
}

function normalizeImageCount(value: string) {
    return String(Math.max(1, Math.min(15, Math.floor(Math.abs(Number(value)) || 3))));
}


function uniqueModels(models: string[]) {
    return Array.from(new Set(models.map((model) => model.trim()).filter(Boolean)));
}

function formatBytes(bytes: number) {
    if (bytes < 1024) return `${bytes}B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)}KB`;
    return `${(bytes / 1024 / 1024).toFixed(1)}MB`;
}
