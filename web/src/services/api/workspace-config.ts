import { apiDelete, apiGet, apiPost } from "@/services/api/request";
import type { AiConfig } from "@/stores/use-config-store";
import type { WorkflowChannelData } from "@/lib/workflow-channel";
import { toProviderPayload, type CustomS3StorageProvider, type CustomStorageProvider, type CustomWebDAVStorageProvider } from "@/services/image-storage";

export type WorkspaceConfigPayload = {
    modelConfig?: Partial<AiConfig> & { workflowChannels?: WorkflowChannelData[] };
    storageProvider?: {
        s3?: Partial<CustomS3StorageProvider>;
        webdav?: Partial<CustomWebDAVStorageProvider>;
    };
    imageHistory?: unknown;
    assetData?: unknown;
    syncCapabilities?: {
        workspaceData?: boolean;
        workflows?: boolean;
        assets?: boolean;
    };
};

export type StorageCapacityResult = {
    bytes: number;
    limitBytes: number;
    overLimit: boolean;
    checkedAt: string;
    providerName: string;
};

export async function fetchWorkspaceConfig() {
    return apiGet<WorkspaceConfigPayload>("/api/v1/config", undefined);
}

export async function syncWorkspaceModelConfig(config: AiConfig, workflowChannels?: WorkflowChannelData[]) {
    const cleanConfig = Object.fromEntries(Object.entries(config).filter(([key]) => !["channelMode", "baseUrl", "apiKey", "localChannels", "publicChannels"].includes(key)));
    return apiPost<WorkspaceConfigPayload>("/api/v1/config/model", { config: workflowChannels === undefined ? cleanConfig : { ...cleanConfig, workflowChannels } });
}

export type CustomStorageProviders = {
    s3?: CustomS3StorageProvider;
    webdav?: CustomWebDAVStorageProvider;
};

export async function syncCustomStorageProvider(provider: CustomStorageProviders) {
    return apiPost<WorkspaceConfigPayload>("/api/v1/config/storage", {
        provider: {
            ...(provider.s3 ? { s3: toProviderPayload(provider.s3) } : {}),
            ...(provider.webdav ? { webdav: toProviderPayload(provider.webdav) } : {}),
        },
    });
}

export async function measureCustomStorageProvider(provider: CustomStorageProvider) {
    return apiPost<StorageCapacityResult>("/api/v1/storage/measure", { provider: toProviderPayload(provider) });
}

export async function fetchWorkspaceImageHistory<T>() {
    return apiGet<T>("/api/v1/data/image-history", undefined);
}

export async function syncWorkspaceImageHistory<T>(data: T) {
    return apiPost<T>("/api/v1/data/image-history", { data });
}

export async function fetchWorkspaceAssetData<T>() {
    return apiGet<T>("/api/v1/data/assets", undefined);
}

export async function syncWorkspaceAssetData<T>(data: T) {
    return apiPost<T>("/api/v1/data/assets", { data });
}

export type CreativeWorkflowRecord<T = unknown> = {
    id: string;
    ownerWorkspaceId?: string;
    scope: "private" | "public";
    name: string;
    category: string;
    description: string;
    data: T;
    createdAt: string;
    updatedAt: string;
    lastRunAt?: string;
    editable: boolean;
};

export async function fetchWorkspaceWorkflows<T>() {
    return apiGet<Array<CreativeWorkflowRecord<T>>>("/api/v1/workflows", undefined);
}

export async function saveWorkspaceWorkflow<T>(workflow: CreativeWorkflowRecord<T>) {
    return apiPost<CreativeWorkflowRecord<T>>("/api/v1/workflows", workflow);
}

export async function deleteWorkspaceWorkflow(id: string) {
    return apiDelete<boolean>(`/api/v1/workflows/${encodeURIComponent(id)}`);
}

export type WorkflowAgentDraftResponse<T = unknown> = {
    draft: T;
    warnings: string[];
    model: string;
};

export async function draftWorkspaceWorkflow<T>(
    payload: {
        prompt: string;
        scope: "private" | "public";
        model?: string;
        channelId?: string;
        references?: string[];
    },
) {
    return apiPost<WorkflowAgentDraftResponse<T>>("/api/v1/workflows/agent-draft", payload);
}
