"use client";

import type { ReactNode } from "react";
import { useEffect } from "react";
import { App } from "antd";
import { fetchWorkspaceConfig } from "@/services/api/workspace-config";
import { replaceWorkflowChannels } from "@/services/workflow-channel-storage";
import { STORAGE_SYNC_FAILED_EVENT, defaultCustomStorageProvider, defaultCustomWebDAVStorageProvider, saveCustomStorageProvider, saveCustomWebDAVStorageProvider } from "@/services/image-storage";
import { defaultConfig, useConfigStore, type AiConfig } from "@/stores/use-config-store";
import { useBackendStore } from "@/stores/use-backend-store";

export function ClientRootInit({ children }: { children: ReactNode }) {
    const { message } = App.useApp();
    const available = useBackendStore((state) => state.available);
    const initialize = useBackendStore((state) => state.initialize);
    const loadPublicSettings = useConfigStore((state) => state.loadPublicSettings);
    const updateConfig = useConfigStore((state) => state.updateConfig);
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
                const storedConfig = payload.modelConfig as Partial<AiConfig> & Record<string, unknown>;
                const { workflowChannels, ...config } = storedConfig;
                if (workflowChannels !== undefined) await replaceWorkflowChannels(workflowChannels);
                if (canceled) return;
                delete config.workflowSyncTouched;
                Object.entries(config).filter(([key]) => key in defaultConfig && key !== "publicChannels").forEach(([key, value]) => updateConfig(key as keyof AiConfig, value as never));
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
    /* Removed legacy URL-based API credential injection. AI channels are managed by the local Go service. */


    return <>{children}</>;
}
