"use client";

import type { ReactNode } from "react";
import { useEffect, useRef } from "react";
import { App } from "antd";
import { fetchWorkspaceConfig } from "@/services/api/workspace-config";
import { replaceWorkflowChannels } from "@/services/workflow-channel-storage";
import { STORAGE_SYNC_FAILED_EVENT, defaultCustomStorageProvider, defaultCustomWebDAVStorageProvider, saveCustomStorageProvider, saveCustomWebDAVStorageProvider } from "@/services/image-storage";
import { useConfigStore, type AiConfig } from "@/stores/use-config-store";
import { useBackendStore } from "@/stores/use-backend-store";

export function ClientRootInit({ children }: { children: ReactNode }) {
    const { message } = App.useApp();
    const handledConfigParams = useRef(false);
    const available = useBackendStore((state) => state.available);
    const initialize = useBackendStore((state) => state.initialize);
    const loadPublicSettings = useConfigStore((state) => state.loadPublicSettings);
    const publicSettings = useConfigStore((state) => state.publicSettings);
    const updateConfig = useConfigStore((state) => state.updateConfig);
    const openConfigDialog = useConfigStore((state) => state.openConfigDialog);
    useEffect(() => { void initialize(); void loadPublicSettings().catch(() => {}); }, [initialize, loadPublicSettings]);
    useEffect(() => {
        const onSyncFailed = (event: Event) => { const detail = (event as CustomEvent<string>).detail; message.warning({ key: STORAGE_SYNC_FAILED_EVENT, content: "云端同步失败，已保留原始素材" + (detail ? "：" + detail : "") }); };
        window.addEventListener(STORAGE_SYNC_FAILED_EVENT, onSyncFailed);
        return () => window.removeEventListener(STORAGE_SYNC_FAILED_EVENT, onSyncFailed);
    }, [message]);
    useEffect(() => {
        if (!available) { updateConfig("workflowSyncTouched", true); return; }
        let canceled = false;
        void fetchWorkspaceConfig().then(async (payload) => {
            if (canceled) return;
            if (payload.modelConfig) {
                const { workflowChannels, ...config } = payload.modelConfig;
                if (workflowChannels !== undefined) await replaceWorkflowChannels(workflowChannels);
                if (canceled) return;
                delete config.workflowSyncTouched;
                Object.entries(config).forEach(([key, value]) => updateConfig(key as keyof AiConfig, value as never));
            }
            const syncS3 = payload.modelConfig?.syncStorageConfig === true;
            const syncWebDAV = payload.modelConfig?.syncWebDAVStorageConfig === true;
            updateConfig("workflowSyncTouched", true);
            updateConfig("syncStorageConfig", syncS3);
            updateConfig("syncWebDAVStorageConfig", syncWebDAV);
            if (syncS3 && payload.storageProvider?.s3) saveCustomStorageProvider({ ...defaultCustomStorageProvider(), ...payload.storageProvider.s3, type: "s3" });
            if (syncWebDAV && payload.storageProvider?.webdav) saveCustomWebDAVStorageProvider({ ...defaultCustomWebDAVStorageProvider(), ...payload.storageProvider.webdav, type: "webdav" });
        }).catch(() => { if (!canceled) updateConfig("workflowSyncTouched", true); });
        return () => { canceled = true; };
    }, [available, updateConfig]);
    useEffect(() => {
        if (handledConfigParams.current) return;
        const searchParams = new URLSearchParams(window.location.search);
        const baseUrl = searchParams.get("baseUrl") || searchParams.get("baseurl");
        const apiKey = searchParams.get("apiKey") || searchParams.get("apikey");
        if (!baseUrl && !apiKey) return;
        if (!publicSettings) return;
        handledConfigParams.current = true;
        searchParams.delete("baseUrl");
        searchParams.delete("baseurl");
        searchParams.delete("apiKey");
        searchParams.delete("apikey");
        window.history.replaceState(null, "", `${window.location.pathname}${searchParams.size ? `?${searchParams}` : ""}${window.location.hash}`);
        if (!publicSettings.modelChannel.allowCustomChannel) {
            openConfigDialog(false);
            message.error("设置中尚未启用自定义渠道");
            return;
        }
        updateConfig("channelMode", "local");
        if (baseUrl) updateConfig("baseUrl", baseUrl);
        if (apiKey) updateConfig("apiKey", apiKey);
        openConfigDialog(false);
    }, [message, openConfigDialog, publicSettings, updateConfig]);


    return <>{children}</>;
}
